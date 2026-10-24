package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBrowserEvidenceRequiresTaskScopeAndPersistsProvenance(t *testing.T) {
	s, _ := newRetestServer(t)
	s.seedChatAgents()
	c, err := s.m.pg.CreateConversation("src_agent", "browser evidence fixture", nil)
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.ensureSRCConversationTask(c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.bindSRCTargetContext(*c.TaskID, "Target: https://shu560.cn/\nTest Range: shu560.cn\nPorts: 80,443"); err != nil {
		t.Fatal(err)
	}
	screenshot := filepath.Join(t.TempDir(), "home.png")
	if err := os.WriteFile(screenshot, []byte("png fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"run_id": "run-1", "node_id": "node-1", "kind": "page_state", "url": "https://shu560.cn/",
		"timestamp": time.Date(2026, 10, 24, 1, 2, 3, 0, time.UTC), "page_state": map[string]any{"title": "home"},
		"screenshot_path": screenshot, "provenance": "chrome/readonly",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/"+task.ID+"/browser-evidence", strings.NewReader(string(body)))
	req.SetPathValue("id", task.ID)
	w := httptest.NewRecorder()
	s.browserEvidence(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register browser evidence: %d %s", w.Code, w.Body.String())
	}
	var got browserEvidenceRecord
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.URL != "https://shu560.cn/" || got.Provenance != "chrome/readonly" || got.ScreenshotPath == "" {
		t.Fatalf("unexpected evidence record: %+v", got)
	}
	if _, err := os.Stat(filepath.Join(s.m.dir, "tasks", task.ID, got.ScreenshotPath)); err != nil {
		t.Fatalf("copied screenshot missing: %v", err)
	}
	refs := loadBrowserEvidence(s.browserEvidenceDir(task.ID))
	if len(refs) != 1 || refs[0].ID != got.ID || refs[0].URL != got.URL {
		t.Fatalf("stored evidence refs: %+v", refs)
	}
	outside, _ := json.Marshal(map[string]any{"kind": "page_state", "url": "https://example.com/"})
	bad := httptest.NewRequest(http.MethodPost, "/api/tasks/"+task.ID+"/browser-evidence", strings.NewReader(string(outside)))
	bad.SetPathValue("id", task.ID)
	badW := httptest.NewRecorder()
	s.browserEvidence(badW, bad)
	if badW.Code != http.StatusForbidden {
		t.Fatalf("out-of-range browser URL was accepted: %d %s", badW.Code, badW.Body.String())
	}
}
