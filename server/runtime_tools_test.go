package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLocalExecutableFindsNmapOnPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nmap")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	got, ok := detectLocalExecutable("nmap")
	if !ok {
		t.Fatal("expected nmap to be detected")
	}
	if got != path {
		t.Fatalf("path = %q, want %q", got, path)
	}
}

func TestDetectLocalExecutableRejectsMissingTool(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if path, ok := detectLocalExecutable("definitely-not-installed-argus-test-tool"); ok || path != "" {
		t.Fatalf("got path=%q ok=%v, want missing", path, ok)
	}
}
