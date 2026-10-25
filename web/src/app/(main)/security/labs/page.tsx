"use client";

import Link from "next/link";
import { ArrowRightIcon, BugIcon, FolderOpenIcon, WrenchIcon } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const labs = [
  { title: "CTF-Web 靶场", description: "进入 CTF Agent，绑定题目上下文并开始 Web 题分析。", href: "/chat?agent=ctf_agent", action: "开始 CTF 解题" },
  { title: "Reverse / Pwn 靶场", description: "在隔离实验环境中记录二进制、逆向和利用链分析。", href: "/chat?agent=ctf_agent", action: "打开分析会话" },
  { title: "SRC 授权靶场", description: "进入 Argus，设置 Target 与 Test Range 后开始授权测试。", href: "/chat?agent=src_agent", action: "开始 SRC 测试" },
  { title: "工具包与实验室", description: "直接打开工具中心，选择工具、查看参数并执行已授权操作。", href: "/system/tools", action: "打开工具中心" },
];

const onlineLabs = [
  ["picoCTF", "https://picoctf.org/"],
  ["PortSwigger Web Security Academy", "https://portswigger.net/web-security"],
  ["CTF+", "https://ctfplus.cn/"],
  ["CryptoHack", "https://cryptohack.org/"],
  ["Hack The Box", "https://www.hackthebox.com/"],
  ["TryHackMe", "https://tryhackme.com/"],
  ["青岑官方靶场", "https://www.qingcen.net/"],
  ["CTFShow", "https://ctf.show/"],
  ["看雪 CTF", "https://ctf.kanxue.com/"],
  ["好靶场", "https://www.hackthebox.cn/"],
  ["Bugku", "https://ctf.bugku.com/"],
  ["攻防世界", "https://adworld.xctf.org.cn/"],
  ["BUUCTF", "https://buuoj.cn/"],
  ["NSSCTF", "https://www.nssctf.cn/"],
  ["CTFHub", "https://www.ctfhub.com/"],
  ["VulnHub / VulFocus", "https://vulfocus.cn/"],
  ["封神台", "https://hack.zkaq.cn/"],
  ["玄机靶场", "https://xj.edisec.net/"],
  ["i春秋", "https://www.ichunqiu.com/"],
] as const;

export default function LabsPage() {
  return (
    <div className="space-y-6 p-6">
      <div className="space-y-2">
        <Badge variant="outline">安全研究</Badge>
        <h1 className="text-2xl font-semibold tracking-tight">本地靶场</h1>
        <p className="max-w-3xl text-sm text-muted-foreground">靶场和工具可以直接进入使用。每个入口都会保留当前会话、证据和运行记录。</p>
      </div>
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {labs.map((lab) => (
          <Card key={lab.title} className="flex flex-col">
            <CardHeader><CardTitle className="flex items-center gap-2 text-base"><BugIcon className="size-4" />{lab.title}</CardTitle></CardHeader>
            <CardContent className="flex flex-1 flex-col justify-between gap-4 text-sm text-muted-foreground">
              <p>{lab.description}</p>
              <Button asChild className="w-full"><Link href={lab.href}>{lab.action}<ArrowRightIcon /></Link></Button>
            </CardContent>
          </Card>
        ))}
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <Card><CardHeader><CardTitle className="flex items-center gap-2 text-base"><WrenchIcon className="size-4" />已有工具</CardTitle></CardHeader><CardContent className="flex items-center justify-between gap-4 text-sm text-muted-foreground"><span>工具中心支持参数填写、授权检查和结果回写。</span><Button asChild variant="outline"><Link href="/system/tools">查看工具</Link></Button></CardContent></Card>
        <Card><CardHeader><CardTitle className="flex items-center gap-2 text-base"><FolderOpenIcon className="size-4" />实验记录</CardTitle></CardHeader><CardContent className="flex items-center justify-between gap-4 text-sm text-muted-foreground"><span>任务、证据和报告会留在对应工作空间。</span><Button asChild variant="outline"><Link href="/function/workspace">打开工作空间</Link></Button></CardContent></Card>
      </div>
      <Card>
        <CardHeader><CardTitle>在线靶场</CardTitle><p className="text-sm text-muted-foreground">只打开官方平台，不自动上传内容或执行第三方操作。</p></CardHeader>
        <CardContent className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
          {onlineLabs.map(([name, href]) => <Button key={name} asChild variant="outline" className="justify-between"><a href={href} target="_blank" rel="noreferrer">{name}<ArrowRightIcon /></a></Button>)}
        </CardContent>
      </Card>
    </div>
  );
}
