import { StrategyPage } from "@/app/(main)/_components/strategy-page";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { approvedSkillsForSrcAgent, SHARED_CTF_SKILLS } from "@/lib/ctf-skill-bridge";

export default function CtfToSkillPage() {
  const approved = approvedSkillsForSrcAgent();
  return (
    <div>
      <StrategyPage eyebrow="资源与运行时" title="CTF 解题沉淀为 Skill" description={`从已完成的 CTF 任务中提取目标、前置条件、步骤、验证证据和失败经验，生成草稿 Skill，审核后加入 Skill 库。当前 ${approved.length} 个已审核 Skill 可供 ARTEX 作为候选检查项，草稿仍需审核。`} cards={["选择已完成题目", "提取解法与证据", "生成 Skill 草稿", `审核、指纹化与发布（${SHARED_CTF_SKILLS.length} 条记录）`, "共享给 ARTEX"]} />
      <div className="grid gap-4 px-6 pb-6 md:grid-cols-2 xl:grid-cols-3">
        {SHARED_CTF_SKILLS.map((skill) => (
          <Card key={skill.id}>
            <CardHeader className="space-y-2">
              <div className="flex items-center justify-between gap-2">
                <CardTitle className="text-base">{skill.name}</CardTitle>
                <Badge variant={skill.status === "approved" ? "default" : "outline"}>{skill.status === "approved" ? "已审核" : "待审核"}</Badge>
              </div>
              <p className="text-xs text-muted-foreground">{skill.category} · {skill.sourceChallenge}</p>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">{skill.summary}</CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
