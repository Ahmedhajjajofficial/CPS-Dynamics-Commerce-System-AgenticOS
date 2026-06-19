/**
 * CPS Enterprise DCS - GlassBadge Component
 * Material3 Liquid Glass Design System
 * Copyright 2026 Ahmed Hajjaj - CP'S Enterprise Tech Solutions
 */

import * as React from "react"
import { cn } from "@/lib/utils"

interface GlassBadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: "default" | "success" | "warning" | "danger" | "info" | "neutral"
  size?: "sm" | "md"
  dot?: boolean
}

function GlassBadge({
  className,
  variant = "default",
  size = "sm",
  dot = false,
  children,
  ...props
}: GlassBadgeProps) {
  const variants = {
    default: "bg-[rgba(0,229,255,0.1)] text-[#00e5ff] border-[rgba(0,229,255,0.2)]",
    success: "bg-[rgba(16,185,129,0.1)] text-emerald-400 border-[rgba(16,185,129,0.2)]",
    warning: "bg-[rgba(245,158,11,0.1)] text-amber-400 border-[rgba(245,158,11,0.2)]",
    danger: "bg-[rgba(239,68,68,0.1)] text-red-400 border-[rgba(239,68,68,0.2)]",
    info: "bg-[rgba(99,102,241,0.1)] text-indigo-400 border-[rgba(99,102,241,0.2)]",
    neutral: "bg-white/5 text-slate-400 border-white/10",
  }

  const dotColors = {
    default: "bg-[#00e5ff]",
    success: "bg-emerald-400",
    warning: "bg-amber-400",
    danger: "bg-red-400",
    info: "bg-indigo-400",
    neutral: "bg-slate-400",
  }

  const sizes = {
    sm: "text-[10px] px-2 py-0.5",
    md: "text-xs px-2.5 py-1",
  }

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border font-medium text-brand",
        "backdrop-blur-sm",
        variants[variant],
        sizes[size],
        className
      )}
      {...props}
    >
      {dot && (
        <span className={cn("w-1.5 h-1.5 rounded-full", dotColors[variant])} />
      )}
      {children}
    </span>
  )
}

export { GlassBadge }
