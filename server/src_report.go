package server

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shuhuNB515/Argus/db"
	"github.com/shuhuNB515/Argus/report"
)

func (s *Server) renderSRCReport(t *Task) (report.Bundle, error) {
	if t == nil {
		return report.Bundle{}, os.ErrInvalid
	}
	taskID, err := strconv.ParseInt(t.ID, 10, 64)
	if err != nil {
		return report.Bundle{}, err
	}
	ctx, err := s.m.pg.TaskTargetContext(taskID)
	if err != nil {
		return report.Bundle{}, err
	}
	confirmed, _, err := s.m.pg.ListFindingsPage(db.FindingFilter{TaskID: t.ID, Status: db.FindingConfirmed}, 1, 1000)
	if err != nil {
		return report.Bundle{}, err
	}
	allFindings, _, err := s.m.pg.ListFindingsPage(db.FindingFilter{TaskID: t.ID}, 1, 1000)
	if err != nil {
		return report.Bundle{}, err
	}
	candidates := make([]*db.DBFinding, 0, len(allFindings))
	for _, finding := range allFindings {
		if finding.Status != db.FindingConfirmed && finding.Status != db.FindingResolved && finding.Status != db.FindingFixed {
			candidates = append(candidates, finding)
		}
	}
	observations := []string{}
	if draft, readErr := os.ReadFile(filepath.Join(s.m.dir, "tasks", t.ID, "src-draft.md")); readErr == nil {
		text := strings.TrimSpace(string(draft))
		text = strings.TrimSpace(strings.TrimPrefix(text, "# SRC Draft"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "> 会话研究草稿；候选问题需经证据和影响复核，不代表已确认漏洞。"))
		if text != "" {
			observations = append(observations, text)
		}
	}
	input := report.BundleInput{Title: t.Description, Goal: t.Goal, GeneratedAt: time.Now().UTC(), Confirmed: confirmed, Candidates: candidates, Observations: observations, BrowserEvidence: loadBrowserEvidence(s.browserEvidenceDir(t.ID))}
	if ctx != nil {
		input.Target, input.TestRange, input.Ports = ctx.Target, append([]string(nil), ctx.TestRange...), append([]int(nil), ctx.Ports...)
	}
	input.AssetCounts = s.taskAssetCounts(taskID)
	return report.WriteBundle(filepath.Join(s.m.dir, "tasks", t.ID, "reports"), input)
}

func (s *Server) taskAssetCounts(taskID int64) map[string]int {
	counts := map[string]int{}
	for _, kind := range []string{"root_domain", "ip", "subdomain", "app", "service", "endpoint"} {
		assets, err := s.m.Assets().QueryByTask(taskID, kind, 100000, 0)
		if err == nil && len(assets) > 0 {
			counts[kind] = len(assets)
		}
	}
	return counts
}
