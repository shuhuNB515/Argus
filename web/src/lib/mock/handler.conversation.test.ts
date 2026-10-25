import { beforeEach, describe, expect, it, vi } from "vitest";

describe("mock conversation messages", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.resetModules();
  });

  it("starts a new conversation without loading sample history", async () => {
    const { mockHandle } = await import("./handler");
    const conversation = await mockHandle<{ id: number }>(
      "POST",
      "/conversations",
      JSON.stringify({ agent_key: "ctf-agent", title: "青岑 CTF" }),
    );

    const history = await mockHandle<{ items: unknown[] }>("GET", `/conversations/${conversation.id}/messages?limit=200`);
    expect(history.items).toEqual([]);
  });

  it("keeps the exact user message and responds in Chinese without fake execution results", async () => {
    const { mockHandle } = await import("./handler");
    const conversation = await mockHandle<{ id: number }>(
      "POST",
      "/conversations",
      JSON.stringify({ agent_key: "ctf-agent", title: "青岑 CTF" }),
    );
    const message = "docker.qingcen.net:49830，解出flag";

    await mockHandle("POST", `/conversations/${conversation.id}/messages`, JSON.stringify({ message }));
    const history = await mockHandle<{ items: Array<{ kind: string; summary: string }> }>(
      "GET",
      `/conversations/${conversation.id}/messages?limit=200`,
    );

    expect(history.items[0]).toMatchObject({ kind: "user", summary: message });
    expect(history.items[1].kind).toBe("text");
    expect(history.items[1].summary).toContain("当前本地预览未连接真实 Agent 运行时");
    expect(history.items[1].summary).not.toMatch(/admin\.acme|admin123|It worked|flag\s*[:：]/i);

    vi.resetModules();
    const { mockHandle: reloadedHandle } = await import("./handler");
    const reloaded = await reloadedHandle<{ items: Array<{ summary: string }> }>(
      "GET",
      `/conversations/${conversation.id}/messages?limit=200`,
    );
    expect(reloaded.items[0].summary).toBe(message);
  });
});
