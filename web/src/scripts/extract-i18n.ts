/**
 * Script: extract-i18n.ts
 *
 * ARGUS 韩化 — UI 字符串 1茶叶提取器.
 *
 * `src/` 以下全部 `.ts`/`.tsx` 到 TypeScript AST 解析它与, 暴露给用户
 * 仅提取硬编码的中文字符串. 提取目标分为三种类型：.
 *   1) 字符串文字示例) title: "仪表盘"
 *   2) 模板文字示例) `已删除 ${n} 个对话` → "已删除 {var0} 个对话"
 *   3) JSX 文本节点示例) <span>故障转移</span>
 * 锡(`//`, `/* *\/`)银 AST 由于它不是节点，因此被自动排除。. 做好事·在尾随评论中
 * 所含中文无需翻译。(BRIEF: 注释翻译是最低优先级).
 *
 * 输出(web/messages/):
 *   - zh.json          与命名空间嵌套 { 钥匙: "中文原版" }  (原始保留用于上游校对)
 *   - ko.json          相同的骨架, 其值为 "" (B3 充满了中文)
 *   - zh.sources.json  按键来源(文件:线)·类型·占位符 — B2 接线, B3 翻译, 用于漂移跟踪
 *
 * 重新提取是非破坏性的. extract 创建的顶级命名空间(基于文件路径:
 * app·components·lib 等等)仅更新代码和, 否则 "策展命名空间"
 * (手工翻译 nav·search·header 等等)按原样保留在现有文件中。. 规则:
 * 提取的命名空间是一个清单，因此无法手动修改。, 运行时翻译是精心策划的
 * 只放在命名空间中. 这可确保重新提取不会覆盖翻译。.
 *
 * 命名空间是从文件路径获取的. 是的) src/app/(main)/system/llm/page.tsx
 *   → app.main.system.llm.page  (删除路由组括号, _从文件夹中删除下划线)
 * 关键在原文中 sha1 前面 8睡觉. 订单·即使文件发生变化，相同的文本也会收到相同的密钥
 * 重做 diff 很小(有利于上游更新对比).
 *
 * 跑:  npm run extract:i18n
 */

import * as ts from "typescript";

import crypto from "node:crypto";
import fs from "node:fs";
import path from "node:path";

const SRC_DIR = path.resolve(__dirname, "..");
const MESSAGES_DIR = path.resolve(__dirname, "../../messages");
const REPO_WEB_DIR = path.resolve(__dirname, "../..");

/** 汉字(Han) 1是否包含多个字符. UI 确定翻译目标的唯一标准. */
const HAN = /\p{Script=Han}/u;

/** 要从提取中排除的目录(自己用解析工具·构建输出). */
const SKIP_DIRS = new Set(["scripts", "node_modules", ".next"]);

type Kind = "string" | "template" | "jsx";

interface Message {
  ns: string;
  key: string;
  text: string;
  kind: Kind;
  placeholders: string[]; // 模板的 ${...} 原始表达(仅供参考)
  occurrences: string[]; // "相对路径:线"
}

/** src 基本相对路径 → 用点连接的命名空间. */
function toNamespace(absFile: string): string {
  const rel = path.relative(SRC_DIR, absFile).replace(/\.[tj]sx?$/, "");
  return rel
    .split(path.sep)
    .map((seg) =>
      seg
        .replace(/^\((.*)\)$/, "$1") // (main) → main  (路线组)
        .replace(/^_/, "") // _components → components
        .replace(/[^A-Za-z0-9去-呵呵]+/g, "-")
        .replace(/^-+|-+$/g, ""),
    )
    .filter(Boolean)
    .join(".");
}

function hashKey(text: string): string {
  return crypto.createHash("sha1").update(text).digest("hex").slice(0, 8);
}

/** 目录递归遍历 .ts/.tsx 文件收集. */
function collectFiles(dir: string, out: string[] = []): string[] {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.isDirectory()) {
      if (SKIP_DIRS.has(entry.name)) continue;
      collectFiles(path.join(dir, entry.name), out);
    } else if (/\.(ts|tsx)$/.test(entry.name)) {
      out.push(path.join(dir, entry.name));
    }
  }
  return out;
}

/** 模板文字 "文字部分 + {varN}" 重建为. 返回 null = 文字部分没有中文字符. */
function reconstructTemplate(node: ts.TemplateExpression): { text: string; placeholders: string[] } | null {
  let text = node.head.text;
  const placeholders: string[] = [];
  node.templateSpans.forEach((span, i) => {
    placeholders.push(span.expression.getText());
    text += `{var${i}}${span.literal.text}`;
  });
  // 汉字 ${} 如果只是在里面而不是在固定文本中，, 汉字被留作内部节点。.
  if (!HAN.test(node.head.text) && !node.templateSpans.some((s) => HAN.test(s.literal.text))) {
    return null;
  }
  return { text, placeholders };
}

/** JSX 文本规范化: 去掉两端空格 + 内部连续空间/换行成一个空格(与浏览器渲染规则相同). */
function normalizeJsxText(raw: string): string {
  return raw.replace(/\s+/g, " ").trim();
}

const messages = new Map<string, Message>(); // "ns\u0000key" → Message
const scannedFiles = collectFiles(SRC_DIR).sort();
const capturedLines = new Map<string, Set<number>>(); // 文件 → 提取节点挂起的行号集
let fileWithHan = 0;

function record(ns: string, text: string, kind: Kind, placeholders: string[], relFile: string, line: number) {
  const key = hashKey(text);
  const id = `${ns}\u0000${key}`;
  let msg = messages.get(id);
  if (!msg) {
    msg = { ns, key, text, kind, placeholders, occurrences: [] };
    messages.set(id, msg);
  }
  const occ = `${relFile}:${line}`;
  if (!msg.occurrences.includes(occ)) msg.occurrences.push(occ);
}

for (const absFile of scannedFiles) {
  const source = fs.readFileSync(absFile, "utf8");
  if (!HAN.test(source)) continue;
  fileWithHan++;

  const relFile = path.relative(REPO_WEB_DIR, absFile);
  const ns = toNamespace(absFile);
  const scriptKind = absFile.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
  const sf = ts.createSourceFile(absFile, source, ts.ScriptTarget.Latest, true, scriptKind);
  const lineOf = (pos: number) => sf.getLineAndCharacterOfPosition(pos).line + 1;
  const markLine = (pos: number) => {
    const set = capturedLines.get(relFile) ?? new Set<number>();
    set.add(lineOf(pos));
    capturedLines.set(relFile, set);
  };

  const visit = (node: ts.Node): void => {
    if (ts.isStringLiteralLike(node) && !ts.isTemplateExpression(node.parent)) {
      // 字符串文字 + 模板无需替换. (带替换的模板 head/middle 不会被抓到这里)
      if (HAN.test(node.text)) {
        record(ns, node.text, "string", [], relFile, lineOf(node.getStart(sf)));
        markLine(node.getStart(sf));
      }
    } else if (ts.isTemplateExpression(node)) {
      const r = reconstructTemplate(node);
      if (r) {
        record(ns, r.text, "template", r.placeholders, relFile, lineOf(node.getStart(sf)));
        markLine(node.getStart(sf));
      }
    } else if (ts.isJsxText(node)) {
      const text = normalizeJsxText(node.text);
      if (text && HAN.test(text)) {
        record(ns, text, "jsx", [], relFile, lineOf(node.getStart(sf)));
        markLine(node.getStart(sf));
      }
    }
    ts.forEachChild(node, visit);
  };
  visit(sf);
}

// --- 筑巢 JSON 组装 ---------------------------------------------------------
type Tree = { [k: string]: Tree | string };

function setNested(root: Tree, nsPath: string, key: string, value: string) {
  const segs = nsPath.split(".");
  let node: Tree = root;
  for (const seg of segs) {
    if (typeof node[seg] !== "object") node[seg] = {};
    node = node[seg] as Tree;
  }
  node[key] = value;
}

/** 递归地对键进行排序并重新运行 diff 最小化. */
function sortTree(t: Tree): Tree {
  const out: Tree = {};
  for (const k of Object.keys(t).sort()) {
    const v = t[k];
    out[k] = typeof v === "string" ? v : sortTree(v);
  }
  return out;
}

const zhTree: Tree = {};
const koTree: Tree = {};
const sources: Record<string, { text: string; kind: Kind; placeholders: string[]; occurrences: string[] }> = {};

for (const msg of messages.values()) {
  setNested(zhTree, msg.ns, msg.key, msg.text);
  setNested(koTree, msg.ns, msg.key, "");
  sources[`${msg.ns}.${msg.key}`] = {
    text: msg.text,
    kind: msg.kind,
    placeholders: msg.placeholders,
    occurrences: msg.occurrences.sort(),
  };
}

const sortedSources: typeof sources = {};
for (const k of Object.keys(sources).sort()) sortedSources[k] = sources[k];

fs.mkdirSync(MESSAGES_DIR, { recursive: true });
const write = (name: string, data: unknown) =>
  fs.writeFileSync(path.join(MESSAGES_DIR, name), `${JSON.stringify(data, null, 2)}\n`, "utf8");

/** 读取现有消息文件. 如果丢失或损坏则为空树. */
function readTree(name: string): Tree {
  try {
    return JSON.parse(fs.readFileSync(path.join(MESSAGES_DIR, name), "utf8")) as Tree;
  } catch {
    return {};
  }
}

/**
 * 非破坏性合并. 提取创建一棵新树(fresh)的顶级命名空间是代码真相，所以
 * 按原样写, 现有文件(existing)顶级命名空间仅存在于(=手工翻译的策展
 * 名称空间)蜜饯. 提取的命名空间中的旧值(消失在上游) 身高自然
 * 掉出来(有利于上游对比).
 */
function mergeCurated(fresh: Tree, existing: Tree): Tree {
  const extractedTop = new Set(Object.keys(fresh));
  const out: Tree = { ...fresh };
  for (const [ns, subtree] of Object.entries(existing)) {
    if (!extractedTop.has(ns)) out[ns] = subtree;
  }
  return out;
}

write("zh.json", sortTree(mergeCurated(zhTree, readTree("zh.json"))));
write("ko.json", sortTree(mergeCurated(koTree, readTree("ko.json"))));
write("zh.sources.json", sortedSources);

// --- 总结报告 ------------------------------------------------------------
const byKind: Record<Kind, number> = { string: 0, template: 0, jsx: 0 };
const byNs = new Map<string, number>();
let totalOccurrences = 0;
for (const msg of messages.values()) {
  byKind[msg.kind]++;
  byNs.set(msg.ns, (byNs.get(msg.ns) ?? 0) + 1);
  totalOccurrences += msg.occurrences.length;
}

// 覆盖范围估计: 包含汉字的源行, 具有提取节点的线的百分比.
// 剩下的大部分都是评论(=不用于翻译)是.
let hanLines = 0;
let capturedHanLines = 0;
for (const absFile of scannedFiles) {
  const relFile = path.relative(REPO_WEB_DIR, absFile);
  const lines = fs.readFileSync(absFile, "utf8").split(/\r?\n/);
  const capSet = capturedLines.get(relFile) ?? new Set<number>();
  lines.forEach((ln, i) => {
    if (HAN.test(ln)) {
      hanLines++;
      if (capSet.has(i + 1)) capturedHanLines++;
    }
  });
}

const topNs = [...byNs.entries()].sort((a, b) => b[1] - a[1]).slice(0, 15);

console.log("─".repeat(64));
console.log("ARGUS i18n 提取完成");
console.log("─".repeat(64));
console.log(`扫描文件            : ${scannedFiles.length} (src/ 下面 .ts/.tsx, scripts 排除)`);
console.log(`包含汉字的文件         : ${fileWithHan}`);
console.log(`独特的信息(钥匙)        : ${messages.size}`);
console.log(`  ├─ 字符串文字      : ${byKind.string}`);
console.log(`  ├─ 模板文字      : ${byKind.template}`);
console.log(`  └─ JSX 文本         : ${byKind.jsx}`);
console.log(`总出现位置           : ${totalOccurrences} (包含重复使用)`);
console.log(`命名空间数量        : ${byNs.size}`);
console.log("");
console.log(`包含汉字的源行       : ${hanLines}`);
console.log(`  └─ 通过提取捕获的线   : ${capturedHanLines} (${((capturedHanLines / hanLines) * 100).toFixed(1)}%)`);
console.log(`     其余的 ${hanLines - capturedHanLines} 大多数行都是代码注释(不用于翻译)`);
console.log("");
console.log("包含许多消息的顶级命名空间 15:");
for (const [ns, n] of topNs) console.log(`  ${String(n).padStart(4)}  ${ns}`);
console.log("");
console.log(`输出: web/messages/{zh.json, ko.json, zh.sources.json}`);
console.log("─".repeat(64));
