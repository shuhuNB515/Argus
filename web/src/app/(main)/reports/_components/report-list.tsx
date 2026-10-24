"use client";

import * as React from "react";

import Link from "next/link";

import { CheckSquareIcon, FileTextIcon, SearchIcon } from "lucide-react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { mergeCtfReports, type ReportKind, readReports, type WorkspaceReport } from "@/lib/report-store";
import { saveTaskReport } from "@/lib/report-store";
import { api } from "@/lib/api";

const severityLabels = { critical: "严重", high: "高危", medium: "中危", low: "低危", info: "信息" } as const;
const severityOrder = { critical: 5, high: 4, medium: 3, low: 2, info: 1 } as const;
const categoryFilters = ["all", "Web", "密码学", "逆向", "Pwn", "杂项", "取证", "移动安全", "其它"];
const severityFilters = ["all", "critical", "high", "medium", "low", "info"];

function useWorkspaceReports(kind: ReportKind) {
  const [reports, setReports] = React.useState<WorkspaceReport[]>([]);
  React.useEffect(() => {
    const refresh = () => setReports(readReports().filter((item) => item.kind === kind));
    refresh();
	let disposed = false;
	let loading = false;
	const sync = async () => {
	  if (loading || kind !== "src") return;
	  loading = true;
	  try {
	    const [{ tasks }, conversations] = await Promise.all([api.tasks(), api.conversations()]);
	    const linked = new Set(conversations.filter((c) => c.agent_key === "src_agent" && c.task_id).map((c) => String(c.task_id)));
	    for (const task of tasks) {
	      if (disposed || !linked.has(task.id) || task.status !== "done") continue;
	      const cached = readReports().find((r) => r.taskId === task.id && r.kind === "src");
	      if (cached && task.completed_at && Date.parse(cached.updatedAt) >= Date.parse(task.completed_at)) continue;
	      const markdown = await api.report(task.id);
	      if (!disposed) saveTaskReport(task, markdown);
	    }
	  } catch { /* Retain existing reports on temporary connectivity failures. */ }
	  finally { loading = false; }
	};
	void sync();
	const timer = window.setInterval(() => void sync(), 5000);
    window.addEventListener("argus:reports-updated", refresh);
    window.addEventListener("storage", refresh);
    return () => {
	  disposed = true;
	  window.clearInterval(timer);
      window.removeEventListener("argus:reports-updated", refresh);
      window.removeEventListener("storage", refresh);
    };
  }, [kind]);
  return reports;
}

export function ReportList({ kind }: { kind: ReportKind }) {
  const reports = useWorkspaceReports(kind);
  const [query, setQuery] = React.useState("");
  const [filter, setFilter] = React.useState("all");
  const [sort, setSort] = React.useState<"time" | "severity">("time");
  const [selected, setSelected] = React.useState<string[]>([]);
  const choices = kind === "ctf" ? categoryFilters : severityFilters;
  const filterLabel = (value: string) => {
    if (value === "all") return "全部";
    if (kind === "src") return severityLabels[value as keyof typeof severityLabels];
    return value;
  };
  const filtered = reports
    .filter((item) => {
      const matchesQuery = `${item.title} ${item.taskName} ${item.target} ${item.competition}`
        .toLowerCase()
        .includes(query.toLowerCase());
      const matchesFilter = filter === "all" || (kind === "ctf" ? item.category === filter : item.severity === filter);
      return matchesQuery && matchesFilter;
    })
    .sort((a, b) => {
      if (kind === "src" && sort === "severity") {
        const severityDelta = severityOrder[b.severity] - severityOrder[a.severity];
        if (severityDelta !== 0) return severityDelta;
      }
      return b.updatedAt.localeCompare(a.updatedAt);
    });
  const toggle = (id: string) =>
    setSelected((current) => (current.includes(id) ? current.filter((item) => item !== id) : [...current, id]));
  function merge() {
    try {
      const report = mergeCtfReports(selected);
      if (!report) return;
      setSelected([]);
      toast.success(`已生成：${report.title}`);
    } catch (e) {
      toast.error((e as Error).message);
    }
  }
  const stats =
    kind === "src"
      ? [
          { label: "报告总数", value: reports.length },
          ...(["critical", "high", "medium", "low"] as const).map((value) => ({
            label: severityLabels[value],
            value: reports.filter((item) => item.severity === value).length,
          })),
        ]
      : [
          { label: "报告总数", value: reports.length },
          ...(["Web", "密码学", "逆向", "Pwn"] as const).map((value) => ({
            label: value,
            value: reports.filter((item) => item.category === value).length,
          })),
        ];
  return (
    <div className="space-y-5">
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
        {stats.map((item) => (
          <Card key={item.label}>
            <CardContent className="py-5">
              <p className="text-muted-foreground text-sm">{item.label}</p>
              <p className="mt-2 text-3xl font-semibold tabular-nums">{item.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>
      <div className="flex flex-wrap items-center gap-3">
        <div className="relative min-w-64 flex-1">
          <SearchIcon className="text-muted-foreground absolute left-3 top-1/2 size-4 -translate-y-1/2" />
          <Input
            className="pl-9"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={kind === "ctf" ? "搜索题目、比赛或任务" : "搜索报告、目标或任务"}
          />
        </div>
        <Select value={filter} onValueChange={setFilter}>
          <SelectTrigger className="w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {choices.map((value) => (
              <SelectItem key={value} value={value}>
                {filterLabel(value)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={sort} onValueChange={(value) => setSort(value as "time" | "severity")}>
          <SelectTrigger className="w-36">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="time">按时间排序</SelectItem>
            {kind === "src" && <SelectItem value="severity">按严重程度</SelectItem>}
          </SelectContent>
        </Select>
        {kind === "ctf" && (
          <Button onClick={merge} disabled={selected.length < 2}>
            <CheckSquareIcon />
            合并报告（{selected.length}）
          </Button>
        )}
      </div>
      <div className="overflow-hidden rounded-xl border">
        <Table>
          <TableHeader>
            <TableRow>
              {kind === "ctf" && <TableHead className="w-12" />}
              <TableHead>{kind === "ctf" ? "方向" : "严重度"}</TableHead>
              <TableHead>报告名称</TableHead>
              <TableHead>{kind === "ctf" ? "比赛" : "目标"}</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>所属任务</TableHead>
              <TableHead>时间</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((item) => (
              <TableRow key={item.id}>
                {kind === "ctf" && (
                  <TableCell>
                    <Checkbox checked={selected.includes(item.id)} onCheckedChange={() => toggle(item.id)} />
                  </TableCell>
                )}
                <TableCell>
                  <Badge variant="outline">{kind === "ctf" ? item.category : severityLabels[item.severity]}</Badge>
                </TableCell>
                <TableCell className="max-w-72 truncate font-medium">{item.title}</TableCell>
                <TableCell className="max-w-52 truncate">{kind === "ctf" ? item.competition : item.target}</TableCell>
                <TableCell>
                  <Badge variant="secondary">{item.status}</Badge>
                </TableCell>
                <TableCell className="max-w-40 truncate">{item.taskName}</TableCell>
                <TableCell className="whitespace-nowrap text-muted-foreground">
                  {new Date(item.updatedAt).toLocaleString("zh-CN")}
                </TableCell>
                <TableCell className="text-right">
                  <Button size="sm" variant="ghost" asChild>
                    <Link href={`/reports/view?id=${encodeURIComponent(item.id)}`}>查看</Link>
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={8} className="h-52 text-center">
                  <div className="flex flex-col items-center gap-2 text-muted-foreground">
                    <FileTextIcon className="size-8 opacity-50" />
                    <span>{reports.length ? "没有符合筛选条件的报告" : "暂无真实报告"}</span>
                    <span className="text-xs">
                      {kind === "ctf" ? "使用 CTF Agent 验证 flag 后，会自动生成独立解题报告。" : "完成授权 SRC 测试并生成报告后会自动出现在这里。"}
                    </span>
                  </div>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
