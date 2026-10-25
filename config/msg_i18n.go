package config

// 启动阶段(DB 连接前)面向操作员的消息的翻译. 此时 settings 可以读取表
// 没有运行时设置(ActiveLanguage)这不是配置的默认语言(Language(): ARGUS_LANG >
// 配置文件 > en)替换基于. 操作所需的标识符(ARGUS_PG_DSN·database·dsn·
// host/user/dbname 等等)不要翻译并保留原文原样。. "zh" 是密钥本身，所以不要保存它
// 没有, 查找失败·ko 在某些情况下，会返回原始文本。.
var cfgCatalog = map[string]map[string]string{
	"环境变量 ARGUS_PG_DSN": {
		"en": "environment variable ARGUS_PG_DSN",
		"zh": "环境变量 ARGUS_PG_DSN",
	},
	"配置文件 %s (database.dsn)": {
		"en": "config file %s (database.dsn)",
		"zh": "配置文件 %s (database.dsn)",
	},
	"配置文件 %s (database 字段)": {
		"en": "config file %s (database fields)",
		"zh": "配置文件 %s (database 字段)",
	},
	"找不到数据库设置: 环境变量 ARGUS_PG_DSN 未设置, 配置文件 %s 江户 database (dsn 或 host/user/dbname) 没有设置. 请创建配置文件或设置环境变量并重试": {
		"en": "database configuration not found: environment variable ARGUS_PG_DSN is not set, and config file %s has no database (dsn or host/user/dbname) configuration. Create a config file or set the environment variable, then try again",
		"zh": "找不到数据库配置：环境变量 ARGUS_PG_DSN 未设置，配置文件 %s 中也没有 database（dsn 或 host/user/dbname）配置。请创建配置文件或设置环境变量后重试",
	},
}

// logCatalog holds operator-facing console log templates (startup + runtime
// diagnostics). Keyed by the Chinese source template; format verbs are preserved in
// the same positional order so fmt.Printf/log.Printf arguments still line up.
var logCatalog = map[string]map[string]string{
	"[config] 配置文件: %s": {
		"en": "[config] config file: %s", "zh": "[config] 配置文件: %s",
	},
	"[config] 配置文件: %s (没有文件 · 环境变量 ARGUS_PG_DSN 仅使用)": {
		"en": "[config] config file: %s (not found · using only environment variable ARGUS_PG_DSN)",
		"zh": "[config] 配置文件: %s (文件不存在 · 仅使用环境变量 ARGUS_PG_DSN)",
	},
	"[config] skill 目录: %s": {
		"en": "[config] skill directory: %s", "zh": "[config] skill 目录: %s",
	},
	"已暂停": {"en": "paused", "zh": "已暂停"},
	"已恢复": {"en": "resumed", "zh": "已恢复"},
	"[auth] 已将 JWT 密钥从 %s 迁移到 %s（移出可被发现的工作区）": {
		"en": "[auth] moved JWT key from %s to %s (out of the discoverable workspace)",
		"zh": "[auth] 已将 JWT 密钥从 %s 迁移到 %s（移出可被发现的工作区）",
	},
	"[auth] 已将新的 JWT 密钥写入 %s": {
		"en": "[auth] wrote new JWT key to %s", "zh": "[auth] 已将新的 JWT 密钥写入 %s",
	},
	"[auth] %s 少于 %d 个字符，已忽略。请使用设置令牌方式初始化": {
		"en": "[auth] ignoring %s because it is shorter than %d characters; initialise with the setup token instead",
		"zh": "[auth] %s 少于 %d 个字符，已忽略。请使用设置令牌方式初始化",
	},
	"[auth] %s 超过 %d 字节，已忽略。请使用设置令牌方式初始化": {
		"en": "[auth] ignoring %s because it exceeds %d bytes; initialise with the setup token instead",
		"zh": "[auth] %s 超过 %d 字节，已忽略。请使用设置令牌方式初始化",
	},
	"[auth] 无法读取密码状态，跳过首次运行初始化: %v": {
		"en": "[auth] could not read the password state; skipping first-run initialisation: %v",
		"zh": "[auth] 无法读取密码状态，跳过首次运行初始化: %v",
	},
	"[auth] 生成 %s 哈希失败: %v": {
		"en": "[auth] failed to hash %s: %v",
		"zh": "[auth] 生成 %s 哈希失败: %v",
	},
	"[auth] 保存 %s 失败: %v": {
		"en": "[auth] failed to store %s: %v",
		"zh": "[auth] 保存 %s 失败: %v",
	},
	"[auth] 已通过环境变量 %s 初始化管理员密码（用户名 Argus）": {
		"en": "[auth] initialised the admin password from environment variable %s (username Argus)",
		"zh": "[auth] 已通过环境变量 %s 初始化管理员密码（用户名 Argus）",
	},
	"[auth] 管理员密码尚未设置。设置令牌为环境变量 %s 的值（不输出到日志）": {
		"en": "[auth] the admin password has not been set yet. The setup token is the value of environment variable %s (not printed to the log)",
		"zh": "[auth] 管理员密码尚未设置。设置令牌为环境变量 %s 的值（不输出到日志）",
	},
	"[auth] 生成设置令牌失败: %v": {
		"en": "[auth] failed to generate the setup token: %v",
		"zh": "[auth] 生成设置令牌失败: %v",
	},
	"[auth] 管理员密码尚未设置。设置令牌: %s": {
		"en": "[auth] the admin password has not been set yet. Setup token: %s",
		"zh": "[auth] 管理员密码尚未设置。设置令牌: %s",
	},
	"[auth] 主屏幕(/setup)的 \"设置令牌\" Enter the above value in the field or POST /api/auth/init 文本的 setup_token 发送至. 该令牌不会传递到网络，只能在此控制台日志中看到": {
		"en": "[auth] enter the value above in the \"Setup token\" field on the first screen (/setup), or send it as setup_token in the POST /api/auth/init body. This token is never sent over the network and is only visible in this console log",
		"zh": "[auth] 请在首个页面（/setup）的“设置令牌”栏中输入上述值，或将其作为 POST /api/auth/init 请求体中的 setup_token 发送。该令牌不会通过网络传输，只能在此控制台日志中查看",
	},
	"[auth] 警告：管理员密码尚未设置，但服务器绑定在 %q，其他主机也可以连接。虽然没有设置令牌无法初始化，但在完成初始设置之前请勿将此端口暴露到外部网络（默认 127.0.0.1:8787）": {
		"en": "[auth] warning: the admin password has not been set yet and the server is bound to %q, so other hosts can connect. Initialisation still requires the setup token, but do not expose this port to external networks until initial setup is complete (default 127.0.0.1:8787)",
		"zh": "[auth] 警告：管理员密码尚未设置，但服务器绑定在 %q，其他主机也可以连接。虽然没有设置令牌无法初始化，但在完成初始设置之前请勿将此端口暴露到外部网络（默认 127.0.0.1:8787）",
	},
	"[auth] 拒绝初始设置：设置令牌不匹配 (client=%s)": {
		"en": "[auth] initial setup rejected: setup token mismatch (client=%s)",
		"zh": "[auth] 拒绝初始设置：设置令牌不匹配 (client=%s)",
	},
	"[auth] 已初始化管理员密码 (client=%s)": {
		"en": "[auth] admin password initialised (client=%s)",
		"zh": "[auth] 已初始化管理员密码 (client=%s)",
	},
	"[auth] 拒绝修改密码：当前密码不匹配 (client=%s)": {
		"en": "[auth] password change rejected: current password mismatch (client=%s)",
		"zh": "[auth] 拒绝修改密码：当前密码不匹配 (client=%s)",
	},
	"[auth] 已修改管理员密码，所有现有会话均已失效 (client=%s)": {
		"en": "[auth] admin password changed; all existing sessions are invalidated (client=%s)",
		"zh": "[auth] 已修改管理员密码，所有现有会话均已失效 (client=%s)",
	},
	"[auth] 登录失败 (client=%s)": {
		"en": "[auth] login failed (client=%s)",
		"zh": "[auth] 登录失败 (client=%s)",
	},
	"[notify] 后台通知发送循环已通过 %s 关闭（仅测试）— 不会发送漏洞 IM 通知": {
		"en": "[notify] background notification loop disabled via %s (test only) — vulnerability IM alerts will not be sent",
		"zh": "[notify] 后台通知发送循环已通过 %s 关闭（仅测试）— 不会发送漏洞 IM 通知",
	},
	"[mcp] 从 %s 发现并缓存了 %d 个工具": {
		"en": "[mcp] from %s, discovered and cached %d tool(s)",
		"zh": "[mcp] 从 %s 发现并缓存了 %d 个工具",
	},
	"[mcp] 启动时自动发现：读取列表失败：%v": {
		"en": "[mcp] auto-discovery at startup: failed to read the list: %v",
		"zh": "[mcp] 启动时自动发现：读取列表失败：%v",
	},
	"[mcp] 启动时自动发现 %s 失败：%v": {
		"en": "[mcp] auto-discovery of %s at startup failed: %v",
		"zh": "[mcp] 启动时自动发现 %s 失败：%v",
	},
	"[tools] 已将 orchestration/platform 工具 schema 刷新为代码默认值（一次性）": {
		"en": "[tools] refreshed orchestration/platform tool schemas to code defaults (one-time)",
		"zh": "[tools] 已将 orchestration/platform 工具 schema 刷新为代码默认值（一次性）",
	},
	"[reporter] 已预配置「%s」agent + finding 触发器": {
		"en": "[reporter] pre-configured the \"%s\" agent + finding trigger",
		"zh": "[reporter] 已预配置「%s」agent + finding 触发器",
	},
	"[custom-tool] 自动检测到 python 解释器：%s": {
		"en": "[custom-tool] auto-detected python interpreter: %s",
		"zh": "[custom-tool] 自动检测到 python 解释器：%s",
	},
	"[pg] 数据库配置来源：%s": {
		"en": "[pg] database configuration source: %s", "zh": "[pg] 数据库配置来源：%s",
	},
	"[mcp] browser MCP 抓包代理已连接：%s (CA %s)": {
		"en": "[mcp] browser MCP capture proxy attached: %s (CA %s)",
		"zh": "[mcp] browser MCP 抓包代理已连接：%s (CA %s)",
	},
	"[mcp] 已移除 browser MCP 抓包代理配置": {
		"en": "[mcp] browser MCP capture proxy configuration removed",
		"zh": "[mcp] 已移除 browser MCP 抓包代理配置",
	},
}

// trCfg translates a startup/operator message to the configured default language
// (Language()). Format templates are translated the same way: pass the Chinese
// template through trCfg before fmt.Sprintf/fmt.Errorf.
func trCfg(ko string) string {
	if m, ok := cfgCatalog[ko]; ok {
		if v, ok2 := m[Language()]; ok2 && v != "" {
			return v
		}
	}
	return ko
}

// T translates an operator-facing console log template to the configured default
// language (Language()). Exposed for cmd and server log.Printf call sites. Format
// verbs are preserved in order, so wrap the template only: log.Printf(config.T(ko), args...).
func T(ko string) string {
	if m, ok := logCatalog[ko]; ok {
		if v, ok2 := m[Language()]; ok2 && v != "" {
			return v
		}
	}
	return ko
}
