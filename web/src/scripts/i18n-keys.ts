/**
 * i18n 区域设置键一致性助手.
 *
 * zh.json 比较每个语言环境的扁平键集(missing)和
 * 添加(extra)找到所有. CLI(check-i18n-keys.ts)哇 Vitest 测试这个纯粹的逻辑
 * 共享，因此检查规则仅在一处定义.
 */

export interface KeyDiff {
  missing: string[]; // baseline 有 target 钥匙不在
  extra: string[]; // baseline 没有 target 有一把钥匙在
}

/** 筑巢 JSON 对象 "a.b.c" 使用一组以下形式的键进行展平. 安排·原始值被视为叶子. */
export function flattenKeys(tree: unknown, prefix = ""): Set<string> {
  const out = new Set<string>();
  if (tree === null || typeof tree !== "object" || Array.isArray(tree)) {
    if (prefix) out.add(prefix);
    return out;
  }
  for (const [key, value] of Object.entries(tree as Record<string, unknown>)) {
    const path = prefix ? `${prefix}.${key}` : key;
    for (const leaf of flattenKeys(value, path)) out.add(leaf);
  }
  return out;
}

/** baseline 对比 target 遗漏·以排序数组的形式返回附加键. */
export function diffKeys(baseline: Set<string>, target: Set<string>): KeyDiff {
  const missing = [...baseline].filter((k) => !target.has(k)).sort();
  const extra = [...target].filter((k) => !baseline.has(k)).sort();
  return { missing, extra };
}

/** 失踪·If all additions are empty true. */
export function isInSync(diff: KeyDiff): boolean {
  return diff.missing.length === 0 && diff.extra.length === 0;
}
