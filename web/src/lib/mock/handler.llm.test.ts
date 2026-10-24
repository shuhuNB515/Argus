import { beforeEach, describe, expect, it, vi } from "vitest";

type Profile = { id: string; name: string; is_default: boolean };

async function handler() {
  return import("./handler");
}

async function profiles() {
  const { mockHandle } = await handler();
  return mockHandle<{ profiles: Profile[] }>("GET", "/llm/profiles");
}

describe("mock LLM profile persistence", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.resetModules();
  });

  it("activates a profile without mutating an earlier response", async () => {
    const first = await profiles();
    const { mockHandle } = await handler();

    await mockHandle("POST", "/llm/profiles/active", JSON.stringify({ id: 2 }));
    const second = await profiles();

    expect(first.profiles.map((profile) => profile.is_default)).toEqual([true, false]);
    expect(second.profiles.map((profile) => profile.is_default)).toEqual([false, true]);
  });

  it("keeps deleted profiles removed after the mock module reloads", async () => {
    const { mockHandle } = await handler();
    await mockHandle("DELETE", "/llm/profiles/2");

    vi.resetModules();
    const afterReload = await profiles();

    expect(afterReload.profiles.map((profile) => profile.name)).toEqual(["Claude Opus 4.8"]);
  });

  it("reports the activated profile without invented breaker failures", async () => {
    const { mockHandle } = await handler();
    await mockHandle("POST", "/llm/profiles/active", JSON.stringify({ id: 2 }));

    const pool = await mockHandle<{ chain: Array<{ profile_id: string; active: boolean; state: string; cooldown_secs: number; last_error?: string }> }>("GET", "/llm/pool");

    expect(pool.chain.find((member) => member.profile_id === "2")).toMatchObject({
      active: true,
      state: "ok",
      cooldown_secs: 0,
    });
    expect(pool.chain.some((member) => member.last_error)).toBe(false);
  });

  it("uses the configured host nmap path in the local preview", async () => {
    vi.stubEnv("NEXT_PUBLIC_LOCAL_NMAP_PATH", "/opt/homebrew/bin/nmap");
    const { mockHandle } = await handler();

    await expect(mockHandle("POST", "/settings/nmap/detect")).resolves.toEqual({
      available: true,
      path: "/opt/homebrew/bin/nmap",
      source: "local",
    });
    vi.unstubAllEnvs();
  });
});
