package db

import (
	"os"
	"path/filepath"
	"testing"
)

// The release job supplies a bundled archive and disables network access via
// invalid proxies. Exercise first boot, Unicode paths, schema, and persistence.
func TestWindowsOfflineEmbeddedDatabase(t *testing.T) {
	if os.Getenv("ARGUS_POSTGRES_CACHE") == "" {
		t.Skip("release runtime not supplied")
	}
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")
	dir := filepath.Join(t.TempDir(), "新建文件夹", "Argus data")
	d, err := OpenEmbedded(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if _, err := d.Exec("CREATE TABLE argus_release_smoke (value TEXT); INSERT INTO argus_release_smoke VALUES ('中文数据持久化')"); err != nil {
		t.Fatal(err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	d2, err := OpenEmbedded(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer d2.Close()
	var value string
	if err := d2.QueryRow("SELECT value FROM argus_release_smoke").Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "中文数据持久化" {
		t.Fatalf("persistent value = %q", value)
	}
}
