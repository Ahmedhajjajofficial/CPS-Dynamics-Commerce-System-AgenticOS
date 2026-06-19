/**
 * CPS Enterprise DCS - GlassTable Component
 * Material3 Liquid Glass Design System
 * Copyright 2026 Ahmed Hajjaj - CP'S Enterprise Tech Solutions
 */

import * as React from "react"
import { cn } from "@/lib/utils"
import { GlassBadge } from "./GlassBadge"

interface Column<T> {
  key: string
  header: string
  render?: (item: T) => React.ReactNode
  align?: "right" | "left" | "center"
}

interface GlassTableProps<T> {
  columns: Column<T>[]
  data: T[]
  className?: string
  onRowClick?: (item: T) => void
}

function GlassTable<T extends Record<string, unknown>>({
  columns,
  data,
  className,
  onRowClick,
}: GlassTableProps<T>) {
  return (
    <div className={cn("overflow-x-auto", className)}>
      <table className="w-full">
        <thead>
          <tr className="border-b border-white/[0.06]">
            {columns.map((col) => (
              <th
                key={col.key}
                className={cn(
                  "px-4 py-3 text-[11px] font-semibold text-slate-500 uppercase tracking-wider text-brand",
                  col.align === "left" ? "text-left" : col.align === "center" ? "text-center" : "text-right"
                )}
              >
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((item, i) => (
            <tr
              key={i}
              onClick={() => onRowClick?.(item)}
              className={cn(
                "border-b border-white/[0.03] transition-colors duration-200",
                "hover:bg-white/[0.02]",
                onRowClick && "cursor-pointer"
              )}
            >
              {columns.map((col) => (
                <td
                  key={col.key}
                  className={cn(
                    "px-4 py-3.5 text-sm text-brand",
                    col.align === "left" ? "text-left" : col.align === "center" ? "text-center" : "text-right"
                  )}
                >
                  {col.render
                    ? col.render(item)
                    : (item[col.key] as React.ReactNode)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export { GlassTable, GlassBadge }
export type { Column }
