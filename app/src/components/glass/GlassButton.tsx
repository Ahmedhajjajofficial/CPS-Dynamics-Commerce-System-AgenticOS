/**
 * CPS Enterprise DCS - GlassButton Component
 * Material3 Liquid Glass Design System
 * Copyright 2026 Ahmed Hajjaj - CP'S Enterprise Tech Solutions
 */

import * as React from "react"
import { cn } from "@/lib/utils"

interface GlassButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "ghost" | "outline" | "danger" | "glow"
  size?: "sm" | "md" | "lg" | "icon"
  loading?: boolean
}

const GlassButton = React.forwardRef<HTMLButtonElement, GlassButtonProps>(
  ({ className, variant = "primary", size = "md", loading = false, children, disabled, ...props }, ref) => {
    const variants = {
      primary:
        "bg-[rgba(0,229,255,0.12)] border-[rgba(0,229,255,0.3)] text-[#00e5ff] hover:bg-[rgba(0,229,255,0.2)] hover:border-[rgba(0,229,255,0.5)] hover:shadow-[0_0_30px_rgba(0,229,255,0.15)] active:bg-[rgba(0,229,255,0.25)]",
      secondary:
        "bg-white/5 border-white/10 text-slate-200 hover:bg-white/10 hover:border-white/20 active:bg-white/15",
      ghost:
        "bg-transparent border-transparent text-slate-300 hover:bg-white/5 hover:text-slate-100 active:bg-white/10",
      outline:
        "bg-transparent border-white/15 text-slate-200 hover:bg-white/5 hover:border-white/25 active:bg-white/10",
      danger:
        "bg-[rgba(239,68,68,0.1)] border-[rgba(239,68,68,0.3)] text-red-400 hover:bg-[rgba(239,68,68,0.2)] hover:border-[rgba(239,68,68,0.4)] active:bg-[rgba(239,68,68,0.25)]",
      glow:
        "bg-[rgba(0,229,255,0.15)] border-[rgba(0,229,255,0.4)] text-[#00e5ff] shadow-[0_0_20px_rgba(0,229,255,0.1)] hover:bg-[rgba(0,229,255,0.25)] hover:shadow-[0_0_40px_rgba(0,229,255,0.2)] hover:border-[rgba(0,229,255,0.6)] active:bg-[rgba(0,229,255,0.3)]",
    }

    const sizes = {
      sm: "h-8 px-3 text-xs gap-1.5 rounded-lg",
      md: "h-10 px-5 text-sm gap-2 rounded-xl",
      lg: "h-12 px-7 text-base gap-2.5 rounded-xl",
      icon: "h-10 w-10 rounded-xl",
    }

    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={cn(
          "inline-flex items-center justify-center font-medium",
          "border backdrop-blur-sm",
          "transition-all duration-300 ease-out",
          "outline-none focus-visible:ring-2 focus-visible:ring-[rgba(0,229,255,0.5)] focus-visible:ring-offset-2 focus-visible:ring-offset-[#020617]",
          "disabled:opacity-40 disabled:pointer-events-none",
          "text-brand whitespace-nowrap select-none",
          variants[variant],
          sizes[size],
          className
        )}
        {...props}
      >
        {loading && (
          <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
        )}
        {children}
      </button>
    )
  }
)
GlassButton.displayName = "GlassButton"

export { GlassButton }
