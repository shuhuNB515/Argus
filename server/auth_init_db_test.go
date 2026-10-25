package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/shuhuNB515/Argus/db"
)

// 首次运行管理员帐户抢占报告(/api/auth/init 未经身份验证的呼叫)修改实际情况 DB 到最后
// 这是一个集成测试来检查. 拒绝没有设置令牌的初始化, 令牌初始化成功, 更改密码和
// 外面(reset-password.sh) 重置后使现有令牌失效, 限制登录失败, 环境变量
// 处理初始化. DB 如果没有配置，则会像任何其他集成测试一样被跳过。.

type authTestRig struct {
	t *testing.T
	s *Server
	h http.Handler
}

func newAuthTestRig(t *testing.T) *authTestRig {
	t.Helper()
	dsn, _, err := db.DSN()
	if err != nil {
		t.Skipf("no database config (%v) — skipping", err)
	}
	d, err := db.Open(dsn)
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	clear := func() { _, _ = d.Exec(`DELETE FROM settings WHERE key = $1`, authPassKey) }
	clear()
	t.Cleanup(clear)
	t.Setenv(setupTokenEnv, "")
	t.Setenv(adminPasswordEnv, "")

	s := &Server{m: &Manager{pg: d}, jwtKey: []byte(strings.Repeat("s", 32))}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/auth/status", s.authStatus)
	mux.HandleFunc("POST /api/auth/init", s.authInit)
	mux.HandleFunc("POST /api/auth/login", s.authLogin)
	mux.HandleFunc("POST /api/auth/change-password", s.authChangePassword)
	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return &authTestRig{t: t, s: s, h: s.requireAuth(mux)}
}

func (r *authTestRig) do(method, path, token string, body map[string]any) (int, map[string]any) {
	r.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.h.ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func (r *authTestRig) expect(code int, got int, body map[string]any, wantErr string) {
	r.t.Helper()
	if got != code {
		r.t.Fatalf("状态码 = %d, 期望 %d (body=%v)", got, code, body)
	}
	if wantErr != "" {
		if msg, _ := body["error"].(string); msg != wantErr {
			r.t.Fatalf("错误文本 = %q, 期望 %q", msg, wantErr)
		}
	}
}

func TestAuthInitRequiresSetupTokenDB(t *testing.T) {
	rig := newAuthTestRig(t)
	s := rig.s
	s.bootstrapAuth()
	if h, _ := s.cachedPassHash(); h != "" {
		t.Fatal("空的 DB 里面应该没有密码")
	}
	setup := s.currentSetupToken()
	if setup == "" {
		t.Fatal("必须在启动时创建配置令牌")
	}

	code, body := rig.do(http.MethodGet, "/api/auth/status", "", nil)
	rig.expect(200, code, body, "")
	if body["initialized"] != false || body["setup_token_required"] != true {
		t.Fatalf("初始化前 status 是不同的: %v", body)
	}

	// 提示 PoC: 仅发送密码而不发送令牌的调用应被拒绝.
	code, body = rig.do(http.MethodPost, "/api/auth/init", "", map[string]any{"password": "attacker-set-password-123!"})
	rig.expect(403, code, body, authErrSetupTokenInvalid)
	if _, has := body["token"]; has {
		t.Fatal("拒绝响应中不应包含令牌")
	}
	code, body = rig.do(http.MethodPost, "/api/auth/init", "", map[string]any{"password": "attacker-set-password-123!", "setup_token": "guess"})
	rig.expect(403, code, body, authErrSetupTokenInvalid)
	if h, _ := s.reloadPassHash(); h != "" {
		t.Fatal("被拒绝的呼叫不应存储密码")
	}

	// 即使令牌正确，服务器也会检查最小长度。.
	code, body = rig.do(http.MethodPost, "/api/auth/init", "", map[string]any{"password": "short", "setup_token": setup})
	rig.expect(400, code, body, authErrPasswordTooShort)
	code, body = rig.do(http.MethodPost, "/api/auth/init", "", map[string]any{"password": "", "setup_token": setup})
	rig.expect(400, code, body, authErrPasswordEmpty)

	// 正确的标记 + 密码 → 代币发行, 受保护路径遍历.
	code, body = rig.do(http.MethodPost, "/api/auth/init", "", map[string]any{"password": "operator-pass-123", "setup_token": " " + setup + "\n"})
	rig.expect(200, code, body, "")
	tok1, _ := body["token"].(string)
	if tok1 == "" {
		t.Fatal("成功初始化后必须存在令牌")
	}
	code, body = rig.do(http.MethodGet, "/api/ping", tok1, nil)
	rig.expect(200, code, body, "")
	code, body = rig.do(http.MethodGet, "/api/ping", "", nil)
	rig.expect(401, code, body, authErrUnauthorized)
	code, body = rig.do(http.MethodGet, "/api/auth/status", "", nil)
	rig.expect(200, code, body, "")
	if body["initialized"] != true || body["setup_token_required"] != false {
		t.Fatalf("初始化后 status 是不同的: %v", body)
	}

	// 一旦设置完毕，即使您知道配置令牌，也无法再次初始化。.
	code, body = rig.do(http.MethodPost, "/api/auth/init", "", map[string]any{"password": "second-pass-123", "setup_token": setup})
	rig.expect(403, code, body, authErrPasswordAlreadySet)

	// 登录: 用户名·所有密码必须正确.
	code, body = rig.do(http.MethodPost, "/api/auth/login", "", map[string]any{"username": "Argus", "password": "wrong-pass-123"})
	rig.expect(401, code, body, authErrBadCredential)
	code, body = rig.do(http.MethodPost, "/api/auth/login", "", map[string]any{"username": "admin", "password": "operator-pass-123"})
	rig.expect(401, code, body, authErrBadCredential)
	code, body = rig.do(http.MethodPost, "/api/auth/login", "", map[string]any{"username": "Argus", "password": "operator-pass-123"})
	rig.expect(200, code, body, "")
	tok2, _ := body["token"].(string)
	code, body = rig.do(http.MethodGet, "/api/ping", tok2, nil)
	rig.expect(200, code, body, "")

	// 更改密码: 之前的所有代币都会消失, 只有响应中的新令牌才有效.
	code, body = rig.do(http.MethodPost, "/api/auth/change-password", tok1, map[string]any{"old_password": "nope-nope-123", "new_password": "changed-pass-123"})
	rig.expect(401, code, body, authErrCurrentPasswordWrong)
	code, body = rig.do(http.MethodPost, "/api/auth/change-password", tok1, map[string]any{"old_password": "operator-pass-123", "new_password": "short"})
	rig.expect(400, code, body, authErrPasswordTooShort)
	code, body = rig.do(http.MethodPost, "/api/auth/change-password", tok1, map[string]any{"old_password": "operator-pass-123", "new_password": "changed-pass-123"})
	rig.expect(200, code, body, "")
	tok3, _ := body["token"].(string)
	if tok3 == "" {
		t.Fatal("密码更改响应中应该有一个新令牌")
	}
	for _, old := range []string{tok1, tok2} {
		code, body = rig.do(http.MethodGet, "/api/ping", old, nil)
		rig.expect(401, code, body, authErrTokenInvalid)
	}
	code, body = rig.do(http.MethodGet, "/api/ping", tok3, nil)
	rig.expect(200, code, body, "")
	code, body = rig.do(http.MethodPost, "/api/auth/change-password", tok1, map[string]any{"old_password": "changed-pass-123", "new_password": "another-pass-123"})
	rig.expect(401, code, body, authErrUnauthorized)

	// reset-password.sh 同样，即使哈希值在进程外部发生更改，现有令牌也会在不重新启动的情况下消失。
	// 您将使用新密码登录. 缓存过期窗口(passHashCacheTTL) 在里面，前一个令牌是
	// 还在路过, 窗口过去后(这里 0 立即减少到) 应该被拒绝.
	resetHash, err := bcrypt.GenerateFromPassword([]byte("reset-pass-123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.m.pg.SetSetting(authPassKey, string(resetHash)); err != nil {
		t.Fatal(err)
	}
	code, body = rig.do(http.MethodGet, "/api/ping", tok3, nil)
	rig.expect(200, code, body, "")
	prevTTL := passHashCacheTTL
	passHashCacheTTL = 0
	t.Cleanup(func() { passHashCacheTTL = prevTTL })
	code, body = rig.do(http.MethodGet, "/api/ping", tok3, nil)
	rig.expect(401, code, body, authErrTokenInvalid)
	code, body = rig.do(http.MethodPost, "/api/auth/login", "", map[string]any{"username": "Argus", "password": "changed-pass-123"})
	rig.expect(401, code, body, authErrBadCredential)
	code, body = rig.do(http.MethodPost, "/api/auth/login", "", map[string]any{"username": "Argus", "password": "reset-pass-123"})
	rig.expect(200, code, body, "")
	tok4, _ := body["token"].(string)
	code, body = rig.do(http.MethodGet, "/api/ping", tok4, nil)
	rig.expect(200, code, body, "")
}

func TestAuthLoginThrottledDB(t *testing.T) {
	rig := newAuthTestRig(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("operator-pass-123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := rig.s.m.pg.SetSetting(authPassKey, string(hash)); err != nil {
		t.Fatal(err)
	}
	rig.s.bootstrapAuth()
	for i := 0; i < authFreeFailures; i++ {
		code, body := rig.do(http.MethodPost, "/api/auth/login", "", map[string]any{"username": "Argus", "password": "brute-force-attempt"})
		rig.expect(401, code, body, authErrBadCredential)
	}
	code, body := rig.do(http.MethodPost, "/api/auth/login", "", map[string]any{"username": "Argus", "password": "operator-pass-123"})
	rig.expect(429, code, body, authErrTooManyAttempts)
	if _, has := body["token"]; has {
		t.Fatal("冷却期间，即使密码正确，也不应给予令牌。")
	}
	rig.s.limiter().reset(clientKey(httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)))
	code, body = rig.do(http.MethodPost, "/api/auth/login", "", map[string]any{"username": "Argus", "password": "operator-pass-123"})
	rig.expect(200, code, body, "")
}

func TestAuthBootstrapAdminPasswordEnvDB(t *testing.T) {
	rig := newAuthTestRig(t)
	t.Setenv(adminPasswordEnv, "env-admin-pass-123")
	rig.s.bootstrapAuth()
	code, body := rig.do(http.MethodGet, "/api/auth/status", "", nil)
	rig.expect(200, code, body, "")
	if body["initialized"] != true {
		t.Fatalf("必须初始化为环境变量: %v", body)
	}
	code, body = rig.do(http.MethodPost, "/api/auth/init", "", map[string]any{"password": "attacker-set-password-123!", "setup_token": rig.s.currentSetupToken()})
	rig.expect(403, code, body, authErrPasswordAlreadySet)
	code, body = rig.do(http.MethodPost, "/api/auth/login", "", map[string]any{"username": "Argus", "password": "env-admin-pass-123"})
	rig.expect(200, code, body, "")
	tok, _ := body["token"].(string)
	code, body = rig.do(http.MethodGet, "/api/ping", tok, nil)
	rig.expect(200, code, body, "")

	// 太短的环境变量值将被忽略并保留为配置令牌。.
	_, _ = rig.s.m.pg.Exec(`DELETE FROM settings WHERE key = $1`, authPassKey)
	t.Setenv(adminPasswordEnv, "short")
	s2 := &Server{m: rig.s.m, jwtKey: rig.s.jwtKey}
	s2.bootstrapAuth()
	if h, _ := s2.cachedPassHash(); h != "" {
		t.Fatal("8应忽略少于 1 个字符的环境变量密码")
	}
	if s2.currentSetupToken() == "" {
		t.Fatal("被覆盖后，必须发出设置令牌")
	}
}
