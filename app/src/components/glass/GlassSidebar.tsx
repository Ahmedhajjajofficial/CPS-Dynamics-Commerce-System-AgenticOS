/**
 * CPS Enterprise DCS - GlassSidebar Component
 * Material3 Liquid Glass Design System
 * Copyright 2026 Ahmed Hajjaj - CP'S Enterprise Tech Solutions
 */

import * as React from "react"
import { cn } from "@/lib/utils"

interface SidebarItem {
  id: string
  label: string
  icon: React.ReactNode
  badge?: string | number
  active?: boolean
}

interface SidebarGroup {
  title: string
  items: SidebarItem[]
}

interface GlassSidebarProps extends React.HTMLAttributes<HTMLElement> {
  groups: SidebarGroup[]
  collapsed?: boolean
  activeId?: string
  onItemClick?: (id: string) => void
}

function GlassSidebar({
  className,
  groups,
  collapsed = false,
  activeId,
  onItemClick,
  ...props
}: GlassSidebarProps) {
  return (
    <aside
      className={cn(
        "h-[calc(100vh-4rem)] sticky top-16",
        "glass-surface",
        "border-l border-white/[0.06]",
        "transition-all duration-300 ease-out",
        "flex flex-col",
        "overflow-y-auto overflow-x-hidden",
        collapsed ? "w-[68px]" : "w-[260px]",
        className
      )}
      {...props}
    >
      <nav className="flex-1 p-3 space-y-6">
        {groups.map((group, gi) => (
          <div key={gi}>
            {!collapsed && (
              <p className="text-[10px] font-semibold text-slate-500 uppercase tracking-widest px-3 mb-2 text-brand">
                {group.title}
              </p>
            )}
            <ul className="space-y-1">
              {group.items.map((item) => {
                const isActive = item.id === activeId
                return (
                  <li key={item.id}>
                    <button
                      onClick={() => onItemClick?.(item.id)}
                      className={cn(
                        "w-full flex items-center gap-3 rounded-xl text-sm font-medium",
                        "transition-all duration-200 ease-out",
                        "text-brand",
                        collapsed ? "justify-center p-2.5" : "px-3 py-2.5",
                        isActive
                          ? "bg-[rgba(0,229,255,0.1)] text-[#00e5ff] border border-[rgba(0,229,255,0.15)] shadow-[0_0_20px_rgba(0,229,255,0.05)]"
                          : "text-slate-400 hover:text-slate-200 hover:bg-white/[0.04] border border-transparent"
                      )}
                    >
                      <span className={cn(
                        "shrink-0 transition-colors",
                        isActive ? "text-[#00e5ff]" : "text-slate-500"
                      )}>
                        {item.icon}
                      </span>
                      {!collapsed && (
                        <>
                          <span className="flex-1 text-right truncate">{item.label}</span>
                          {item.badge && (
                            <span className={cn(
                              "text-[10px] font-bold px-1.5 py-0.5 rounded-md",
                              isActive
                                ? "bg-[rgba(0,229,255,0.15)] text-[#00e5ff]"
                                : "bg-white/5 text-slate-500"
                            )}>
                              {item.badge}
                            </span>
                          )}
                        </>
                      )}
                    </button>
                  </li>
                )
              })}
            </ul>
          </div>
        ))}
      </nav>

      {/* Footer Branding */}
      {!collapsed && (
        <div className="p-4 border-t border-white/[0.04]">
          <div className="text-center">
            <p className="text-[9px] text-slate-600 font-medium text-brand">
              Developed by Ahmed Hajjaj
            </p>
            <p className="text-[8px] text-slate-700 mt-0.5">
              CP'S Enterprise Tech Solutions
            </p>
          </div>
        </div>
      )}
    </aside>
  )
}

export { GlassSidebar }
export type { SidebarItem, SidebarGroup }
