import { StrategyPage } from "@/app/(main)/_components/strategy-page";
export default function SrcPage() { return <StrategyPage eyebrow="安全研究" title="SRC 测试" description="在明确 Target 与 Test Range 后运行侦察、资产分析、验证、证据和报告流程。超出范围的调用会在 ARTEX 执行前阻止，并优先参考已审核的 CTF Skill。" cards={["目标与测试范围", "自动化侦察", "验证与影响评估", "CTF Skill 候选", "证据与报告"]} />; }
