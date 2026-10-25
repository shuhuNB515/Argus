"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export function StrategyPage({ title, eyebrow, description, cards }: { title: string; eyebrow: string; description: string; cards: string[] }) {
  return (
    <div className="space-y-6 p-6">
      <div className="space-y-2">
        <Badge variant="outline">{eyebrow}</Badge>
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        <p className="max-w-3xl text-sm text-muted-foreground">{description}</p>
      </div>
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {cards.map((card) => (
          <Card key={card}>
            <CardHeader><CardTitle className="text-base">{card}</CardTitle></CardHeader>
            <CardContent className="text-sm text-muted-foreground">已接入 Argus 任务图、证据链与运行事件，等待配置后使用。</CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
