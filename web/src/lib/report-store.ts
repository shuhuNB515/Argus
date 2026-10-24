import type { Task } from "@/lib/types";

export type ReportKind = "src" | "ctf";
export type ReportSeverity = "critical" | "high" | "medium" | "low" | "info";
export type CtfCategory = "Web" | "密码学" | "逆向" | "Pwn" | "杂项" | "取证" | "移动安全" | "其它";

export interface WorkspaceReport {
  id: string;
  kind: ReportKind;
  title: string;
  taskId: string;
  taskName: string;
  target: string;
  competition: string;
  severity: ReportSeverity;
  category: CtfCategory;
  status: "已生成" | "草稿";
  markdown: string;
  createdAt: string;
  updatedAt: string;
  mergedFrom?: string[];
}

const KEY = "argus.workspace.reports.v1";

export function readReports(): WorkspaceReport[] {
  if (typeof window === "undefined") return [];
  try {
    const value = JSON.parse(localStorage.getItem(KEY) ?? "[]");
    return Array.isArray(value) ? value : [];
  } catch {
    return [];
  }
}

function writeReports(reports: WorkspaceReport[]) {
  localStorage.setItem(KEY, JSON.stringify(reports));
  window.dispatchEvent(new CustomEvent("argus:reports-updated"));
}

function categoryOf(text: string): CtfCategory {
  const value = text.toLowerCase();
  if (/web|xss|sql|ssrf|http/.test(value)) return "Web";
  if (/crypto|密码|rsa|aes|hash/.test(value)) return "密码学";
  if (/reverse|逆向|re\b/.test(value)) return "逆向";
  if (/pwn|栈|堆|rop/.test(value)) return "Pwn";
  if (/forensic|取证|流量/.test(value)) return "取证";
  if (/mobile|android|ios|移动/.test(value)) return "移动安全";
  if (/misc|杂项/.test(value)) return "杂项";
  return "其它";
}

function severityOf(task: Task): ReportSeverity {
  if (task.findings?.critical) return "critical";
  if (task.findings?.high) return "high";
  if (task.findings?.medium) return "medium";
  if (task.findings?.low) return "low";
  return "info";
}

export function saveTaskReport(task: Task, markdown: string) {
  if (!markdown.trim()) return;
  const descriptor = `${task.name ?? ""} ${task.category_name ?? ""} ${task.description}`;
  const kind: ReportKind = /ctf|题目|比赛/i.test(descriptor) ? "ctf" : "src";
  const now = new Date().toISOString();
  const reports = readReports();
  const id = `${kind}:${task.id}`;
  const existing = reports.find((item) => item.id === id);
  const report: WorkspaceReport = {
    id,
    kind,
    title: task.name?.trim() || task.description.trim() || `任务 ${task.id}`,
    taskId: task.id,
    taskName: task.name?.trim() || `任务 ${task.id}`,
    target: task.goal.trim() || "当前工作区",
    competition: task.category_name?.trim() || "未分组",
    severity: severityOf(task),
    category: categoryOf(descriptor),
    status: kind === "src" ? "草稿" : "已生成",
    markdown,
    createdAt: existing?.createdAt ?? now,
    updatedAt: now,
  };
  writeReports([report, ...reports.filter((item) => item.id !== id)]);
}

/**
 * Persist a CTF write-up independently of task execution. The CTF chat emits
 * the solved marker only after it has verified a flag, so solving a challenge
 * does not need to create a platform task first.
 */
export function saveCtfSolveReport(input: {
  conversationId: string;
  sequence: number;
  title?: string;
  competition?: string;
  category?: string;
  target?: string;
  flag?: string;
  markdown: string;
}) {
  if (!input.markdown.trim()) return;
  const now = new Date().toISOString();
  const reports = readReports();
  const id = `ctf:conversation:${input.conversationId}:${input.sequence}`;
  const existing = reports.find((item) => item.id === id);
  const descriptor = `${input.title ?? ""} ${input.category ?? ""} ${input.markdown}`;
  const category = (input.category && ["Web", "密码学", "逆向", "Pwn", "杂项", "取证", "移动安全", "其它"].includes(input.category)
    ? input.category
    : categoryOf(descriptor)) as CtfCategory;
  const title = input.title?.trim() || `CTF 题目 ${input.sequence}`;
  const report: WorkspaceReport = {
    id,
    kind: "ctf",
    title,
    taskId: input.conversationId,
    taskName: title,
    target: input.target?.trim() || input.flag?.trim() || "当前工作区",
    competition: input.competition?.trim() || "未分组",
    severity: "info",
    category,
    status: "已生成",
    markdown: input.flag?.trim() ? `${input.markdown.trim()}\n\n> Flag：\`${input.flag.trim()}\`` : input.markdown.trim(),
    createdAt: existing?.createdAt ?? now,
    updatedAt: now,
  };
  writeReports([report, ...reports.filter((item) => item.id !== id)]);
}

export function mergeCtfReports(ids: string[]): WorkspaceReport | null {
  const reports = readReports().filter((item) => ids.includes(item.id) && item.kind === "ctf");
  if (reports.length < 2) return null;
  const competition = reports[0].competition;
  if (reports.some((item) => item.competition !== competition)) throw new Error("只能合并同一比赛的解题报告");
  const now = new Date().toISOString();
  const merged: WorkspaceReport = {
    id: `ctf:merged:${Date.now()}`,
    kind: "ctf",
    title: `${competition} 合并 WP`,
    taskId: "",
    taskName: `${reports.length} 道题`,
    target: "当前工作区",
    competition,
    severity: "info",
    category: "其它",
    status: "已生成",
    markdown: `# ${competition} Writeup\n\n${reports.map((item) => `## ${item.title}\n\n${item.markdown}`).join("\n\n---\n\n")}`,
    createdAt: now,
    updatedAt: now,
    mergedFrom: reports.map((item) => item.id),
  };
  writeReports([merged, ...readReports()]);
  return merged;
}
