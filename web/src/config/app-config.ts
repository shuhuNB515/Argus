import { DEFAULT_LOCALE, type Locale, resolveLocale } from "@/i18n/config";

import packageJson from "../../package.json";

const currentYear = new Date().getFullYear();

// 浏览器选项卡标题·搜索引擎元描述. layout.tsx 静态的 metadata 进入全部
// 页面的 <title>·<meta name="description"> 因为它是渲染在, 构建基础知识 locale 符合
// 分支为五种语言. zh 保留原始文本以供上游比较。. 标题不是 M 破折号。
// 使用冒号 "产品名称: 描述" 成形. 这个元在构建时是一个静态值，所以它
// 不响应转换(这是允许的，因为选项卡标题只有一行长。).
const META_BY_LOCALE: Record<Locale, { title: string; description: string }> = {
  en: {
    title: "Argus: Autonomous Security Research Console",
    description: "An LLM-driven system console for autonomous penetration testing",
  },
  zh: {
    title: "Argus：安全研究工作台",
    description: "Argus SRC、CTF 与靶场安全研究工作台",
  },
};

export const APP_CONFIG = {
  name: "Argus",
  version: packageJson.version,
  copyright: `© ${currentYear}, Argus.`,
  meta: META_BY_LOCALE[resolveLocale()] ?? META_BY_LOCALE[DEFAULT_LOCALE],
};
