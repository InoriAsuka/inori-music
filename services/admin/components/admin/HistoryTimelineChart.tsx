"use client";

import { useState } from "react";
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer,
} from "recharts";

interface Bucket { label: string; count: number; }
type Granularity = "day" | "week" | "month";

interface HistoryTimelineChartProps {
  buckets: Bucket[];
  granularity?: Granularity;
  onGranularityChange?: (g: Granularity) => void;
}

export function HistoryTimelineChart({
  buckets,
  granularity = "day",
  onGranularityChange,
}: HistoryTimelineChartProps) {
  const GRANULARITIES: Granularity[] = ["day", "week", "month"];

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
          Play Timeline
        </h3>
        <div className="flex gap-1">
          {GRANULARITIES.map((g) => (
            <button
              key={g}
              onClick={() => onGranularityChange?.(g)}
              className={
                granularity === g
                  ? "rounded bg-[var(--color-primary)] px-2 py-0.5 text-xs font-semibold text-[var(--color-primary-fg)]"
                  : "rounded px-2 py-0.5 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
              }
            >
              {g}
            </button>
          ))}
        </div>
      </div>

      <div className="h-40">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={buckets} margin={{ top: 4, right: 4, left: -24, bottom: 0 }}>
            <defs>
              <linearGradient id="histGrad" x1="0" y1="0" x2="1" y2="0">
                <stop offset="0%" stopColor="#9b5cff" stopOpacity={0.8} />
                <stop offset="100%" stopColor="#0fd4c0" stopOpacity={0.8} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#1e1e3f" />
            <XAxis dataKey="label" tick={{ fill: "#4a4a7a", fontSize: 10 }} axisLine={false} tickLine={false} />
            <YAxis tick={{ fill: "#4a4a7a", fontSize: 10 }} axisLine={false} tickLine={false} />
            <Tooltip
              contentStyle={{ background: "#0d0d1a", border: "1px solid #1e1e3f", borderRadius: 8, fontSize: 12, color: "#e8e8f4" }}
              cursor={{ stroke: "#9b5cff", strokeWidth: 1 }}
            />
            <Area
              type="monotone"
              dataKey="count"
              stroke="url(#histGrad)"
              fill="url(#histGrad)"
              fillOpacity={0.2}
              strokeWidth={2}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
