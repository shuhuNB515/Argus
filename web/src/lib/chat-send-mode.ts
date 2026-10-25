"use client";

import * as React from "react";

import { useTranslations } from "next-intl";

import { getLocalStorageValue, setLocalStorageValue } from "@/lib/local-storage.client";

// 会话输入窗口的传输/换行键组合. 这是一个纯粹的前端配置。 localStorage 仅
// 已保存，不存储在服务器上，也不与您的帐户同步. 所以如果你更换浏览器
// 需要重置. issue #39 注意事项: 0.3.2 去 Ctrl+Enter 转移 Enter 我把它改成了传输,
// 在这里，旧的组合键将恢复为选项。.
export type ChatSendMode = "enter" | "ctrl-enter";

export const CHAT_SEND_MODE_KEY = "argus_chat_send_mode";
export const DEFAULT_CHAT_SEND_MODE: ChatSendMode = "enter";

// 设置屏幕上传输方法下拉列表中使用的选项. 标签与区域设置相关，因此模块
// 它作为钩子返回，而不是作为常量。(如果将中文硬编码到纯模块中， zh 语言环境是中文
// 它打破了). value 是一个稳定的标识符。 localStorage 钥匙·在动作判断中按原样使用。.
export function useChatSendModeOptions(): { value: ChatSendMode; label: string }[] {
  const t = useTranslations("chatSendMode.option");
  return [
    { value: "enter", label: t("enter") },
    { value: "ctrl-enter", label: t("ctrlEnter") },
  ];
}

function parseMode(raw: string | null): ChatSendMode {
  return raw === "ctrl-enter" || raw === "enter" ? raw : DEFAULT_CHAT_SEND_MODE;
}

// 同一选项卡中的一组订阅者. localStorage 的 storage 该活动是 "其他" 因为它只发生在选项卡上,
// 更改当前选项卡中的设置后， emit 必须在同一选项卡的输入窗口中通知. 否则
// 必须刷新才能反映.
const listeners = new Set<() => void>();

function subscribe(listener: () => void) {
  listeners.add(listener);
  window.addEventListener("storage", listener);
  return () => {
    listeners.delete(listener);
    window.removeEventListener("storage", listener);
  };
}

// 因为它返回一个字符串文字 Object.is 按值比较 useSyncExternalStore 是无限的
// 不要陷入循环.
function getSnapshot(): ChatSendMode {
  return parseMode(getLocalStorageValue(CHAT_SEND_MODE_KEY));
}

// 在服务器上 localStorage 由于没有人，我们首先渲染默认值并, hydrate 之后 getSnapshot 这个
// 修正值.
function getServerSnapshot(): ChatSendMode {
  return DEFAULT_CHAT_SEND_MODE;
}

export function useChatSendMode(): ChatSendMode {
  return React.useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}

export function setChatSendMode(mode: ChatSendMode) {
  setLocalStorageValue(CHAT_SEND_MODE_KEY, mode);
  for (const listener of listeners) listener();
}

// shouldSubmitOnKey 判断单次击键是否对应传输.
// isComposing / keyCode 229 是中文的·因为中文等输入设备是组合字符，
// 必须通过. 否则 Enter 当用 确认信件时，信件发送错误。.
// enter 模式是 Shift 除了 0.3.2 保持动作甚至字母相同。(设置
// 不改变的用户的操作手感不会改变。). ctrl-enter 模式是 Ctrl 和 Cmd(macOS)到
// 一起领取.
export function shouldSubmitOnKey(e: React.KeyboardEvent, mode: ChatSendMode): boolean {
  if (e.key !== "Enter") return false;
  if (e.nativeEvent.isComposing || e.nativeEvent.keyCode === 229) return false;
  if (mode === "ctrl-enter") return e.ctrlKey || e.metaKey;
  return !e.shiftKey;
}
