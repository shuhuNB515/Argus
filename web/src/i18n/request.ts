import { getRequestConfig } from "next-intl/server";

import { type Locale, resolveLocale } from "./config";

// 仅加载中文和英文消息，避免把无关的历史语言包打进客户端。
const loaders: Record<Locale, () => Promise<{ default: Record<string, unknown> }>> = {
  zh: () => import("../../messages/zh.json"),
  en: () => import("../../messages/en.json"),
};

// next-intl 请求设置. i18n 路由选择(段·中间件)这是一个不使用的配置
// locale 直接在这里决定. requestLocale does not reference it, so dynamic rendering
// 不会导致. 此设置适用于服务器组件·只需设置 Freerender 的默认语言即可, 跑步
// 用户转化是客户的 IntlRuntimeProvider 负责(messages/zh·en).
export default getRequestConfig(async () => {
  const locale = resolveLocale();
  const messages = (await loaders[locale]()).default;
  return { locale, messages };
});
