package agent

import (
	"strings"

	"github.com/Autumn-27/artex/aoci"
)

func aociContextSpec(workDir string) string {
	context := strings.TrimSpace(aoci.RuntimeContextFor(workDir))
	if context == "" {
		return ""
	}
	status := aoci.RuntimeStatusFor(workDir)
	return "\n\n**AOCI-CODE Whole-Index（当前工作区项目认知）**（runtime_state=" + status.State + ", fingerprint=" + status.Fingerprint + "）：\n```text\n" + context + "\n```"
}
