import { NextResponse } from "next/server";

export function proxy() {
  // Auth mode is decided by the backend (loopback desktop installs default to
  // no login). Keep the Next layer transparent so it cannot redirect before the
  // client has learned the backend's mode from /api/auth/status.
  return NextResponse.next();
}

export const config = {
  // 跳过 Next.js 内部路由、API 路由、favicon 及 public/ 下的静态文件（含图片、字体等）
  matcher: [
    "/((?!_next/static|_next/image|favicon\\.ico|api/|.*\\.(?:png|jpg|jpeg|gif|webp|svg|ico|woff2?|ttf|otf)$).*)",
  ],
};
