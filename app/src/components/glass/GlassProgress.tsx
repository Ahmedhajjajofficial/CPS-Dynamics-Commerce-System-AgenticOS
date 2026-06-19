/**
 * CPS Enterprise DCS - GlassProgress Component
 * Material3 Liquid Glass Design System
 * Copyright 2026 Ahmed Hajjaj - CP'S Enterprise Tech Solutions
 */

import { cn } from "@/lib/utils"

interface GlassProgressProps {
  value: number
  max?: number
  label?: string
  showValue?: boolean
  color?: "cyan" | "emerald" | "amber" | "rose" | "indigo"
  size?: "sm" | "md" | "lg"
  className?: string
}

function GlassProgress({
  value,
  max = 100,
  label,
  showValue = false,
  color = "cyan",
  size = "md",
  className,
}: GlassProgressProps) {
  const percentage = Math.min(Math.max((value / max) * 100, 0), 100)

  const colors = {
    cyan: "from-[#00e5ff] to-[#00b8d4]",
    emerald: "from-emerald-400 to-emerald-500",
    amber: "from-amber-400 to-amber-500",
    rose: "from-rose-400 to-rose-500",
    indigo: "from-indigo-400 to-indigo-500",
  }

  const glowColors = {
    cyan: "shadow-[0_0_12px_rgba(0,229,255,0.3)]",
    emerald: "shadow-[0_0_12px_rgba(16,185,129,0.3)]",
    amber: "shadow-[0_0_12px_rgba(245,158,11,0.3)]",
    rose: "shadow-[0_0_12px_rgba(244,63,94,0.3)]",
    indigo: "shadow-[0_0_12px_rgba(99,102,241,0.3)]",
  }

  const sizes = {
    sm: "h-1",
    md: "h-2",
    lg: "h-3",
  }

  return (
    <div className={cn("space-y-1.5", className)}>
      {(label || showValue) && (
        <div className="flex items-center justify-between">
          {label && <span className="text-xs text-slate-400 text-brand">{label}</span>}
          {showValue && <span className="text-xs font-semibold text-slate-300 text-mono">{Math.round(percentage)}%</span>}
        </div>
      )}
      <div className={cn(
        "w-full rounded-full overflow-hidden",
        "bg-white/[0.04]",
        sizes[size]
      )}>
        <div
          className={cn(
            "h-full rounded-full transition-all duration-700 ease-out",
            "bg-gradient-to-r",
            colors[color],
            glowColors[color]
          )}
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  )
}

export { GlassProgress }
