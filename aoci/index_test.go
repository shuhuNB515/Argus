package aoci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildCreatesWholeIndexAndSkipsSecretsAndDependencies(t *testing.T) {
	root := t.TempDir()
	mustWrite := func(rel, content string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("cmd/app/main.go", "package main\n\nfunc main() {}\n")
	mustWrite("schema.sql", "create table reports (id integer primary key);\n")
	mustWrite(".env", "API_KEY=must-not-be-indexed\n")
	mustWrite("node_modules/pkg/index.js", "export const ignored = true\n")

	status, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	if status.State != StateAligned || status.Files != 2 {
		t.Fatalf("status = %+v, want aligned with 2 files", status)
	}
	for _, name := range []string{ManifestName, MetaVolumeName, CodeVolumeName} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("%s not written: %v", name, err)
		}
	}
	code, err := os.ReadFile(filepath.Join(root, CodeVolumeName))
	if err != nil {
		t.Fatal(err)
	}
	got := string(code)
	if !strings.Contains(got, "cmd/app/main.go") || !strings.Contains(got, "schema.sql") {
		t.Fatalf("code volume is missing indexed files:\n%s", got)
	}
	if strings.Contains(got, "must-not-be-indexed") || strings.Contains(got, "node_modules") {
		t.Fatalf("code volume contains excluded content:\n%s", got)
	}
	if context := RuntimeContext(root); !strings.Contains(context, "#AOCI-ROOT-MANIFEST") || !strings.Contains(context, "cmd/app/main.go") {
		t.Fatalf("runtime context did not load the generated index:\n%s", context)
	}
}

func TestStatusBecomesStaleAfterSourceChanges(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.py")
	if err := os.WriteFile(path, []byte("print('one')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(root); err != nil {
		t.Fatal(err)
	}
	if got := StatusOf(root); got.State != StateAligned {
		t.Fatalf("state after build = %s, want %s", got.State, StateAligned)
	}
	time.Sleep(2 * time.Millisecond)
	if err := os.WriteFile(path, []byte("print('two')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := StatusOf(root); got.State != StateStale {
		t.Fatalf("state after edit = %s, want %s", got.State, StateStale)
	}
}

func TestActiveProjectContextResolvesFromNestedAgentDirectory(t *testing.T) {
	workspace := t.TempDir()
	project := filepath.Join(workspace, "projects", "demo")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "main.go"), []byte("package demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(project); err != nil {
		t.Fatal(err)
	}
	if err := SetActiveProject(workspace, project); err != nil {
		t.Fatal(err)
	}
	agentDir := filepath.Join(workspace, "tasks", "7", "i3")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	context := RuntimeContextFor(agentDir)
	if !strings.Contains(context, "#Project: demo") || !strings.Contains(context, "main.go") {
		t.Fatalf("nested agent directory did not resolve active project:\n%s", context)
	}
}

func TestRuntimeContextForRefusesStaleWholeIndex(t *testing.T) {
	workspace := t.TempDir()
	project := filepath.Join(workspace, "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(project, "main.go")
	if err := os.WriteFile(path, []byte("package demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(project); err != nil {
		t.Fatal(err)
	}
	if err := SetActiveProject(workspace, project); err != nil {
		t.Fatal(err)
	}
	if got := RuntimeStatusFor(filepath.Join(workspace, "tasks", "1")); got.State != StateAligned || got.Fingerprint == "" {
		t.Fatalf("runtime status = %+v, want aligned fingerprint", got)
	}
	if err := os.WriteFile(path, []byte("package changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := RuntimeStatusFor(filepath.Join(workspace, "tasks", "1")); got.State != StateStale {
		t.Fatalf("runtime status after edit = %+v, want stale", got)
	}
	if got := RuntimeContextFor(filepath.Join(workspace, "tasks", "1")); got != "" {
		t.Fatalf("stale AOCI context was injected: %q", got[:min(len(got), 80)])
	}
}

func TestExternalActiveProjectResolvesFromRuntimeWorkspace(t *testing.T) {
	data := t.TempDir()
	project := filepath.Join(t.TempDir(), "argus")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "main.go"), []byte("package argus\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(project); err != nil {
		t.Fatal(err)
	}
	if err := SetExternalActiveProject(data, project); err != nil {
		t.Fatal(err)
	}
	status := RuntimeStatusFor(filepath.Join(data, "tasks", "1"))
	if status.State != StateAligned || status.Root != filepath.Clean(project) {
		t.Fatalf("external runtime status = %+v", status)
	}
	if context := RuntimeContextFor(filepath.Join(data, "tasks", "1")); !strings.Contains(context, "#Project: argus") {
		t.Fatalf("external project context missing: %s", context)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestSelectingWorkspaceRootDoesNotMakeItsIndexStale(t *testing.T) {
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, "main.go"), []byte("package demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(workspace); err != nil {
		t.Fatal(err)
	}
	if err := SetActiveProject(workspace, workspace); err != nil {
		t.Fatal(err)
	}
	if got := StatusOf(workspace); got.State != StateAligned {
		t.Fatalf("state after selecting workspace root = %s, want %s", got.State, StateAligned)
	}
}
