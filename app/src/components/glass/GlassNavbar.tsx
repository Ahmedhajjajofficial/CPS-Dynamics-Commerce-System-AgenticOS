/**
 * CPS Enterprise DCS - GlassNavbar Component
 * Material3 Liquid Glass Design System
 * Copyright 2026 Ahmed Hajjaj - CP'S Enterprise Tech Solutions
 */

import * as React from "react"
import { cn } from "@/lib/utils"
import { GlassButton } from "./GlassButton"
import { GlassInput } from "./GlassInput"
import { GlassBadge } from "./GlassBadge"

interface GlassNavbarProps extends React.HTMLAttributes<HTMLElement> {
  onMenuToggle?: () => void
}

function GlassNavbar({ className, onMenuToggle, ...props }: GlassNavbarProps) {
  return (
    <header
      className={cn(
        "sticky top-0 z-50 w-full",
        "glass-elevated",
        "border-b border-white/[0.06]",
        className
      )}
      {...props}
    >
      <div className="flex items-center justify-between h-16 px-6">
        {/* Right Side - Logo & Menu */}
        <div className="flex items-center gap-4">
          <GlassButton
            variant="ghost"
            size="icon"
            onClick={onMenuToggle}
            aria-label="القائمة"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="15" y2="12" />
              <line x1="3" y1="18" x2="18" y2="18" />
            </svg>
          </GlassButton>

          <div className="flex items-center gap-3">
            <div className="relative w-9 h-9 rounded-xl bg-gradient-to-br from-[#00e5ff] to-[#7c3aed] p-[1px]">
              <div className="w-full h-full rounded-xl bg-slate-900 flex items-center justify-center">
                <span className="text-[#00e5ff] font-bold text-sm">CP</span>
              </div>
            </div>
            <div className="hidden sm:block">
              <h1 className="text-sm font-bold text-white tracking-wide text-brand">CPS Enterprise</h1>
              <p className="text-[10px] text-slate-500 font-medium">Dynamics Commerce System</p>
            </div>
          </div>
        </div>

        {/* Center - Search */}
        <div className="hidden md:flex flex-1 max-w-md mx-8">
          <GlassInput
            placeholder="بحث في النظام..."
            icon={
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                <circle cx="11" cy="11" r="8" />
                <line x1="21" y1="21" x2="16.65" y2="16.65" />
              </svg>
            }
          />
        </div>

        {/* Left Side - Actions */}
        <div className="flex items-center gap-2">
          {/* Notifications */}
          <GlassButton variant="ghost" size="icon" className="relative">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
              <path d="M13.73 21a2 2 0 0 1-3.46 0" />
            </svg>
            <span className="absolute -top-0.5 -right-0.5 w-4 h-4 bg-[#00e5ff] rounded-full text-[9px] font-bold text-slate-900 flex items-center justify-center">
              3
            </span>
          </GlassButton>

          {/* Status */}
          <GlassBadge variant="success" dot className="hidden sm:inline-flex">
            متصل
          </GlassBadge>

          {/* User Avatar */}
          <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-slate-700 to-slate-800 border border-white/10 flex items-center justify-center cursor-pointer hover:border-white/20 transition-colors">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" className="text-slate-400">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
              <circle cx="12" cy="7" r="4" />
            </svg>
          </div>
        </div>
      </div>
    </header>
  )
}

export { GlassNavbar }
