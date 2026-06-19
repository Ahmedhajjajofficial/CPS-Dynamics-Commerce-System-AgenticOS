/**
 * CPS Enterprise DCS - GlassInput Component
 * Material3 Liquid Glass Design System
 * Copyright 2026 Ahmed Hajjaj - CP'S Enterprise Tech Solutions
 */

import * as React from "react"
import { cn } from "@/lib/utils"

interface GlassInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  icon?: React.ReactNode
  label?: string
  error?: string
}

const GlassInput = React.forwardRef<HTMLInputElement, GlassInputProps>(
  ({ className, icon, label, error, id, ...props }, ref) => {
    const generatedId = React.useId()
    const inputId = id || generatedId

    return (
      <div className="space-y-2">
        {label && (
          <label
            htmlFor={inputId}
            className="block text-sm font-medium text-slate-300 text-brand"
          >
            {label}
          </label>
        )}
        <div className="relative">
          {icon && (
            <div className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none">
              {icon}
            </div>
          )}
          <input
            ref={ref}
            id={inputId}
            className={cn(
              "w-full h-10 rounded-xl text-sm text-brand",
              "bg-white/[0.04] border border-white/10",
              "backdrop-blur-sm",
              "text-slate-100 placeholder:text-slate-500",
              "px-4 py-2",
              icon && "pr-10",
              "outline-none transition-all duration-300",
              "focus:bg-white/[0.07] focus:border-[rgba(0,229,255,0.3)]",
              "focus:shadow-[0_0_0_3px_rgba(0,229,255,0.08)]",
              "hover:border-white/20",
              "disabled:opacity-40 disabled:cursor-not-allowed",
              error && "border-red-500/50 focus:border-red-500/70 focus:shadow-[0_0_0_3px_rgba(239,68,68,0.1)]",
              className
            )}
            {...props}
          />
        </div>
        {error && (
          <p className="text-xs text-red-400 mt-1">{error}</p>
        )}
      </div>
    )
  }
)
GlassInput.displayName = "GlassInput"

export { GlassInput }
