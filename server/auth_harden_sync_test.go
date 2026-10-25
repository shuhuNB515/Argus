package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

// 上游 a951e4a 移植(积压: 阅读认证 fail-closed + 首次设置雾化 + 密码上限)的
// 回归防御. GetSetting 银 "没有钥匙"和 "读取失败"返回值 error 只有一件事不同(两者
// value=""). 密码处理程序 error 到 "未设置"数据库故障期间视为初始化
// 入口打开. 下面的两个测试关闭连接池并导致读取失败。 fail-closed(503)到
// 确认.
func TestAuthStatusFailsClosedWhenDataSourceUnavailable(t *testing.T) {
	m, err := NewManager(t.TempDir(), "")
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	defer m.Close()
	if err := m.pg.Close(); err != nil {
		t.Fatal(err)
	}

	s := &Server{m: m}
	w := httptest.NewRecorder()
	s.authStatus(w, httptest.NewRequest("GET", "/api/auth/status", nil))

	if w.Code != 503 {
		t.Fatalf("status=%d, 503 我正期待着 (如果读取失败被视为未初始化，用户可以 /setup 前往并覆盖您的密码。); body=%s", w.Code, w.Body.String())
	}
	var payload struct {
		Initialized *bool `json:"initialized"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err == nil && payload.Initialized != nil {
		t.Fatalf("当读取失败时 initialized 你一定不能回答: %v", *payload.Initialized)
	}
}

func TestAuthInitFailsClosedWhenDataSourceUnavailable(t *testing.T) {
	m, err := NewManager(t.TempDir(), "")
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	defer m.Close()
	if err := m.pg.Close(); err != nil {
		t.Fatal(err)
	}

	s := &Server{m: m}
	w := httptest.NewRecorder()
	body := strings.NewReader(`{"password":"correct horse battery","setup_token":"x"}`)
	s.authInit(w, httptest.NewRequest("POST", "/api/auth/init", body))

	if w.Code != 503 {
		t.Fatalf("status=%d, 503 我正期待着 (如果传递读取失败，未经身份验证的请求将覆盖现有密码。); body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "token") {
		t.Fatalf("当读取失败时 token 不应发出: %s", w.Body.String())
	}
}

// 密码字节上限(bcrypt 72字节)检查服务器是否强制执行. 73字节是 400 到
// 应该被拒绝(以前 bcrypt 在哈希步骤失败 500 这件事发生了).
func TestAuthInitRejectsTooLongPasswordDB(t *testing.T) {
	rig := newAuthTestRig(t)
	tok := rig.s.currentSetupToken()

	code, body := rig.do("POST", "/api/auth/init", "", map[string]any{
		"password":    strings.Repeat("a", 73),
		"setup_token": tok,
	})
	rig.expect(400, code, body, authErrPasswordTooLong)
}

// 启动时的令牌验证和首次执行处理也会导致读取失败。 "无密码"不要折叠它. 折叠时
// 缓存空哈希值(passHashLoaded=true), signingKey("") 生吃 jwt.key 下降到.
func TestVerifyTokenFailsClosedWhenDataSourceUnavailable(t *testing.T) {
	m, err := NewManager(t.TempDir(), "")
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	defer m.Close()
	if err := m.pg.Close(); err != nil {
		t.Fatal(err)
	}

	s := &Server{m: m, jwtKey: []byte(strings.Repeat("s", 32))}
	tok, err := signJWT(s.jwtKey)
	if err != nil {
		t.Fatal(err)
	}
	if s.verifyToken(tok) {
		t.Fatal("当密码状态不可读时不应传递令牌")
	}
}

func TestBootstrapAuthSkipsFirstRunWhenDataSourceUnavailable(t *testing.T) {
	m, err := NewManager(t.TempDir(), "")
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	defer m.Close()
	if err := m.pg.Close(); err != nil {
		t.Fatal(err)
	}
	t.Setenv(adminPasswordEnv, "env-admin-pass-123")

	s := &Server{m: m}
	s.bootstrapAuth()
	s.authMu.Lock()
	loaded := s.passHashLoaded
	s.authMu.Unlock()
	if loaded {
		t.Fatal("读取失败 \"无密码\"它不应该被缓存")
	}
}

// ARGUS_ADMIN_PASSWORD 初始化也是首先 1只能使用生鱼: 使用不同值启动的第二个实例
// 不要覆盖现有密码.
func TestBootstrapAdminPasswordEnvDoesNotOverwriteDB(t *testing.T) {
	rig := newAuthTestRig(t)
	t.Setenv(adminPasswordEnv, "first-admin-pass-123")
	rig.s.bootstrapAuth()
	first, err := rig.s.reloadPassHash()
	if err != nil || first == "" {
		t.Fatalf("第一个实例必须重置密码 (err=%v)", err)
	}

	// 二审: 读取的起点不为空，因此您应该立即返回，但是, 阅读和
	// 直接检查写入路径以模拟写入之间的争用.
	inserted, err := rig.s.m.pg.InsertSettingIfAbsent(authPassKey, "second-hash")
	if err != nil {
		t.Fatal(err)
	}
	if inserted {
		t.Fatal("覆盖现有密码 insert 不应该成功")
	}
	t.Setenv(adminPasswordEnv, "second-admin-pass-456")
	s2 := &Server{m: rig.s.m, jwtKey: rig.s.jwtKey}
	s2.bootstrapAuth()
	if got, _ := s2.reloadPassHash(); got != first {
		t.Fatal("第二个实例覆盖了现有密码")
	}
}

func TestBootstrapAdminPasswordEnvRejectsOverBcryptLimitDB(t *testing.T) {
	rig := newAuthTestRig(t)
	t.Setenv(adminPasswordEnv, strings.Repeat("a", 73))
	rig.s.bootstrapAuth()
	if h, _ := rig.s.reloadPassHash(); h != "" {
		t.Fatal("72应忽略长于字节的环境变量密码")
	}
	if rig.s.currentSetupToken() == "" {
		t.Fatal("被覆盖后，必须发出设置令牌")
	}
}
