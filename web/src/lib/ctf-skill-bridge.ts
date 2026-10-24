/**
 * Shared knowledge seam between the CTF and SRC agents.
 *
 * CTF solutions enter this list as drafts. After review, an approved skill is
 * available to SRC routing as a candidate check, while the SRC authorization
 * policy still decides whether it can run against a target.
 */
export type SharedSkillStatus = "draft" | "approved";

export type SharedCtfSkill = {
  id: string;
  name: string;
  category: "Web" | "Reverse" | "Pwn" | "Crypto" | "Forensics";
  status: SharedSkillStatus;
  sourceChallenge: string;
  summary: string;
};

export const SHARED_CTF_SKILLS: SharedCtfSkill[] = [
  {
    id: "ctf-web-auth-bypass",
    name: "Web 授权绕过验证链",
    category: "Web",
    status: "approved",
    sourceChallenge: "Acme Web · session-lab",
    summary: "从身份边界、对象权限和证据回放中提取可复用的验证步骤。",
  },
  {
    id: "ctf-reverse-config-trace",
    name: "Reverse 配置路径追踪",
    category: "Reverse",
    status: "approved",
    sourceChallenge: "Binary Lab · config-trace",
    summary: "将字符串、调用图和输入约束整理为低风险的分析候选项。",
  },
  {
    id: "ctf-pending-example",
    name: "待审核：JWT 密钥线索",
    category: "Web",
    status: "draft",
    sourceChallenge: "CTF 练习题 · jwt-01",
    summary: "保留失败尝试和待补证据，审核前不会进入 SRC Agent。",
  },
];

export function approvedSkillsForSrcAgent() {
  return SHARED_CTF_SKILLS.filter((skill) => skill.status === "approved");
}
