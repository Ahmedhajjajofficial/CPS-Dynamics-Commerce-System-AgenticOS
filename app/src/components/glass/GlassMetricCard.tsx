/**
 * CPS Enterprise DCS - GlassMetricCard Component
 * Material3 Liquid Glass Design System
 * Copyright 2026 Ahmed Hajjaj - CP'S Enterprise Tech Solutions
 */

import * as React from "react"
import { cn } from "@/lib/utils"
import { GlassCard } from "./GlassCard"

interface GlassMetricCardProps {
  title: string
  value: string | number
  change?: string
  changeType?: "positive" | "negative" | "neutral"
  icon: React.ReactNode
  iconColor?: string
  subtitle?: string
  className?: string
  animationDelay?: string
}

function GlassMetricCard({
  title,
  value,
  change,
  changeType = "neutral",
  icon,
  iconColor = "text-[#00e5ff]",
  subtitle,
  className,
  animationDelay,
}: GlassMetricCardProps) {
  const changeColors = {
    positive: "text-emerald-400",
    negative: "text-red-400",
    neutral: "text-slate-400",
  }

  const changeIcons = {
    positive: "↑",
    negative: "↓",
    neutral: "→",
  }

  return (
    <GlassCard
      interactive
      className={cn(
        "animate-fade-up opacity-0",
        className
      )}
      style={{ animationDelay: animationDelay || "0s", animationFillMode: "forwards" }}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <p className="text-xs font-medium text-slate-500 text-brand mb-2">{title}</p>
          <p className="text-2xl font-bold text-white text-display tracking-tight">{value}</p>
          {change && (
            <div className="flex items-center gap-1.5 mt-2">
              <span className={cn("text-xs font-semibold", changeColors[changeType])}>
                {changeIcons[changeType]} {change}
              </span>
              {subtitle && (
                <span className="text-[10px] text-slate-600">{subtitle}</span>
              )}
            </div>
          )}
        </div>
        <div className={cn(
          "w-11 h-11 rounded-2xl flex items-center justify-center",
          "bg-white/[0.04] border border-white/[0.06]",
          iconColor
        )}>
          {icon}
        </div>
      </div>
    </GlassCard>
  )
}

export { GlassMetricCard }
