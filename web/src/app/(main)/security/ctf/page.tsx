import { StrategyPage } from "@/app/(main)/_components/strategy-page";
export default function CtfPage() { return <StrategyPage eyebrow="安全研究" title="CTF 解题" description="绑定当前题目目标后运行 Web、Reverse、Pwn、Crypto 与 Forensics 子 Agent。通过验证的解题思路会先进入 Skill 草稿，审核后共享给 Argus。" cards={["当前题目与上下文", "解题任务图", "子 Agent 协作", "解法沉淀为 Skill", "共享给 Argus"]} />; }
