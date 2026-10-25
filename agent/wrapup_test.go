package agent

import (
	"strings"
	"testing"
	"unicode"
)

// A2: wrap-up / settlement 提示翻译成中文.
//
// 这些常数是 run 或 task 跌倒的脚步/当达到时间预算并结束时 settlement 步调一致
// 注入, 直接指示向用户展示的最终摘要。. 因此 (1) 中文
// 必须写, (2) 中国人(CJK 汉字) 不应有残留, (3) 工具名称和
// "一句话纯文本" 必须保留相同的参考意义.
//
// DB 枯萎 wrapup_prompt / task_timeout_wrapup_prompt 保留为空字符串(db/db.go 的
// builtin 代理人 INSERT 不填写此栏), 如果为空，则降至该常数.
// 也就是说，这些常数 wrap-up 是短语的唯一来源.
func TestWrapupPromptsLocalizedToChinese(t *testing.T) {
	all := map[string]string{
		"settleWrapUpPrompt":        settleWrapUpPrompt,
		"plannerWrapUpDefault":      plannerWrapUpDefault,
		"mainAgentWrapUpDefault":    mainAgentWrapUpDefault,
		"genericWrapUpDefault":      genericWrapUpDefault,
		"workerTaskTimeoutDefault":  workerTaskTimeoutDefault,
		"plannerTaskTimeoutDefault": plannerTaskTimeoutDefault,
	}

	hasScript := func(s string, table *unicode.RangeTable) bool {
		for _, r := range s {
			if unicode.Is(table, r) {
				return true
			}
		}
		return false
	}

	for name, p := range all {
		if !hasScript(p, unicode.Han) {
			t.Errorf("%s: 它没有被翻译成中文，因为根本没有中文字符。", name)
		}
	}

	// 工具名称是标识符，应保留而不翻译。.
	// worker 系列(per-run·task-timeout)银 record_fact 结束语, report_finding 到
	// 写一个漏洞, 维护指令在最后用一句话纯文本进行总结.
	mustContain := func(name, p string, subs ...string) {
		for _, s := range subs {
			if !strings.Contains(p, s) {
				t.Errorf("%s: 指令意义 %q 应该保留但不存在", name, s)
			}
		}
	}
	mustContain("settleWrapUpPrompt", settleWrapUpPrompt,
		"insert_assets", "record_fact", "report_finding", "一句话", "纯文本")
	mustContain("workerTaskTimeoutDefault", workerTaskTimeoutDefault,
		"insert_assets", "record_fact", "report_finding", "一句话", "纯文本")
	mustContain("genericWrapUpDefault", genericWrapUpDefault, "一句话", "纯文本")
	mustContain("mainAgentWrapUpDefault", mainAgentWrapUpDefault, "一句话", "纯文本")
	// planner 没有提供总结句子(只要做出决定就结束了) 意图·目标·维护待办事项工具.
	mustContain("plannerWrapUpDefault", plannerWrapUpDefault, "add_intent", "prove_goal", "TodoWrite")
	mustContain("plannerTaskTimeoutDefault", plannerTaskTimeoutDefault, "prove_goal")
}

// per-run 和 task-timeout 应该有不同的含义(尤其是 planner): per-run 银 "就这一轮
// 结束了"并且 task-timeout 银 "整个工作完成了"是. 常量映射不得更改和混合.
func TestWrapupDefaultsRouting(t *testing.T) {
	if WrapupDefault("worker") != settleWrapUpPrompt {
		t.Error("worker per-run 默认是 settleWrapUpPrompt 不")
	}
	if WrapupDefault("planner") != plannerWrapUpDefault {
		t.Error("planner per-run 默认是 plannerWrapUpDefault 不")
	}
	if WrapupDefault("mainagent") != mainAgentWrapUpDefault {
		t.Error("mainagent per-run 默认是 mainAgentWrapUpDefault 不")
	}
	// 未注册的密钥(定制代理)是 generic 下降到.
	if WrapupDefault("unknown-agent") != genericWrapUpDefault {
		t.Error("未注册的密钥 genericWrapUpDefault 不落到")
	}
	// task-timeout 银 worker/planner 只有, 否则为空字符串(寻呼机 per-run 返回到).
	if TaskTimeoutWrapupDefault("worker") != workerTaskTimeoutDefault {
		t.Error("worker task-timeout 默认是 workerTaskTimeoutDefault 不")
	}
	if TaskTimeoutWrapupDefault("planner") != plannerTaskTimeoutDefault {
		t.Error("planner task-timeout 默认是 plannerTaskTimeoutDefault 不")
	}
	if TaskTimeoutWrapupDefault("mainagent") != "" {
		t.Error("mainagent 银 task-timeout 不应该有文字(空字符串)")
	}
	// per-run 和 task-timeout 如果短语相同，语义区别就会消失。.
	if workerTaskTimeoutDefault == settleWrapUpPrompt {
		t.Error("worker 的 per-run 和 task-timeout 措辞是一样的")
	}
	if plannerTaskTimeoutDefault == plannerWrapUpDefault {
		t.Error("planner 的 per-run 和 task-timeout 措辞是一样的")
	}
}
