package report

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Autumn-27/artex/db"
)

// EvidenceRef is a sanitized pointer to evidence that can be cited by a report.
type EvidenceRef struct {
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	URL        string    `json:"url"`
	RunID      string    `json:"run_id,omitempty"`
	NodeID     string    `json:"node_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Provenance string    `json:"provenance"`
}

// BundleInput is the canonical current-task input for SRC reports. Confirmed
// findings and candidates are deliberately separate so pending observations can
// never appear in the confirmed section by accident.
type BundleInput struct {
	Title           string
	Goal            string
	Target          string
	TestRange       []string
	Ports           []int
	GeneratedAt     time.Time
	AssetCounts     map[string]int
	Confirmed       []*db.DBFinding
	Candidates      []*db.DBFinding
	Observations    []string
	BrowserEvidence []EvidenceRef
}

type Bundle struct {
	Markdown []byte
	HTML     []byte
	PDF      []byte
	Metadata []byte
}

func RenderBundle(in BundleInput) Bundle {
	if in.GeneratedAt.IsZero() {
		in.GeneratedAt = time.Now().UTC()
	}
	md := renderMarkdown(in)
	metadata, _ := json.MarshalIndent(map[string]any{
		"version":               1,
		"generated_at":          in.GeneratedAt,
		"target":                in.Target,
		"test_range":            in.TestRange,
		"ports":                 in.Ports,
		"confirmed_findings":    len(in.Confirmed),
		"candidate_findings":    len(in.Candidates),
		"browser_evidence_refs": in.BrowserEvidence,
	}, "", "  ")
	metadata = append(metadata, '\n')
	return Bundle{Markdown: []byte(md), HTML: []byte(markdownHTML(md)), PDF: minimalPDF(md), Metadata: metadata}
}

func WriteBundle(dir string, in BundleInput) (Bundle, error) {
	b := RenderBundle(in)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Bundle{}, err
	}
	for name, data := range map[string][]byte{
		"src-report.md":            b.Markdown,
		"src-report.html":          b.HTML,
		"src-report.pdf":           b.PDF,
		"src-report.metadata.json": b.Metadata,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			return Bundle{}, err
		}
	}
	return b, nil
}

func renderMarkdown(in BundleInput) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# SRC 报告：%s\n\n", nz(in.Title, "未命名任务"))
	fmt.Fprintf(&b, "- **任务目标**：%s\n- **Target**：`%s`\n- **Test Range**：`%s`\n- **Ports**：`%s`\n- **生成时间**：%s\n\n", nz(in.Goal, "未指定"), in.Target, strings.Join(in.TestRange, ", "), joinPorts(in.Ports), in.GeneratedAt.UTC().Format(time.RFC3339))
	b.WriteString("## Quality Gate\n\n")
	fmt.Fprintf(&b, "- confirmed findings: **%d**\n- candidate / pending findings: **%d**\n- browser evidence refs: **%d**\n", len(in.Confirmed), len(in.Candidates), len(in.BrowserEvidence))
	if len(in.Confirmed) == 0 {
		b.WriteString("- 本轮没有通过验证、影响证据和来源校验的 confirmed vulnerability。\n")
	}
	b.WriteString("\n## Confirmed Findings\n\n")
	if len(in.Confirmed) == 0 {
		b.WriteString("_0 confirmed findings._\n\n")
	} else {
		for i, f := range in.Confirmed {
			fmt.Fprintf(&b, "### %d. [%s] %s\n\n%s\n\n", i+1, strings.ToUpper(nz(f.Severity, "info")), nz(f.Name, nz(f.VulnClass, "未分类")), nz(f.Summary, ""))
			if f.Evidence != "" {
				fmt.Fprintf(&b, "**Evidence:**\n\n```\n%s\n```\n\n", f.Evidence)
			}
		}
	}
	b.WriteString("## Candidate Findings / Observations\n\n")
	if len(in.Candidates) == 0 && len(in.Observations) == 0 {
		b.WriteString("_无未决候选项。_\n\n")
	}
	for _, f := range in.Candidates {
		fmt.Fprintf(&b, "- [%s] %s — status=%s；%s\n", strings.ToUpper(nz(f.Severity, "info")), nz(f.Name, nz(f.VulnClass, "未分类")), f.Status, nz(f.Summary, ""))
	}
	for _, observation := range in.Observations {
		if strings.TrimSpace(observation) != "" {
			fmt.Fprintf(&b, "- %s\n", strings.TrimSpace(observation))
		}
	}
	b.WriteString("\n## Browser Evidence\n\n")
	if len(in.BrowserEvidence) == 0 {
		b.WriteString("_本轮没有注册 browser evidence。_\n")
	} else {
		refs := append([]EvidenceRef(nil), in.BrowserEvidence...)
		sort.SliceStable(refs, func(i, j int) bool { return refs[i].Timestamp.Before(refs[j].Timestamp) })
		for _, ref := range refs {
			fmt.Fprintf(&b, "- `%s` kind=%s url=%s provenance=%s timestamp=%s\n", ref.ID, ref.Kind, ref.URL, ref.Provenance, ref.Timestamp.UTC().Format(time.RFC3339))
		}
	}
	return b.String()
}

func markdownHTML(md string) string {
	// The report is intentionally rendered as escaped preformatted text. This
	// preserves the exact Markdown artifact while making HTML injection impossible.
	return "<!doctype html><html><head><meta charset=\"utf-8\"><title>SRC Report</title></head><body><pre>" + html.EscapeString(md) + "</pre></body></html>\n"
}

func minimalPDF(text string) []byte {
	lines := strings.Split(strings.ReplaceAll(text, "\r", ""), "\n")
	var stream strings.Builder
	stream.WriteString("BT /F1 9 Tf 36 806 Td 12 TL\n")
	for _, line := range lines {
		line = strings.ReplaceAll(line, "\\", "\\\\")
		line = strings.ReplaceAll(line, "(", "\\(")
		line = strings.ReplaceAll(line, ")", "\\)")
		stream.WriteString("(" + truncatePDFLine(line, 180) + ") Tj T*\n")
	}
	stream.WriteString("ET")
	content := stream.String()
	objects := []string{
		"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n",
		"2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n",
		"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 828] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>\nendobj\n",
		"4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n",
		fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(content), content),
	}
	var b strings.Builder
	b.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for _, obj := range objects {
		offsets = append(offsets, b.Len())
		b.WriteString(obj)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return []byte(b.String())
}

func truncatePDFLine(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func joinPorts(ports []int) string {
	parts := make([]string, 0, len(ports))
	for _, port := range ports {
		parts = append(parts, fmt.Sprint(port))
	}
	return strings.Join(parts, ", ")
}
