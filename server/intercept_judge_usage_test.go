package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shuhuNB515/Argus/db"
)

// 此测试是上游 ARGUS(0.3.15 系列)曾在
// GET /api/intercept/judge/usage 通过将其带回此叉子来添加在一起。.
// 后备决策(judge) 累计消耗的代币总数和按日期排列的序列实际上是 DB 正好在
// 算了吗？, 和 DB 即使没有 500 看看它是否响应一个空系列而不是这个.
//
// DB 如果没有配置，则会像任何其他集成测试一样被跳过。(server/testmain_test.go 注意事项).
//
// These tests cover the endpoint restored from upstream ARGUS: judge token totals
// and the per-day series, plus the no-database graceful path.
const judgeFixtureModel = "judge-usage-test-fixture"

// judgeUsageRig opens the shared test database and registers fixture cleanup.
func judgeUsageRig(t *testing.T) (*Server, *db.DB) {
	t.Helper()
	dsn, _, err := db.DSN()
	if err != nil {
		t.Skipf("no database config (%v) — skipping", err)
	}
	d, err := db.Open(dsn)
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	// llm_usage 是 schema.sql 不是这个 Go 侧面 EnsureLLMUsageTable() 这使得.
	// db 包中现有的测试(finding_traffic_archive_test.go 等等)匹配方式与.
	if err := d.EnsureLLMUsageTable(); err != nil {
		t.Fatalf("EnsureLLMUsageTable 失败: %v", err)
	}
	cleanup := func() {
		_, _ = d.Exec(`DELETE FROM llm_usage WHERE model = $1`, judgeFixtureModel)
	}
	cleanup()
	t.Cleanup(cleanup)
	return &Server{m: &Manager{pg: d}}, d
}

// callJudgeUsage invokes the handler directly (auth is covered by auth tests).
func callJudgeUsage(t *testing.T, s *Server, query string) (int, db.JudgeUsage) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/intercept/judge/usage"+query, nil)
	s.interceptJudgeUsage(rec, req)
	var out db.JudgeUsage
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("反应 JSON 解析失败: %v (body=%s)", err, rec.Body.String())
		}
	}
	return rec.Code, out
}

// insertUsage writes one ledger row. tsShiftDays > 0 backdates it for the window test.
func insertUsage(t *testing.T, d *db.DB, worker string, in, out, cacheRead, cacheWrite int, tsShiftDays int) {
	t.Helper()
	_, err := d.Exec(`
INSERT INTO llm_usage(worker, model, profile_name, input_tokens, output_tokens, cache_read, cache_write, status, ts)
VALUES ($1,$2,'fixture',$3,$4,$5,$6,'ok', now() - ($7 * interval '1 day'))`,
		worker, judgeFixtureModel, in, out, cacheRead, cacheWrite, tsShiftDays)
	if err != nil {
		t.Fatalf("fixture insert 失败: %v", err)
	}
}

// TestInterceptJudgeUsageAggregates: 累计总数为 worker='judge' 只计算行数, 在日期系列中
// 今天创建了一个存储桶, judge 除此以外的行不应影响任何值。.
func TestInterceptJudgeUsageAggregates(t *testing.T) {
	s, d := judgeUsageRig(t)

	base, err := d.JudgeUsageStats(30)
	if err != nil {
		t.Fatalf("baseline JudgeUsageStats 失败: %v", err)
	}

	insertUsage(t, d, "judge", 100, 20, 7, 3, 0)
	insertUsage(t, d, "judge", 50, 10, 0, 0, 0)
	insertUsage(t, d, "worker", 999, 999, 999, 999, 0) // 必须排除在计算之外

	code, got := callJudgeUsage(t, s, "")
	if code != http.StatusOK {
		t.Fatalf("status=%d want 200 (body=%s)", code, "")
	}
	if got.Calls-base.Calls != 2 {
		t.Errorf("calls delta=%d want 2", got.Calls-base.Calls)
	}
	if got.InputTokens-base.InputTokens != 150 {
		t.Errorf("input delta=%d want 150 (judge 混合行的可能性)", got.InputTokens-base.InputTokens)
	}
	if got.OutputTokens-base.OutputTokens != 30 {
		t.Errorf("output delta=%d want 30", got.OutputTokens-base.OutputTokens)
	}
	if got.CacheReadTokens-base.CacheReadTokens != 7 {
		t.Errorf("cache_read delta=%d want 7", got.CacheReadTokens-base.CacheReadTokens)
	}
	if got.CacheWriteTokens-base.CacheWriteTokens != 3 {
		t.Errorf("cache_write delta=%d want 3", got.CacheWriteTokens-base.CacheWriteTokens)
	}

	today := time.Now().UTC().Format("2006-01-02")
	var found bool
	for _, day := range got.Daily {
		if day.Date == today {
			found = true
			if day.Calls < 2 {
				t.Errorf("今天桶 calls=%d want >=2", day.Calls)
			}
			if day.InputTokens < 150 {
				t.Errorf("今天桶 input=%d want >=150", day.InputTokens)
			}
		}
	}
	if !found {
		t.Errorf("日期系列中的今天(%s) 无桶: %+v", today, got.Daily)
	}
}

// TestInterceptJudgeUsageDaysWindow: days 缩小到仅日期系列，累计总数为整个期间。.
func TestInterceptJudgeUsageDaysWindow(t *testing.T) {
	s, d := judgeUsageRig(t)

	insertUsage(t, d, "judge", 11, 1, 0, 0, 3) // 窗外(3几天前)
	insertUsage(t, d, "judge", 22, 2, 0, 0, 0) // 窗户里面

	_, recent := callJudgeUsage(t, s, "?days=1")
	cutoff := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	for _, day := range recent.Daily {
		if day.Date < cutoff {
			t.Errorf("days=1 但窗外有一个水桶(%s)这是返回的 (cutoff=%s)", day.Date, cutoff)
		}
	}
	if recent.Calls < 2 {
		t.Errorf("累计 calls=%d want >=2 (无论窗口如何，都必须计算整个周期)", recent.Calls)
	}

	_, all := callJudgeUsage(t, s, "?days=0") // 0 = 基本的 30通过工作纠正
	if len(all.Daily) > 31 {
		t.Errorf("days=0 修正后桶数异常多: %d", len(all.Daily))
	}
}

// TestInterceptJudgeUsageWithoutDatabase: DB 即使未连接，也可以防止设置屏幕被破坏。
// 200 + 回复空系列(判断本身已关闭).
func TestInterceptJudgeUsageWithoutDatabase(t *testing.T) {
	s := &Server{m: &Manager{}}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/intercept/judge/usage", nil)
	s.interceptJudgeUsage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", rec.Code)
	}
	var out db.JudgeUsage
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("反应 JSON 解析失败: %v", err)
	}
	if out.Calls != 0 || len(out.Daily) != 0 {
		t.Errorf("DB 如果不存在，则它一定是一个空系列。: %+v", out)
	}
}
