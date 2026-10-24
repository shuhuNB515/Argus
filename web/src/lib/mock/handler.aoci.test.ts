import { beforeEach, describe, expect, it, vi } from "vitest";

describe("mock AOCI workspace flow", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.resetModules();
  });

  it("moves a workspace from missing to aligned and persists it", async () => {
    let { mockHandle } = await import("./handler");
    const before = await mockHandle<{ state: string }>("GET", "/aoci/status?path=projects/demo");
    expect(before.state).toBe("missing");

    const built = await mockHandle<{ state: string; root: string; files: number }>(
      "POST",
      "/aoci/index",
      JSON.stringify({ path: "projects/demo", reload_mcp: true }),
    );
    expect(built).toMatchObject({ state: "aligned", root: "projects/demo" });
    expect(built.files).toBeGreaterThan(0);

    vi.resetModules();
    ({ mockHandle } = await import("./handler"));
    const after = await mockHandle<{ state: string }>("GET", "/aoci/status?path=projects/demo");
    expect(after.state).toBe("aligned");
  });
});
