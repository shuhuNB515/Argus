"use client";
import * as React from "react";
import Link from "next/link";
import { ArrowLeftIcon, PackageIcon, UploadIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";

type Plugin = { name: string; version?: string; description?: string; tools?: string[]; enabled?: boolean };
const KEY = "argus.plugins";
export default function PluginsPage() {
  const [plugins, setPlugins] = React.useState<Plugin[]>([]);
  const input = React.useRef<HTMLInputElement>(null);
  React.useEffect(() => { try { setPlugins(JSON.parse(localStorage.getItem(KEY) ?? "[]")); } catch { setPlugins([]); } }, []);
  function persist(next: Plugin[]) { setPlugins(next); localStorage.setItem(KEY, JSON.stringify(next)); }
  function importPlugin(file?: File) {
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => { try { const p = JSON.parse(String(reader.result)) as Plugin; if (!p.name) throw new Error("缺少 name"); persist([...plugins.filter((x) => x.name !== p.name), { ...p, enabled: true }]); toast.success(`已导入插件：${p.name}`); } catch (e) { toast.error(`插件清单无效：${(e as Error).message}`); } };
    reader.readAsText(file);
  }
  return <div className="flex flex-1 flex-col gap-5 p-1">
    <div className="flex items-start justify-between gap-3"><div><div className="flex items-center gap-2"><Button asChild size="sm" variant="ghost"><Link href="/system/mcp"><ArrowLeftIcon />MCP</Link></Button><Badge variant="secondary">插件</Badge></div><h1 className="mt-3 text-xl font-semibold">插件管理</h1><p className="text-muted-foreground text-sm">导入插件清单并管理启用状态；插件运行时由本地 ARTEX/AOCI 环境加载。</p></div><><input ref={input} type="file" accept=".json,application/json" className="hidden" onChange={(e) => importPlugin(e.target.files?.[0])} /><Button onClick={() => input.current?.click()}><UploadIcon />一键导入</Button></></div>
    {plugins.length === 0 ? <Card><CardContent className="flex flex-col items-center gap-2 py-12 text-center text-muted-foreground"><PackageIcon className="size-8" /><p>暂无已导入插件</p><p className="text-xs">选择插件 JSON 清单即可导入，不会自动上传文件。</p></CardContent></Card> : <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">{plugins.map((p) => <Card key={p.name}><CardHeader><CardTitle className="flex items-center justify-between gap-2">{p.name}<Badge variant={p.enabled ? "default" : "outline"}>{p.enabled ? "已启用" : "已停用"}</Badge></CardTitle></CardHeader><CardContent className="space-y-3"><p className="text-muted-foreground text-sm">{p.description || "未提供描述"}</p><p className="text-muted-foreground text-xs">版本 {p.version || "未标注"} · 工具 {p.tools?.length ?? 0}</p><div className="flex gap-2"><Button size="sm" variant="outline" onClick={() => persist(plugins.map((x) => x.name === p.name ? { ...x, enabled: !x.enabled } : x))}>{p.enabled ? "停用" : "启用"}</Button><Button size="sm" variant="ghost" onClick={() => persist(plugins.filter((x) => x.name !== p.name))}>移除</Button></div></CardContent></Card>)}</div>}
  </div>;
}
