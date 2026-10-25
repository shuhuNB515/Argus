/**
 * Script: check-i18n-keys.ts
 *
 * 检查 web/messages/{zh,en}.json 的键是否一致。
 * 如有差异，-0 结束于. 跑: npm run check:i18n
 */

import { diffKeys, flattenKeys, isInSync } from "./i18n-keys";
import fs from "node:fs";
import path from "node:path";

const MESSAGES_DIR = path.resolve(__dirname, "../../messages");
const BASELINE = "zh";
const LOCALES = ["en"] as const;

function readLocale(locale: string): Set<string> {
  const file = path.join(MESSAGES_DIR, `${locale}.json`);
  const tree = JSON.parse(fs.readFileSync(file, "utf8")) as unknown;
  return flattenKeys(tree);
}

function preview(keys: string[]): string {
  const shown = keys.slice(0, 20).join(", ");
  return keys.length > 20 ? `${shown} …` : shown;
}

const baseline = readLocale(BASELINE);
let failed = false;

console.log("─".repeat(64));
console.log(`i18n 密钥一致性检查 (标准: ${BASELINE}.json, ${baseline.size} 钥匙)`);
console.log("─".repeat(64));

for (const locale of LOCALES) {
  const diff = diffKeys(baseline, readLocale(locale));
  if (isInSync(diff)) {
    console.log(`[OK] ${locale}.json — 比赛`);
    continue;
  }
  failed = true;
  console.log(`[FAIL] ${locale}.json — 不匹配`);
  if (diff.missing.length) console.log(`   失踪 ${diff.missing.length}: ${preview(diff.missing)}`);
  if (diff.extra.length) console.log(`   添加 ${diff.extra.length}: ${preview(diff.extra)}`);
}

console.log("─".repeat(64));
if (failed) {
  console.error("i18n 高度没了. 失踪/匹配其他键并再次运行.");
  process.exit(1);
}
console.log("所有区域设置的键集都匹配.");
