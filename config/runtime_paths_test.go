package config

import (
	"path/filepath"
	"testing"
)

func TestRuntimeHomeOverride(t *testing.T) {
	home := filepath.Join(t.TempDir(), "Argus 用户数据")
	t.Setenv("ARGUS_HOME", home)
	t.Setenv("ARGUS_CONFIG", "")
	if BaseDir() != home || Path() != filepath.Join(home, "config.json") {
		t.Fatalf("runtime paths: base=%q config=%q", BaseDir(), Path())
	}
}

func TestOnlyGoBuildDirectoriesAreThrowaway(t *testing.T) {
	for _, path := range []string{"/tmp/argus-release/Argus.app/Contents/MacOS", "/tmp/argus-0.1", "/home/user/Argus"} {
		if isGoRunDir(path) {
			t.Errorf("release directory misidentified: %s", path)
		}
	}
	for _, path := range []string{"/tmp/go-build12345/b001/exe", "/home/user/.cache/go-build/ab"} {
		if !isGoRunDir(path) {
			t.Errorf("go build directory not identified: %s", path)
		}
	}
}
