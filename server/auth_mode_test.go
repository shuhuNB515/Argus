package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthModeAutoDisablesOnlyForLoopback(t *testing.T) {
	t.Setenv(authModeEnv, "auto")
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	local := &Server{bindAddr: "127.0.0.1:8787"}
	rec := httptest.NewRecorder()
	local.requireAuth(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/tasks", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("loopback auto mode status=%d, want %d", rec.Code, http.StatusNoContent)
	}

	remote := &Server{bindAddr: "0.0.0.0:8787"}
	rec = httptest.NewRecorder()
	remote.requireAuth(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/tasks", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("remote auto mode status=%d, want 401", rec.Code)
	}
}

func TestAuthModeExplicitRequiredAndDisabled(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	t.Setenv(authModeEnv, "required")
	rec := httptest.NewRecorder()
	(&Server{bindAddr: "127.0.0.1:8787"}).requireAuth(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/tasks", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("required mode status=%d, want 401", rec.Code)
	}

	t.Setenv(authModeEnv, "disabled")
	rec = httptest.NewRecorder()
	(&Server{bindAddr: "0.0.0.0:8787"}).requireAuth(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/tasks", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("disabled mode status=%d, want %d", rec.Code, http.StatusNoContent)
	}

}
