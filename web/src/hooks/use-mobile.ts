import * as React from "react";

const MOBILE_BREAKPOINT = 768;
// 当侧边栏展开并且文本宽度变窄时，表格·卡片重叠(768~1279px 部分).
// 低于这个宽度，侧边栏会自动折叠成图标模式，以保证主体的宽度。.
export const TABLET_BREAKPOINT = 1280;

function useBreakpoint(breakpoint: number) {
  const [below, setBelow] = React.useState<boolean | undefined>(undefined);

  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${breakpoint - 1}px)`);
    const onChange = () => setBelow(window.innerWidth < breakpoint);
    mql.addEventListener("change", onChange);
    setBelow(window.innerWidth < breakpoint);
    return () => mql.removeEventListener("change", onChange);
  }, [breakpoint]);

  return below;
}

export function useIsMobile() {
  return !!useBreakpoint(MOBILE_BREAKPOINT);
}

// 药片宽度(768~1279px) 是否. 初始渲染(测量前)在 undefined 按原样归还,
// 寻呼机 "还不知道"和 "false"让您能够区分.
export function useIsTablet() {
  return useBreakpoint(TABLET_BREAKPOINT);
}
