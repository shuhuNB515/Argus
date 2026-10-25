// Argus 支持中文和英文两种界面语言。
export const LOCALES = ["zh", "en"] as const;
export type Locale = (typeof LOCALES)[number];

// 默认显示语言为中文，也可以在设置中切换为英文。
export const DEFAULT_LOCALE: Locale = "zh";

// 每个都显示在语言选择器中 locale 的母语符号. 无论您使用哪种语言查看它，它都是您自己的语言。
// 名称必须用该语言阅读，因此不需要翻译，而是保留固定的拼写。.
export const LOCALE_LABELS: Record<Locale, string> = {
  en: "English",
  zh: "中文",
};

// 存储用户选择的显示语言的键. 屏幕是 localStorage 读取值. 饼干也一样
// 我写下了该值，但现在没有地方可以读取它。(当服务器需要确定语言时保留。).
export const LOCALE_COOKIE = "argus_locale";
export const LOCALE_STORAGE_KEY = "argus_locale";

// 支持随机字符串 locale 缩小意识的警卫.
export function isLocale(raw: unknown): raw is Locale {
  return typeof raw === "string" && (LOCALES as readonly string[]).includes(raw);
}

// 活跃的 locale 在构建时决定。. 静态导出(next.config 的 output: "export")哇
// 因为它们必须兼容 cookies()·headers() 同样的动态 API 只读取环境变量而不使用.
// NEXT_PUBLIC_LOCALE 如果它为空或在支持的列表之外，它将回退到默认值。. 该值是服务器
// 自由渲染的早期阶段 HTML 它仅用于, 实际的用户显示语言由客户端上的 cookie 覆盖。.
export function resolveLocale(): Locale {
  const raw = process.env.NEXT_PUBLIC_LOCALE;
  return isLocale(raw) ? raw : DEFAULT_LOCALE;
}
