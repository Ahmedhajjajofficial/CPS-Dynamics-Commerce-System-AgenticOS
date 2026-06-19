/**
 * CPS Enterprise DCS - GlassCard Component
 * Material3 Liquid Glass Design System
 * Copyright 2026 Ahmed Hajjaj - CP'S Enterprise Tech Solutions
 */

import * as React from "react"
import { cn } from "@/lib/utils"

interface GlassCardProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "elevated" | "subtle" | "surface" | "glow"
  interactive?: boolean
  noPadding?: boolean
}

const GlassCard = React.forwardRef<HTMLDivElement, GlassCardProps>(
  ({ className, variant = "default", interactive = false, noPadding = false, children, ...props }, ref) => {
    const variants = {
      default: "glass",
      elevated: "glass-elevated",
      subtle: "glass-subtle",
      surface: "glass-surface",
      glow: "glass glow-cyan",
    }

    return (
      <div
        ref={ref}
        className={cn(
          variants[variant],
          "rounded-[var(--glass-radius)]",
          interactive && "glass-interactive cursor-pointer",
          !noPadding && "p-6",
          className
        )}
        {...props}
      >
        {children}
      </div>
    )
  }
)
GlassCard.displayName = "GlassCard"

interface GlassCardHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  action?: React.ReactNode
}

function GlassCardHeader({ className, children, action, ...props }: GlassCardHeaderProps) {
  return (
    <div
      className={cn(
        "flex items-start justify-between gap-4 mb-4",
        className
      )}
      {...props}
    >
      <div className="flex-1 min-w-0">{children}</div>
      {action && <div className="shrink-0">{action}</div>}
    </div>
  )
}

function GlassCardTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3
      className={cn(
        "text-base font-semibold text-foreground tracking-tight text-brand",
        className
      )}
      {...props}
    />
  )
}

function GlassCardDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p
      className={cn(
        "text-sm text-muted-foreground mt-1 leading-relaxed",
        className
      )}
      {...props}
    />
  )
}

function GlassCardContent({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("text-sm", className)} {...props} />
  )
}

function GlassCardFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "flex items-center gap-3 mt-4 pt-4 border-t border-white/5",
        className
      )}
      {...props}
    />
  )
}

export {
  GlassCard,
  GlassCardHeader,
  GlassCardTitle,
  GlassCardDescription,
  GlassCardContent,
  GlassCardFooter,
}
