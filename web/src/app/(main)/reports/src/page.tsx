"use client";
import Link from "next/link";

import { ArrowLeftIcon } from "lucide-react";

import { ReportList } from "@/app/(main)/reports/_components/report-list";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

export default function SrcReportsPage() {
  return (
    <div className="space-y-6 p-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <Badge variant="outline">安全研究</Badge>
          <h1 className="mt-2 text-2xl font-semibold">SRC 测试报告</h1>
          <p className="text-muted-foreground mt-1 text-sm">报告只属于当前工作区和授权目标，SRC 报告不可合并。</p>
        </div>
        <Button variant="outline" asChild>
          <Link href="/reports">
            <ArrowLeftIcon />
            返回报告中心
          </Link>
        </Button>
      </div>
      <ReportList kind="src" />
    </div>
  );
}
