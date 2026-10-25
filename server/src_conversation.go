package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shuhuNB515/Argus/db"
	"github.com/shuhuNB515/Argus/srcagentpolicy"
)

func (s *Server) seedSRCTaskBindings() {
	const flag = "src_task_bindings_v1"
	if v, _, _ := s.m.pg.GetSetting(flag); v == "true" {
		return
	}
	if err := s.m.pg.AddAgentToToolBinding("src_agent", []string{"insert_assets", "list_assets", "report_finding", "list_findings", "record_fact", "traffic_search", "traffic_get", "traffic_blob"}); err != nil {
		return
	}
	_ = s.m.pg.SetSetting(flag, "true")
}

// bindSRCTargetContext persists the explicit scope from the user's SRC turn
// before any agent turn is started. The executor later reloads this row; it does
// not infer authorization from titles, model text, DNS answers, or shared IPs.
func (s *Server) bindSRCTargetContext(taskID int64, scopeText string) (*srcagentpolicy.ModeContext, error) {
	parsed, err := srcagentpolicy.ParseTargetContext(scopeText)
	if err != nil {
		return nil, err
	}
	ctx := db.TaskTargetContext{
		Version:         1,
		Mode:            string(parsed.Mode),
		Target:          parsed.Target,
		TestRange:       append([]string(nil), parsed.TestRange...),
		Ports:           append([]int(nil), parsed.Ports...),
		AllowSubdomains: false,
	}
	for _, item := range parsed.TestRange {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(item)), "*.") {
			ctx.AllowSubdomains = true
		}
	}
	if err := s.m.pg.UpsertTaskTargetContext(taskID, ctx); err != nil {
		return nil, err
	}
	if as := s.m.Assets(); as != nil {
		for _, item := range parsed.TestRange {
			kind, value := "subdomain", item
			if strings.HasPrefix(strings.ToLower(item), "*.") {
				kind, value = "root_domain", strings.TrimPrefix(strings.ToLower(item), "*.")
			} else if strings.EqualFold(strings.TrimSuffix(item, "."), strings.TrimSuffix(parsedTargetHost(parsed.Target), ".")) {
				kind = "root_domain"
			}
			if _, err := as.AddAgentScope(taskID, kind, value, "SRC 用户明确 Test Range", "manual"); err != nil {
				return nil, err
			}
		}
	}
	policy := srcagentpolicy.ModeContext{Mode: parsed.Mode, Target: parsed.Target, TestRange: parsed.TestRange, Ports: parsed.Ports}
	return &policy, nil
}

func parsedTargetHost(target string) string {
	ctx, err := srcagentpolicy.ParseTargetContext("Target: " + target)
	if err != nil || len(ctx.TestRange) == 0 {
		return ""
	}
	return ctx.TestRange[0]
}

func (s *Server) srcPolicyForTask(taskID int64) (*srcagentpolicy.ModeContext, error) {
	ctx, err := s.m.pg.TaskTargetContext(taskID)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		return &srcagentpolicy.ModeContext{Mode: srcagentpolicy.ModeSRC}, nil
	}
	return &srcagentpolicy.ModeContext{Mode: srcagentpolicy.Mode(ctx.Mode), Target: ctx.Target, TestRange: ctx.TestRange, Ports: ctx.Ports}, nil
}

// Caller serializes admission per conversation. Creating a task does not start
// the autonomous planner: the SRC conversation owns this run.
func (s *Server) ensureSRCConversationTask(c *db.Conversation) (*Task, error) {
	if c.AgentKey != "src_agent" {
		return nil, nil
	}
	if c.TaskID != nil {
		if task, ok := s.m.Task(strconv.FormatInt(*c.TaskID, 10)); ok {
			return task, nil
		}
		return nil, fmt.Errorf("SRC associated task is unavailable")
	}
	// Reload so follow-up turns and restored conversations reuse the association.
	stored, err := s.m.pg.GetConversation(c.ID)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return nil, fmt.Errorf("SRC conversation no longer exists")
	}
	if stored.TaskID != nil {
		c.TaskID = stored.TaskID
		return s.ensureSRCConversationTask(c)
	}
	task, err := s.m.CreateTaskWithOptions(c.Title, "SRC 授权研究：目标与测试范围待用户明确；仅非破坏性验证", db.TaskCreateOptions{Name: c.Title})
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		return nil, err
	}
	if _, err := s.m.pg.Exec(`UPDATE conversations SET task_id=$2 WHERE id=$1 AND task_id IS NULL`, c.ID, id); err != nil {
		return nil, err
	}
	c.TaskID = &id
	return task, nil
}
