package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAOCIIndexAndStatusUseAWorkspaceDirectory(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "projects", "demo")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "main.go"), []byte("package demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := &Server{m: &Manager{dir: root}}

	body := bytes.NewBufferString(`{"path":"projects/demo","reload_mcp":false}`)
	buildRec := httptest.NewRecorder()
	s.aociBuild(buildRec, httptest.NewRequest(http.MethodPost, "/api/aoci/index", body))
	if buildRec.Code != http.StatusOK {
		t.Fatalf("build status = %d, body=%s", buildRec.Code, buildRec.Body.String())
	}
	var built struct {
		State string `json:"state"`
		Files int    `json:"files"`
		Root  string `json:"root"`
	}
	if err := json.Unmarshal(buildRec.Body.Bytes(), &built); err != nil {
		t.Fatal(err)
	}
	if built.State != "aligned" || built.Files != 1 || built.Root != "projects/demo" {
		t.Fatalf("unexpected build response: %+v", built)
	}

	statusRec := httptest.NewRecorder()
	s.aociStatus(statusRec, httptest.NewRequest(http.MethodGet, "/api/aoci/status?path=projects/demo", nil))
	if statusRec.Code != http.StatusOK || !bytes.Contains(statusRec.Body.Bytes(), []byte(`"state":"aligned"`)) {
		t.Fatalf("status = %d, body=%s", statusRec.Code, statusRec.Body.String())
	}
}

func TestAOCIRejectsFilesAndEscapingPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := &Server{m: &Manager{dir: root}}
	for _, path := range []string{"file.txt", "../../outside"} {
		rec := httptest.NewRecorder()
		s.aociStatus(rec, httptest.NewRequest(http.MethodGet, "/api/aoci/status?path="+path, nil))
		if rec.Code < 400 {
			t.Fatalf("path %q returned %d, want an error", path, rec.Code)
		}
	}
}
