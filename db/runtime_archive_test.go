package db

import "testing"

func TestBundledRuntimePlatformNames(t *testing.T) {
	for _, tc := range []struct{ os, arch, want string }{
		{"linux", "arm64", "embedded-postgres-binaries-linux-arm64v8-18.3.0.txz"},
		{"darwin", "arm64", "embedded-postgres-binaries-darwin-arm64v8-18.3.0.txz"},
		{"linux", "amd64", "embedded-postgres-binaries-linux-amd64-18.3.0.txz"},
		{"windows", "amd64", "embedded-postgres-binaries-windows-amd64-18.3.0.txz"},
	} {
		if got := bundledArchiveName(tc.os, tc.arch); got != tc.want {
			t.Errorf("%s/%s: %s", tc.os, tc.arch, got)
		}
	}
}
