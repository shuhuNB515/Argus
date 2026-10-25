"use client";

import type { ReactNode } from "react";
import * as React from "react";

import { NextIntlClientProvider } from "next-intl";

import enMessages from "../../messages/en.json";
import zhMessages from "../../messages/zh.json";
import type { Locale } from "./config";
import { useActiveLocale } from "./locale-store";

// 两种语言消息静态打包，切换语言时无需重新构建或联网加载。
// 更改时重建或页面刷新, 所有短语都是当场选择，无需异步加载。
// 用语言再次呈现(静态导出 output: "export" 它还兼容).
const MESSAGES: Record<Locale, Record<string, unknown>> = {
  zh: zhMessages as Record<string, unknown>,
  en: enMessages as Record<string, unknown>,
};

// 执行环境(浏览器·自由渲染服务器)时区.
const TIME_ZONE = Intl.DateTimeFormat().resolvedOptions().timeZone;

// 提供者负责执行期间的语言切换. 根布局 NextIntlClientProvider 到
// 代表, 用户选择的活动 locale(饼干·localStorage)通过 messages 选择并供应.
export function IntlRuntimeProvider({ children }: Readonly<{ children: ReactNode }>) {
  const locale = useActiveLocale();

  // 当语言改变时 <html lang> 还匹配. 根布局最初种植在服务器上。 lang 到
  // 用客户选择覆盖, 屏幕阅读器·确保浏览器翻译以正确的语言运行.
  React.useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  // 与根布局的提供程序不同，它不继承服务器请求设置。 timeZone 到
  // 直接交给. 如果你陷入 next-intl 在这个预渲染中 ENVIRONMENT_FALLBACK 留下错误.
  return (
    <NextIntlClientProvider locale={locale} messages={MESSAGES[locale]} timeZone={TIME_ZONE}>
      {children}
    </NextIntlClientProvider>
  );
}
