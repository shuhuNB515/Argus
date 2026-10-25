package server

import (
	"context"
	"encoding/json"
	"github.com/Autumn-27/norma/llm"
	"github.com/shuhuNB515/Argus/db"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A SRC conversation must have a persisted task before any model is called.
func TestSRCConversationCreatesTaskAndKeepsAssociation(t *testing.T) {
	s, _ := newRetestServer(t)
	s.seedChatAgents()
	w := httptest.NewRecorder()
	s.pgCreateConversation(w, httptest.NewRequest("POST", "/api/conversations", strings.NewReader(`{"agent_key":"src_agent","title":"SRC sync regression"}`)))
	if w.Code != 200 {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var got struct {
		ID     int64  `json:"id"`
		TaskID *int64 `json:"task_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.TaskID == nil || *got.TaskID == 0 {
		t.Fatal("SRC conversation has no task association")
	}
	c, err := s.m.pg.GetConversation(got.ID)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(c)
	var restored struct {
		TaskID *int64 `json:"task_id"`
	}
	_ = json.Unmarshal(b, &restored)
	if restored.TaskID == nil || *restored.TaskID != *got.TaskID {
		t.Fatal("association lost on reload")
	}
	if _, ok := s.m.Task(strings.TrimSpace(jsonNumber(*got.TaskID))); !ok {
		t.Fatal("task missing from manager")
	}
	stale, err := s.m.pg.GetConversation(got.ID)
	if err != nil {
		t.Fatal(err)
	}
	stale.TaskID = nil
	reused, err := s.ensureSRCConversationTask(stale)
	if err != nil || reused.ID != jsonNumber(*got.TaskID) {
		t.Fatalf("follow-up created another task: %+v %v", reused, err)
	}
	custom, err := s.m.pg.CreateConversation("ordinary_custom", "ordinary chat", nil)
	if err != nil {
		t.Fatal(err)
	}
	if task, err := s.ensureSRCConversationTask(custom); err != nil || task != nil {
		t.Fatalf("ordinary custom chat got an unsolicited task: %+v %v", task, err)
	}
}

func TestSRCContextIsPersistedAndLoadedByExecutor(t *testing.T) {
	s, _ := newRetestServer(t)
	s.seedChatAgents()
	s.m.assets = s.m.pg.Assets()
	c, err := s.m.pg.CreateConversation("src_agent", "SRC scope fixture", nil)
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.ensureSRCConversationTask(c)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := s.bindSRCTargetContext(*c.TaskID, "Target: https://shu560.cn/\nTest Range: shu560.cn\nPorts: 80,443")
	if err != nil {
		t.Fatal(err)
	}
	if err := policy.AuthorizeURL("https://shu560.cn/robots.txt"); err != nil {
		t.Fatalf("target was rejected: %v", err)
	}
	if err := policy.AuthorizeURL("https://www.shu560.cn/"); err == nil {
		t.Fatal("unlisted subdomain was authorized")
	}
	if err := policy.AuthorizeURL("https://shu560.cn:8443/"); err == nil {
		t.Fatal("unlisted port was authorized")
	}
	stored, err := s.m.pg.TaskTargetContext(*c.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if stored == nil || stored.Target != "https://shu560.cn/" || len(stored.TestRange) != 1 || len(stored.Ports) != 2 {
		t.Fatalf("stored context: %+v", stored)
	}
	loaded, err := s.srcPolicyForTask(*c.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if err := loaded.AuthorizeURL("https://shu560.cn/"); err != nil {
		t.Fatalf("executor did not reload target context: %v", err)
	}
	scope, err := s.m.Assets().ListTaskScope(*c.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if len(scope) != 1 || scope[0].Kind != "root_domain" || scope[0].Value != "shu560.cn" {
		t.Fatalf("unexpected task scope for %s: %+v", task.ID, scope)
	}
}

func TestPublicStreamsNeverExposePrivateReasoningOrUsage(t *testing.T) {
	s, _ := newRetestServer(t)
	s.seedChatAgents()
	c, err := s.m.pg.CreateConversation("src_agent", "SRC privacy fixture", nil)
	if err != nil {
		t.Fatal(err)
	}
	privateID, err := s.m.pg.AppendConvActivity(c.ID, db.Activity{Worker: "src_agent", Kind: "thinking", Summary: "private reasoning", Detail: "secret chain"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.m.pg.AppendConvActivity(c.ID, db.Activity{Worker: "src_agent", Kind: "usage", Summary: "private usage"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.m.pg.AppendConvActivity(c.ID, db.Activity{Worker: "src_agent", Kind: "text", Summary: "public summary"})
	if err != nil {
		t.Fatal(err)
	}
	pathID := strconv.FormatInt(c.ID, 10)
	req := httptest.NewRequest("GET", "/api/conversations/"+pathID+"/messages?since=0", nil)
	req.SetPathValue("id", pathID)
	w := httptest.NewRecorder()
	s.pgConversationMessages(w, req)
	if w.Code != 200 || strings.Contains(w.Body.String(), "private reasoning") || strings.Contains(w.Body.String(), "private usage") {
		t.Fatalf("conversation stream leaked private rows: %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "public summary") {
		t.Fatal("conversation stream dropped public row")
	}
	detailReq := httptest.NewRequest("GET", "/api/conversations/"+pathID+"/messages/"+strconv.FormatInt(privateID, 10)+"/detail", nil)
	detailReq.SetPathValue("id", pathID)
	detailReq.SetPathValue("seq", strconv.FormatInt(privateID, 10))
	detail := httptest.NewRecorder()
	s.pgConversationMsgDetail(detail, detailReq)
	if detail.Code != 200 || strings.Contains(detail.Body.String(), "secret chain") || !strings.Contains(detail.Body.String(), `"detail":""`) {
		t.Fatalf("conversation detail leaked private row: %d %s", detail.Code, detail.Body.String())
	}

	task, err := s.ensureSRCConversationTask(c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := task.Store.AppendActivity(db.Activity{Worker: "src_agent", Kind: "thinking", Summary: "task private", Detail: "task secret"}); err != nil {
		t.Fatal(err)
	}
	if _, err := task.Store.AppendActivity(db.Activity{Worker: "src_agent", Kind: "usage", Summary: "task usage"}); err != nil {
		t.Fatal(err)
	}
	if _, err := task.Store.AppendActivity(db.Activity{Worker: "src_agent", Kind: "text", Summary: "task public"}); err != nil {
		t.Fatal(err)
	}
	streamReq := httptest.NewRequest("GET", "/api/exploration/activity/stream?task="+task.ID+"&since=0", nil)
	streamCtx, cancel := context.WithCancel(streamReq.Context())
	streamReq = streamReq.WithContext(streamCtx)
	cancel()
	streamW := httptest.NewRecorder()
	s.streamActivity(streamW, streamReq)
	if streamW.Code != 200 || strings.Contains(streamW.Body.String(), "task private") || strings.Contains(streamW.Body.String(), "task usage") || strings.Contains(streamW.Body.String(), "task secret") {
		t.Fatalf("task stream leaked private rows: %d %s", streamW.Code, streamW.Body.String())
	}
	if !strings.Contains(streamW.Body.String(), "task public") {
		t.Fatal("task stream dropped public row")
	}
}

func TestSRCTurnWritesTaskAssetsEventsAndDraft(t *testing.T) {
	s, _ := newRetestServer(t)
	s.seedChatAgents()
	s.m.assets = s.m.pg.Assets()
	c, err := s.m.pg.CreateConversation("src_agent", "SRC linked fixture", nil)
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.ensureSRCConversationTask(c)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	setRetestProvider(s, retestProvider{complete: func(_ context.Context, req llm.CompletionRequest) (llm.Message, string, llm.Usage, error) {
		calls++
		if calls == 1 {
			return llm.Message{Role: llm.RoleAssistant, Content: []llm.ContentBlock{{Type: llm.BlockToolUse, ID: "asset-1", Name: "insert_assets", Input: json.RawMessage(`{"assets":[{"type":"root_domain","domain":"src-fixture.example"}]}`)}}}, "tool_use", llm.Usage{}, nil
		}
		steps := []struct{ name, input string }{
			{"record_fact", `{"summary":"fixture observation","detail":"controlled regression data","evidence":"local fixture only","confidence":"observed"}`},
			{"report_finding", `{"vulnclass":"regression-fixture","severity":"low","summary":"controlled test candidate","evidence":"local synthetic regression fixture, not a live vulnerability"}`},
			{"list_findings", `{}`},
			{"list_assets", `{}`},
		}
		if calls <= 5 {
			step := steps[calls-2]
			return llm.Message{Role: llm.RoleAssistant, Content: []llm.ContentBlock{{Type: llm.BlockToolUse, ID: strconv.Itoa(calls), Name: step.name, Input: json.RawMessage(step.input)}}}, "tool_use", llm.Usage{}, nil
		}
		serialized, _ := json.Marshal(req.Messages)
		if !strings.Contains(string(serialized), `results`) {
			t.Error("asset tool result did not return to agent")
		}
		return llm.Message{Role: llm.RoleAssistant, Content: []llm.ContentBlock{llm.TextBlock("Draft: fixture completed; no confirmed vulnerabilities.")}}, "end_turn", llm.Usage{}, nil
	}})
	ctx, cancel := context.WithCancelCause(s.ctx)
	s.runConversationTurn(ctx, cancel, c, "Record the authorized fixture asset; make no network requests.", s.convBusyKey(c.ID))
	assets, err := s.m.assets.QueryByType("root_domain", 1000, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range assets {
		raw, _ := json.Marshal(a)
		if strings.Contains(string(raw), "src-fixture.example") {
			found = true
		}
	}
	if !found {
		t.Fatal("SRC asset not persisted")
	}
	events, _, err := task.Store.ActivityList(nil, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	tools, results := 0, 0
	for _, e := range events {
		if e.Kind == "tool_use" {
			tools++
		}
		if e.Kind == "tool_result" {
			results++
		}
	}
	if tools != 5 || results != 5 {
		t.Fatalf("task events: uses=%d results=%d", tools, results)
	}
	var candidates, confirmed int
	if err := s.m.pg.QueryRow(`SELECT count(*), count(*) FILTER (WHERE status='confirmed') FROM findings WHERE task_id=$1`, *c.TaskID).Scan(&candidates, &confirmed); err != nil {
		t.Fatal(err)
	}
	if candidates != 1 || confirmed != 0 {
		t.Fatalf("candidate persistence: count=%d confirmed=%d", candidates, confirmed)
	}
	intents, err := task.Store.ListByKind(db.KindIntent, 10)
	if err != nil || len(intents) != 1 || intents[0].State != "done" {
		t.Fatalf("work graph missing completed SRC turn: %+v %v", intents, err)
	}
	for _, e := range events {
		if e.NodeID == nil || *e.NodeID != intents[0].ID {
			t.Fatal("work event not attributed to SRC intent")
		}
	}
	draft, err := os.ReadFile(filepath.Join(s.m.dir, "tasks", task.ID, "src-draft.md"))
	if err != nil || !strings.Contains(string(draft), "fixture completed") {
		t.Fatalf("draft missing: %v", err)
	}
	r := httptest.NewRequest("GET", "/api/report?task="+task.ID, nil)
	w := httptest.NewRecorder()
	s.getReport(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "# SRC 报告") || !strings.Contains(w.Body.String(), "0 confirmed findings") {
		t.Fatalf("draft unavailable through report API: %d", w.Code)
	}
	for _, name := range []string{"src-report.md", "src-report.html", "src-report.pdf", "src-report.metadata.json"} {
		if info, err := os.Stat(filepath.Join(s.m.dir, "tasks", task.ID, "reports", name)); err != nil || info.Size() == 0 {
			t.Fatalf("report artifact %s missing: %v", name, err)
		}
	}
	status, err := s.m.pg.GetTask(*c.TaskID)
	if err != nil || status.Status != "done" {
		t.Fatalf("task completion not saved: %+v %v", status, err)
	}
	if _, err := strconv.ParseInt(task.ID, 10, 64); err != nil {
		t.Fatal(err)
	}
}

func jsonNumber(n int64) string { b, _ := json.Marshal(n); return string(b) }

func TestSRCTaskStopCancelsItsConversation(t *testing.T) {
	s, _ := newRetestServer(t)
	s.seedChatAgents()
	c, err := s.m.pg.CreateConversation("src_agent", "SRC stop fixture", nil)
	if err != nil {
		t.Fatal(err)
	}
	task, err := s.ensureSRCConversationTask(c)
	if err != nil {
		t.Fatal(err)
	}
	entered, done := make(chan struct{}), make(chan struct{})
	setRetestProvider(s, retestProvider{complete: func(ctx context.Context, _ llm.CompletionRequest) (llm.Message, string, llm.Usage, error) {
		close(entered)
		<-ctx.Done()
		return llm.Message{}, "", llm.Usage{}, ctx.Err()
	}})
	ctx, cancel := s.conversationRunContext(c.ID, s.convBusyKey(c.ID))
	defer cancel(context.Canceled)
	go func() {
		defer close(done)
		s.runConversationTurn(ctx, cancel, c, "wait for cancellation", s.convBusyKey(c.ID))
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("SRC run did not start")
	}
	if !s.cancelTaskChat(task.ID, context.Canceled) {
		t.Fatal("task stop cannot find its SRC conversation")
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("SRC run did not stop")
	}
	stored, err := s.m.pg.GetTask(*c.TaskID)
	if err != nil || stored.Status != "paused" {
		t.Fatalf("stopped task state: %+v %v", stored, err)
	}
}
