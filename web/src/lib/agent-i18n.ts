import { useCallback, useEffect, useMemo, useState } from "react";

import { useTranslations } from "next-intl";

import { api } from "@/lib/api";
import type { Agent } from "@/lib/types";

// 默认代理(builtin)的名字·描述·提示变量标签是 DB 种子数据
// 只有中文版的(db/db.go 的 builtinAgents, server/finding_retests.go 的
// 漏洞复测 代理人). DB 该值不能是每种语言，并且种子是 ON CONFLICT DO NOTHING 伊拉
// 这不会反映在已安装的实例中。, 在屏幕上 key 到消息文件
// (web/messages/{ko,en,zh,es,ja}.json 的 builtinAgents 名称空间)找到翻译
// 告诉我, 如果没有 DB 按原样写入值. 用户创建的代理始终 DB 值.

// retester 是 DB 在 builtin=false 虽然播种的是(server FindingRetestAgentKey) 系统
// 种植固定 key 包含在翻译目标中. builtin=true 代理人(goals/planner/
// mainagent/worker/auto/pentest)是 agent.builtin 既然已经过滤掉了，就没有必要放在这里了。.
const SYSTEM_SEEDED_AGENT_KEYS = new Set(["retester"]);

function isSystemSeeded(agent: { key: string; builtin?: boolean }): boolean {
  return !!agent.builtin || SYSTEM_SEEDED_AGENT_KEYS.has(agent.key);
}

// useAgentLabels 是默认代理的名称·描述·以当前语言提示变量标签
// 它是返回的一组助手。. 在每个显示点，我们通过这个助手获得一个标签.
export function useAgentLabels() {
  const t = useTranslations("builtinAgents");

  // 调用者使用这些函数 useMemo·useCallback 因为它是作为依赖使用的(聊天群·过滤器计算),
  // 语言(t)仅在更改时创建新包. 每次渲染新对象时都会重复计算。.
  return useMemo(() => {
    // name 如果默认代理名称是, 或者 DB 姓名(agent.name)回报.
    const name = (agent: { key: string; name: string; builtin?: boolean }): string => {
      if (isSystemSeeded(agent) && t.has(`${agent.key}.name`)) return t(`${agent.key}.name`);
      return agent.name;
    };

    // description 如果默认代理翻译描述, 或者 DB 描述(如果不是，则为空字符串)到
    // 我把它还给你. 空描述后备短语将像以前一样粘贴在调用方。.
    const description = (agent: { key: string; description?: string; builtin?: boolean }): string => {
      if (isSystemSeeded(agent) && t.has(`${agent.key}.description`)) return t(`${agent.key}.description`);
      return agent.description ?? "";
    };

    // nameByKey 是 Agent 没有对象 key 当只有(群聊等) 这是写作的一种变体.
    // 观众(agent 一个返回的函数)如果收到请翻译, 如果没有 key 回报.
    const nameByKey = (
      key: string,
      lookup: (key: string) => { key: string; name: string; builtin?: boolean } | undefined,
    ): string => {
      const agent = lookup(key);
      // biome-ignore lint/nursery/useNullishCoalescing: 名称为空("") 代理就跟以前一样 key 因为你必须展示它 || 是预期的行为.
      return (agent && name(agent)) || key;
    };

    // varLabel 是默认代理的提示变量标签(是的: goals 的 "任务总目标")到
    // 翻译. 如果没有 DB 标签(fallback)按原样写.
    const varLabel = (agentKey: string, builtin: boolean, varName: string, fallback: string): string => {
      if ((builtin || SYSTEM_SEEDED_AGENT_KEYS.has(agentKey)) && t.has(`${agentKey}.vars.${varName}`)) {
        return t(`${agentKey}.vars.${varName}`);
      }
      return fallback;
    };

    return { name, description, nameByKey, varLabel };
  }, [t]);
}

// 代理列表位于屏幕上的多个位置。(审批记录等)这个 key → 因为它仅用于显示名称转换
// 接收一次并在模块中共享. If it fails, you get it again on the next call..
let agentListPromise: Promise<Agent[]> | null = null;
function loadAgentList(): Promise<Agent[]> {
  if (!agentListPromise) {
    agentListPromise = api.agents().catch((err) => {
      agentListPromise = null;
      throw err;
    });
  }
  return agentListPromise;
}

// useAgentNameByKey 无代理对象 key 屏幕仅(批准·拦截记录
// conv_agent_key 等等)返回使用的转换器. 默认代理是当前语言名称,
// 用户代理是 DB 姓名, 如果它不在列表中或者您还没有收到 key 按原样显示.
export function useAgentNameByKey(): (key: string) => string {
  const labels = useAgentLabels();
  const [agents, setAgents] = useState<Map<string, Agent>>(() => new Map());
  useEffect(() => {
    let alive = true;
    loadAgentList()
      .then((list) => {
        if (alive) setAgents(new Map(list.map((a) => [a.key, a])));
      })
      .catch(() => {
        // 如果您没有收到列表 key 按原样显示(在下一个安装上重试).
      });
    return () => {
      alive = false;
    };
  }, []);
  return useCallback((key: string) => labels.nameByKey(key, (k) => agents.get(k)), [labels, agents]);
}
