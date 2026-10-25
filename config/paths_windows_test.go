package config

import (
	"path/filepath"
	"testing"
)

func TestWindowsUserConfigPath(t *testing.T) {
	local := filepath.Join(t.TempDir(), "用户数据")
	t.Setenv("LOCALAPPDATA", local)
	t.Setenv("ARGUS_CONFIG", "")
	want := filepath.Join(local, "Argus")
	if got := BaseDir(); got != want {
		t.Fatalf("BaseDir = %q, want %q", got, want)
	}
	if got := Path(); got != filepath.Join(want, "config.json") {
		t.Fatalf("Path = %q", got)
	}
}
