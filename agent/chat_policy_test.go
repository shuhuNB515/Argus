package agent

import (
	"context"
	"encoding/json"
	"testing"

	actool "github.com/Autumn-27/norma/tool"
	"github.com/shuhuNB515/Argus/srcagentpolicy"
)

func TestCheckPolicyValueRejectsOutOfRangeURLs(t *testing.T) {
	p := srcagentpolicy.ModeContext{Mode: srcagentpolicy.ModeSRC, Target: "shu560.cn", TestRange: []string{"shu560.cn"}}
	if err := checkPolicyValue(p, map[string]any{"command": "curl https://shu560.cn/"}); err != nil {
		t.Fatalf("allowed target rejected: %v", err)
	}
	if err := checkPolicyValue(p, map[string]any{"command": "curl https://outside.example/"}); err == nil {
		t.Fatal("out-of-range URL was allowed")
	}
}

func TestNetworkToolSetUsesCanonicalPolicyForBareHostAndURL(t *testing.T) {
	p := srcagentpolicy.ModeContext{Mode: srcagentpolicy.ModeSRC, Target: "https://shu560.cn/", TestRange: []string{"shu560.cn"}, Ports: []int{80, 443}}
	tool := actool.Build(actool.Spec{Name: "traffic_search", Run: func(context.Context, json.RawMessage, *actool.ToolContext) (actool.Result, error) {
		return actool.Text("executed"), nil
	}})
	wrapped := wrapAuthorizedNetworkTools([]actool.CoreTool{tool}, p)
	result, err := wrapped[0].Call(context.Background(), json.RawMessage(`{"host":"outside.example"}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := result.Flatten(); got == "executed" {
		t.Fatal("out-of-range network tool executed")
	}
	result, err = wrapped[0].Call(context.Background(), json.RawMessage(`{"host":"shu560.cn"}`), nil)
	if err != nil || result.Flatten() != "executed" {
		t.Fatalf("in-range network tool blocked: %v %q", err, result.Flatten())
	}
}
