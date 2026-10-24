export function SrcAgentMark({ size = 32 }: { size?: number }) {
  return (
    <span
      aria-hidden="true"
      className="inline-flex shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-violet-600 via-indigo-600 to-cyan-500 text-white shadow-sm"
      style={{ width: size, height: size }}
    >
      <span className="font-semibold leading-none" style={{ fontSize: Math.max(16, Math.round(size * 0.56)) }}>✦</span>
    </span>
  );
}
