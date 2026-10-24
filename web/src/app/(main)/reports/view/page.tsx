"use client";
import * as React from "react";

import Link from "next/link";

import { ArrowLeftIcon, CopyIcon, DownloadIcon, FileTextIcon } from "lucide-react";
import { toast } from "sonner";

import { Markdown } from "@/components/markdown";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { readReports, type WorkspaceReport } from "@/lib/report-store";
import { copyText } from "@/lib/utils";

export default function ReportViewPage() {
  const [report, setReport] = React.useState<WorkspaceReport | null>(null);
  React.useEffect(() => {
    const id = new URLSearchParams(window.location.search).get("id");
    setReport(readReports().find((item) => item.id === id) ?? null);
  }, []);
  async function copy() {
    if (report && (await copyText(report.markdown))) toast.success("已复制报告内容");
  }
  function download() {
    if (!report) return;
    const url = URL.createObjectURL(new Blob([report.markdown], { type: "text/markdown;charset=utf-8" }));
    const a = document.createElement("a");
    a.href = url;
    a.download = `${report.title.replace(/[\\/:*?"<>|]/g, "-")}.md`;
    a.click();
    URL.revokeObjectURL(url);
  }
  const back = report?.kind === "ctf" ? "/reports/ctf" : "/reports/src";
  return (
    <div className="space-y-6 p-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <Badge variant="outline">{report?.kind === "ctf" ? "CTF 解题报告" : "SRC 测试报告"}</Badge>
          <h1 className="mt-2 text-2xl font-semibold">{report?.title ?? "报告不存在"}</h1>
          {report && (
            <p className="text-muted-foreground mt-1 text-sm">
              {report.taskName} · {new Date(report.updatedAt).toLocaleString("zh-CN")}
            </p>
          )}
        </div>
        <div className="flex gap-2">
          <Button variant="outline" asChild>
            <Link href={back}>
              <ArrowLeftIcon />
              返回报告列表
            </Link>
          </Button>
          {report && (
            <>
              <Button variant="outline" onClick={() => void copy()}>
                <CopyIcon />
                复制
              </Button>
              <Button onClick={download}>
                <DownloadIcon />
                下载 Markdown
              </Button>
            </>
          )}
        </div>
      </div>
      {report ? (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileTextIcon className="size-5" />
              报告正文
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="rounded-lg border bg-muted/10 p-5">
              <Markdown text={report.markdown} />
            </div>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="py-16 text-center text-muted-foreground">
            没有找到这份报告，它可能已被删除或属于其他工作区。
          </CardContent>
        </Card>
      )}
    </div>
  );
}
