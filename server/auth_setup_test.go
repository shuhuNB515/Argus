package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// 抢占首次运行管理员帐户(/api/auth/init 未经身份验证的呼叫) 更正报告 DB 没有
// 这是一个可以保护的单元测试. 设置令牌比较, 令牌签名与密码哈希相关, CORS 允许
// 列表, 限制失败的尝试, 检查每个环回绑定决策。. DB 整个流程经过
// auth_init_db_test.go 我负责.

func TestSigningKeyBoundToPasswordHash(t *testing.T) {
	s := &Server{jwtKey: []byte(strings.Repeat("k", 32))}
	if !bytes.Equal(s.signingKey(""), s.jwtKey) {
		t.Fatal("如果没有密码，则必须按原样使用文件密钥。(DB 没有测试兼容性)")
	}
	tok, err := signJWT(s.signingKey("hash-A"))
	if err != nil {
		t.Fatal(err)
	}
	if !verifyJWT(tok, s.signingKey("hash-A")) {
		t.Fatal("必须使用从同一哈希派生的密钥对其进行验证。")
	}
	if verifyJWT(tok, s.signingKey("hash-B")) {
		t.Fatal("密码(散列)如果发生变化，之前的令牌应该失效")
	}
	if verifyJWT(tok, s.signingKey("")) {
		t.Fatal("使用哈希派生密钥签名的令牌不得使用原始密钥进行验证")
	}
}

func TestVerifyTokenInvalidatedByPasswordChange(t *testing.T) {
	s := &Server{jwtKey: []byte(strings.Repeat("v", 32))}
	s.setPassHash("hash-A")
	tok, err := signJWT(s.signingKey("hash-A"))
	if err != nil {
		t.Fatal(err)
	}
	if !s.verifyToken(tok) {
		t.Fatal("使用当前哈希签名的令牌必须通过")
	}
	s.setPassHash("hash-B")
	if s.verifyToken(tok) {
		t.Fatal("即使更改密码后，旧令牌也不应通过")
	}
	if s.verifyToken("") {
		t.Fatal("空令牌不得通过")
	}
}

func TestSetupTokenFromEnvAndRandom(t *testing.T) {
	t.Setenv(setupTokenEnv, "fixed-setup-token")
	s := &Server{}
	if !s.checkSetupToken("  fixed-setup-token \n") {
		t.Fatal("环境变量标记必须匹配，忽略前导和尾随空格")
	}
	if s.checkSetupToken("fixed-setup-tokeN") || s.checkSetupToken("") {
		t.Fatal("应拒绝其他值或空值")
	}

	t.Setenv(setupTokenEnv, "")
	s = &Server{}
	tok := s.currentSetupToken()
	if len(tok) != setupTokenLen {
		t.Fatalf("随机设置令牌长度 = %d, 期望 %d", len(tok), setupTokenLen)
	}
	if s.currentSetupToken() != tok {
		t.Fatal("一旦创建了令牌，它就必须保持不变，直到被删除。")
	}
	if !s.checkSetupToken(tok) || s.checkSetupToken(tok[:len(tok)-1]) {
		t.Fatal("仅当随机令牌完全相同时才应通过")
	}
	s.clearSetupToken()
	if s.currentSetupToken() == tok {
		t.Fatal("初始化完成并擦除后，应该会出现一个新的令牌。")
	}
}

func TestAuthLimiterCooldown(t *testing.T) {
	l := newAuthLimiter()
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return now }
	const key = "203.0.113.9"

	for i := 0; i < authFreeFailures; i++ {
		if !l.allow(key) {
			t.Fatalf("%d必须在第二次失败之前允许", i+1)
		}
		l.fail(key)
	}
	if l.allow(key) {
		t.Fatalf("%d发生故障后，必须停机冷却一段时间。", authFreeFailures)
	}
	now = now.Add(authCooldownBase + time.Second)
	if !l.allow(key) {
		t.Fatal("第一次冷却期后应再次允许")
	}
	l.fail(key)
	now = now.Add(authCooldownBase + time.Second)
	if l.allow(key) {
		t.Fatal("每次额外失败的冷却时间必须加倍")
	}
	now = now.Add(authCooldownBase)
	if !l.allow(key) {
		t.Fatal("应允许双倍冷却时间")
	}
	l.reset(key)
	l.fail(key)
	if !l.allow(key) {
		t.Fatal("初始化成功后，一次失败也是可以接受的。")
	}
	for i := 0; i < 20; i++ {
		l.fail(key)
	}
	if got := l.clients[key].until.Sub(now); got > authCooldownMax {
		t.Fatalf("超过冷却时间上限: %v > %v", got, authCooldownMax)
	}
	now = now.Add(authFailureWindow + time.Second)
	if !l.allow(key) {
		t.Fatal("在故障记录窗口之后，应清除并允许历史记录")
	}
	if l.allow("198.51.100.1") != true {
		t.Fatal("其他客户端不应受到影响")
	}
}

func TestCORSAllowlist(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := corsWith(map[string]bool{"http://localhost:5173": true}, ok)

	call := func(method, origin string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/api/auth/init", nil)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}

	if rec := call(http.MethodPost, ""); rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("Origin 无标头(相同来源) 在请求中 CORS 不应添加标题")
	}
	rec := call(http.MethodPost, "https://evil.example")
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("不属于白名单的来源 Access-Control-Allow-Origin 你一定不能给")
	}
	if rec.Header().Get("Vary") != "Origin" {
		t.Fatal("由于响应因来源而异， Vary: Origin 这必须是")
	}
	if rec := call(http.MethodOptions, "https://evil.example"); rec.Code != 204 || rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("来自未经授权来源的提前请求: code=%d ACAO=%q", rec.Code, rec.Header().Get("Access-Control-Allow-Origin"))
	}
	rec = call(http.MethodPost, "http://LOCALHOST:5173")
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://LOCALHOST:5173" {
		t.Fatalf("允许的来源必须按原样恢复，忽略大小写。: %q", got)
	}
	if !strings.Contains(rec.Header().Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Fatal("接受的来源包括： Authorization 您必须允许使用标头")
	}

	wild := corsWith(map[string]bool{"*": true}, ok)
	r := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	r.Header.Set("Origin", "https://anything.example")
	w := httptest.NewRecorder()
	wild.ServeHTTP(w, r)
	if w.Header().Get("Access-Control-Allow-Origin") != "https://anything.example" {
		t.Fatal("\"*\" 允许所有指定来源(以前的行为)必须做")
	}
}

func TestCORSAllowedOriginsEnv(t *testing.T) {
	t.Setenv(corsOriginsEnv, "")
	got := corsAllowedOrigins()
	for _, o := range defaultCORSOrigins {
		if !got[o] {
			t.Fatalf("在默认白名单中 %s 没有", o)
		}
	}
	t.Setenv(corsOriginsEnv, " https://UI.example.com , http://10.0.0.2:5173,, ")
	got = corsAllowedOrigins()
	if !got["https://ui.example.com"] || !got["http://10.0.0.2:5173"] || len(got) != 2 {
		t.Fatalf("环境变量列表解析结果不同: %v", got)
	}
	if got["http://localhost:5173"] {
		t.Fatal("如果您提供环境变量，则不应再包含默认值")
	}
}

func TestIsLoopbackAddr(t *testing.T) {
	cases := map[string]bool{
		":8787":           false,
		"0.0.0.0:8787":    false,
		"[::]:8787":       false,
		"10.0.0.5:8787":   false,
		"127.0.0.1:8787":  true,
		"127.0.0.1":       true,
		"localhost:8787":  true,
		"[::1]:8787":      true,
		"192.168.0.10:80": false,
	}
	for addr, want := range cases {
		if got := isLoopbackAddr(addr); got != want {
			t.Errorf("isLoopbackAddr(%q) = %v, 期望 %v", addr, got, want)
		}
	}
}

func TestClientKeyIgnoresForwardedFor(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	r.RemoteAddr = "198.51.100.7:40000"
	r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	if got := clientKey(r); got != "198.51.100.7" {
		t.Fatalf("限制密钥可以伪造 X-Forwarded-For 不 TCP 必须是相对地址: %q", got)
	}
	if id := clientLogID(r); !strings.Contains(id, "198.51.100.7:40000") || !strings.Contains(id, "1.2.3.4") {
		t.Fatalf("日志标识符必须具有两个值: %q", id)
	}
}
