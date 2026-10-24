"use client";
import Link from "next/link";

import { ArrowLeftIcon } from "lucide-react";

import { ReportList } from "@/app/(main)/reports/_components/report-list";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

export default function CtfReportsPage() {
  return (
    <div className="space-y-6 p-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <Badge variant="outline">安全研究</Badge>
          <h1 className="mt-2 text-2xl font-semibold">CTF 解题报告</h1>
          <p className="text-muted-foreground mt-1 text-sm">每道题独立生成 WP，可选择同一比赛的题目后合并整理。</p>
        </div>
        <Button variant="outline" asChild>
          <Link href="/reports">
            <ArrowLeftIcon />
            返回报告中心
          </Link>
        </Button>
      </div>
      <ReportList kind="ctf" />
    </div>
  );
}
