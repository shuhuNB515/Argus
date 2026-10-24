import { StrategyPage } from "@/app/(main)/_components/strategy-page";
export default function SkillInspectorPage() { return <StrategyPage eyebrow="资源与运行时" title="Skill 检查器" description="查看 Skill 的描述、指纹、候选队列、失败切换和调用记录，保持 Skill Runtime 可解释、可回滚。" cards={["Skill 描述与指纹", "路由候选队列", "循环与失败检测", "版本与回滚"]} />; }
