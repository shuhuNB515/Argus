import { DEFAULT_LOCALE, isLocale, LOCALE_STORAGE_KEY, type Locale, resolveLocale } from "@/i18n/config";
import { getLocalStorageValue } from "@/lib/local-storage.client";

// 日期时间格式跟随当前界面语言。
const DATE_LOCALE: Record<Locale, string> = {
  en: "en-US",
  zh: "zh-CN",
};

export function getDateLocale(): string {
  if (typeof window === "undefined") return DATE_LOCALE[DEFAULT_LOCALE];
  const raw = getLocalStorageValue(LOCALE_STORAGE_KEY);
  const loc: Locale = isLocale(raw) ? raw : resolveLocale();
  return DATE_LOCALE[loc];
}
