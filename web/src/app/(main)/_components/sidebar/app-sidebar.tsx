"use client";

import * as React from "react";
import Link from "next/link";
import { MoreHorizontal } from "lucide-react";
import { usePathname } from "next/navigation";

import { useShallow } from "zustand/react/shallow";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { APP_CONFIG } from "@/config/app-config";
import { useCurrentUser } from "@/hooks/use-current-user";
import { sidebarItems } from "@/navigation/sidebar/sidebar-items";
import type { NavGroup } from "@/navigation/sidebar/sidebar-items";
import { usePreferencesStore } from "@/stores/preferences/preferences-provider";

import { NavMain } from "./nav-main";
import { NavUser } from "./nav-user";
import { SidebarSupportCard } from "./sidebar-support-card";
import { SrcAgentMark } from "@/components/src-agent-mark";

const railItems = [
  ["⌂", "聊天", "/chat"],
  ["◇", "工作区", "/function/workspace"],
  ["✓", "任务", "/function/tasks"],
  ["◈", "安全", "/security/src"],
  ["▦", "资源", "/system/tools"],
  ["⚙", "设置", "/system/settings"],
] as const;

function SrcAgentRail({ featureItems }: { featureItems?: readonly NavGroup[] }) {
  const [featureOpen, setFeatureOpen] = React.useState(false);
  return (
    <aside className="relative flex w-16 shrink-0 flex-col items-center border-r bg-muted/20 py-3" aria-label="Argus 主导航">
      <Link href="/function/tasks" aria-label="Argus 首页" className="mb-5 flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground text-sm font-semibold shadow-sm">
        <SrcAgentMark size={36} />
      </Link>
      <nav className="flex flex-1 flex-col items-center gap-2">
        {railItems.map(([icon, label, href]) => (
          <Link key={label} href={href} title={label} aria-label={label} className="flex size-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
            <span aria-hidden="true" className="text-base leading-none">{icon}</span>
          </Link>
        ))}
      </nav>
      {featureItems && (
        <button
          type="button"
          onClick={() => setFeatureOpen((value) => !value)}
          className="flex size-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          aria-label={featureOpen ? "关闭功能区" : "打开功能区"}
          aria-expanded={featureOpen}
        >
          <MoreHorizontal className="size-5" />
        </button>
      )}
      {featureItems && featureOpen && (
        <div className="absolute top-2 left-16 z-50 flex h-[calc(100svh-1rem)] w-72 flex-col overflow-hidden rounded-r-xl border-y border-r bg-sidebar text-sidebar-foreground shadow-xl">
          <SidebarHeader>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild>
                  <Link prefetch={false} href="/chat">
                    <SrcAgentMark size={32} />
                    <span className="font-semibold text-base">Argus 功能区</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarHeader>
          <SidebarContent><NavMain items={featureItems} /></SidebarContent>
        </div>
      )}
    </aside>
  );
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const [showMore, setShowMore] = React.useState(false);
  const pathname = usePathname();
  const chatLayout = pathname === "/chat" || pathname.startsWith("/chat/");
  const currentUser = useCurrentUser();
  const { sidebarVariant, sidebarCollapsible, isSynced } = usePreferencesStore(
    useShallow((s) => ({
      sidebarVariant: s.sidebarVariant,
      sidebarCollapsible: s.sidebarCollapsible,
      isSynced: s.isSynced,
    })),
  );

  const variant = isSynced ? sidebarVariant : props.variant;
  const collapsible = isSynced ? sidebarCollapsible : props.collapsible;

  const featureGroups = showMore ? sidebarItems : sidebarItems.filter((group) => group.id === 1);

  if (chatLayout) {
    return (
      <div className="flex h-svh shrink-0" data-src-agent-shell="chat-three-column">
        <SrcAgentRail featureItems={featureGroups} />
      </div>
    );
  }

  return (
    <div className="flex h-svh shrink-0" data-src-agent-shell="two-column">
      <SrcAgentRail />
      <Sidebar {...props} variant={variant} collapsible="none">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton asChild>
              <Link prefetch={false} href="/function/tasks">
                <SrcAgentMark size={32} />
                <span className="font-semibold text-base">{APP_CONFIG.name}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={featureGroups} />
        <div className="px-2 pb-2">
          <button
            type="button"
            onClick={() => setShowMore((value) => !value)}
            className="flex w-full items-center justify-center gap-2 rounded-md px-2 py-2 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            aria-expanded={showMore}
            aria-label={showMore ? "收起安全与资源功能" : "展开安全与资源功能"}
          >
            <MoreHorizontal className="size-4" />
            <span>{showMore ? "收起更多" : "更多功能"}</span>
          </button>
        </div>
        {/* <NavDocuments items={data.documents} /> */}
        {/* <NavSecondary items={data.navSecondary} className="mt-auto" /> */}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={currentUser} />
      </SidebarFooter>
      </Sidebar>
    </div>
  );
}
