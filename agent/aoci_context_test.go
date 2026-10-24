package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Autumn-27/artex/aoci"
)

func TestAOCIContextSpecInjectsActiveWholeIndex(t *testing.T) {
	workspace := t.TempDir()
	project := filepath.Join(workspace, "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "main.go"), []byte("package demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := aoci.Build(project); err != nil {
		t.Fatal(err)
	}
	if err := aoci.SetActiveProject(workspace, project); err != nil {
		t.Fatal(err)
	}
	got := aociContextSpec(filepath.Join(workspace, "tasks", "1"))
	if !strings.Contains(got, "AOCI-CODE Whole-Index") || !strings.Contains(got, "main.go") || !strings.Contains(got, "runtime_state=aligned") {
		t.Fatalf("missing AOCI runtime context:\n%s", got)
	}
}
