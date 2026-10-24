import { getRequestConfig } from "next-intl/server";

import { type Locale, resolveLocale } from "./config";

// 仅加载中文和英文消息，避免把无关的历史语言包打进客户端。
const loaders: Record<Locale, () => Promise<{ default: Record<string, unknown> }>> = {
  zh: () => import("../../messages/zh.json"),
  en: () => import("../../messages/en.json"),
};

// next-intl 요청 설정. i18n 경로 라우팅(세그먼트·미들웨어)을 쓰지 않는 구성이라
// locale 을 여기서 직접 정한다. requestLocale 은 참조하지 않으므로 동적 렌더링을
// 유발하지 않는다. 이 설정은 서버 컴포넌트·프리렌더의 기본 언어만 정하고, 실행 중
// 사용자 전환은 클라이언트의 IntlRuntimeProvider 가 담당한다(messages/zh·en).
export default getRequestConfig(async () => {
  const locale = resolveLocale();
  const messages = (await loaders[locale]()).default;
  return { locale, messages };
});
