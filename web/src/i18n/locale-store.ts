"use client";

import * as React from "react";

import { setClientCookie } from "@/lib/cookie.client";
import { getLocalStorageValue, setLocalStorageValue } from "@/lib/local-storage.client";

import { isLocale, LOCALE_COOKIE, LOCALE_STORAGE_KEY, type Locale, resolveLocale } from "./config";

// 用户选择的屏幕显示语言纯粹是前端环境设置，因此不会存储在服务器上。
// localStorage(和饼干)只留下. 它不会与您的帐户同步，因此您需要更改浏览器或更改站点。
// 清除数据返回默认语言. 代理输出语言单独存储在服务器设置中.

function parseLocale(raw: string | null): Locale {
  return isLocale(raw) ? raw : resolveLocale();
}

// 同一选项卡中的一组订阅者. storage 该活动是 "其他" 因为它只发生在选项卡上, 在当前选项卡中
// 更改语言后 emit 您应该通知同一选项卡的订阅者，以便立即重新呈现他们。.
const listeners = new Set<() => void>();

function subscribe(listener: () => void) {
  listeners.add(listener);
  window.addEventListener("storage", listener);
  return () => {
    listeners.delete(listener);
    window.removeEventListener("storage", listener);
  };
}

function emit() {
  for (const listener of listeners) listener();
}

function getSnapshot(): Locale {
  return parseLocale(getLocalStorageValue(LOCALE_STORAGE_KEY));
}

// 在服务器上 localStorage 由于没有第一次构建默认值 locale 并渲染, hydrate 之后
// getSnapshot Correct to the value selected by this user.
function getServerSnapshot(): Locale {
  return resolveLocale();
}

// 用于订阅活动显示语言的挂钩. 如果语言发生变化，使用此钩子的组件将被重新渲染。.
export function useActiveLocale(): Locale {
  return React.useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}

// 更改显示语言. 饼干·localStorage 写入并通知同一选项卡的订阅者. 1维持多年
// 为您下次访问留下选择.
export function setActiveLocale(locale: Locale) {
  if (!isLocale(locale)) return;
  setLocalStorageValue(LOCALE_STORAGE_KEY, locale);
  setClientCookie(LOCALE_COOKIE, locale, 365);
  emit();
}
