package agent

// 本文件负责选择面向用户的输出语言。Argus 只提供简体中文和英文。

// OutputLanguage 由服务器注入，返回运行时设置的语言代码（"zh" 或 "en"）。
// 未注入或返回空值时使用中文默认值。
var OutputLanguage func() string

const defaultOutputLang = "zh"

// SupportedOutputLangs 是可切换输出语言代码的列表。.
var SupportedOutputLangs = []string{"zh", "en"}

// normalizeLang 将支持列表之外的代码范围缩小到默认值.
func normalizeLang(code string) string {
	for _, c := range SupportedOutputLangs {
		if c == code {
			return code
		}
	}
	return defaultOutputLang
}

// resolveOutputLang 返回当前应用的输出语言代码。. 首先，连接服务器接线的挂钩。
// 报告, 如果没有，则使用默认值.
func resolveOutputLang() string {
	if OutputLanguage != nil {
		if code := OutputLanguage(); code != "" {
			return normalizeLang(code)
		}
	}
	return defaultOutputLang
}

// langPhrase 是 (用中文写的) 在指令范围内 "用这种语言写"这是指出时使用的符号.
// 将本地语言符号放在括号中，以避免对模型使用的语言产生混淆。.
var langPhrase = map[string]string{
	"en": "英语（English）",
	"zh": "中文（简体中文）",
}

// langBare 是在指令中重复指定时使用的简短语言名称。.
var langBare = map[string]string{
	"en": "英语",
	"zh": "中文",
}
