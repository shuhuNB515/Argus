package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	actool "github.com/Autumn-27/norma/tool"
	"github.com/shuhuNB515/Argus/srcagentpolicy"
)

// policyTool applies the SRC authorization boundary at the last point before
// a network-capable tool reaches the operating system or network.
type policyTool struct {
	actool.CoreTool
	policy srcagentpolicy.ModeContext
}

func (t *policyTool) Call(ctx context.Context, in json.RawMessage, tc *actool.ToolContext) (actool.Result, error) {
	var value any
	if json.Unmarshal(in, &value) == nil {
		if err := t.policy.CheckInput(value); err != nil {
			return actool.Errorf(err.Error()), nil
		}
	}
	return t.CoreTool.Call(ctx, in, tc)
}

func checkPolicyValue(policy srcagentpolicy.ModeContext, value any) error {
	if err := policy.CheckInput(value); err != nil {
		return fmt.Errorf("SRC 授权范围拒绝网络目标：%s", strings.TrimPrefix(err.Error(), "network target rejected: "))
	}
	return nil
}

// wrapAuthorizedNetworkTools applies the same policy to every known network
// capability, including tools supplied by traffic/MCP integrations. It is
// intentionally name-based at the registry boundary so a provider cannot make
// an unregistered textual shell instruction executable by changing its label.
func wrapAuthorizedNetworkTools(tools []actool.CoreTool, policy srcagentpolicy.ModeContext) []actool.CoreTool {
	out := make([]actool.CoreTool, 0, len(tools))
	for _, tool := range tools {
		if tool == nil {
			continue
		}
		if isNetworkCapability(tool.Name()) {
			out = append(out, &policyTool{CoreTool: tool, policy: policy})
		} else {
			out = append(out, tool)
		}
	}
	return out
}

func isNetworkCapability(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "bash" || n == "webfetch" || n == "web_fetch" || n == "traffic_search" {
		return true
	}
	for _, marker := range []string{"browser", "http", "fetch", "crawl", "recon", "probe", "nmap", "nuclei", "ffuf", "wget", "curl", "dns", "subdomain"} {
		if strings.Contains(n, marker) {
			return true
		}
	}
	return false
}
