package intercept

import "github.com/shuhuNB515/Argus/config"

// 拦截判定动作标签（allow/deny/ask）的语言表。输出语言来自 config.ActiveLanguage；
// 未配置或不支持的语言回退到中文。枚举键和 [模型] 标记保持稳定，不参与翻译。
var judgeActionLabels = map[string]map[string]string{
	"en": {"allow": "allow", "deny": "deny", "ask": "ask"},
	"zh": {"allow": "允许", "deny": "拦截", "ask": "询问"},
}

func judgeActionLang() string {
	lang := config.ActiveLanguage()
	if _, ok := judgeActionLabels[lang]; ok {
		return lang
	}
	return "zh"
}
