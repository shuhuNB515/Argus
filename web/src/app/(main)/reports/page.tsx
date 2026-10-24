"use client";

import Link from "next/link";

import { FileTextIcon, ShieldCheckIcon, TrophyIcon } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function ReportsPage() {
  return (
    <div className="space-y-6 p-6">
      <div>
        <Badge variant="outline">报告与交付</Badge>
        <h1 className="mt-2 text-2xl font-semibold tracking-tight">报告中心</h1>
        <p className="text-muted-foreground mt-1 text-sm">SRC 测试报告与 CTF 解题报告分开管理，数据来自当前工作区。</p>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <ShieldCheckIcon className="size-5" />
              SRC 测试报告
            </CardTitle>
            <CardDescription>按目标、测试范围和证据整理漏洞报告。</CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild>
              <Link href="/reports/src">查看 SRC 报告</Link>
            </Button>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrophyIcon className="size-5" />
              CTF 解题报告
            </CardTitle>
            <CardDescription>每道题生成一份 WP，可按比赛合并整理。</CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild>
              <Link href="/reports/ctf">查看 CTF 报告</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileTextIcon className="size-5" />
            当前工作区
          </CardTitle>
          <CardDescription>尚未导入演示报告。完成真实测试或解题后，报告会保存到所选工作区。</CardDescription>
        </CardHeader>
      </Card>
    </div>
  );
}
