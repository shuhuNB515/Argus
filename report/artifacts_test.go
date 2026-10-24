package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRenderBundleSeparatesConfirmedCandidatesAndEscapesHTML(t *testing.T) {
	b := RenderBundle(BundleInput{
		Title:        "SRC fixture",
		Target:       "https://shu560.cn/",
		TestRange:    []string{"shu560.cn"},
		Ports:        []int{80, 443},
		GeneratedAt:  time.Unix(0, 0).UTC(),
		Observations: []string{"<script>alert(1)</script>"},
	})
	md := string(b.Markdown)
	if !strings.Contains(md, "0 confirmed findings") || !strings.Contains(md, "<script>alert(1)</script>") {
		t.Fatalf("unexpected markdown: %s", md)
	}
	html := string(b.HTML)
	if strings.Contains(html, "<script>alert(1)</script>") || !strings.Contains(html, "&lt;script&gt;") {
		t.Fatalf("HTML was not escaped: %s", html)
	}
	if !strings.HasPrefix(string(b.PDF), "%PDF-1.4") {
		t.Fatalf("invalid PDF header: %q", b.PDF[:minBytes(len(b.PDF), 12)])
	}
}

func TestWriteBundleCreatesAllSRCArtifacts(t *testing.T) {
	dir := t.TempDir()
	if _, err := WriteBundle(filepath.Join(dir, "reports"), BundleInput{Title: "fixture"}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"src-report.md", "src-report.html", "src-report.pdf", "src-report.metadata.json"} {
		if info, err := os.Stat(filepath.Join(dir, "reports", name)); err != nil || info.Size() == 0 {
			t.Fatalf("artifact %s missing or empty: %v", name, err)
		}
	}
}

func minBytes(a, b int) int {
	if a < b {
		return a
	}
	return b
}
