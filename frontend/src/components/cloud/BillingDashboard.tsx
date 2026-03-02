"use client";

import { useState } from "react";
import useSWR from "swr";
import { cn } from "@/lib/utils";
import { fetcher } from "@/lib/api";
import type { UsageResponse } from "@/lib/types";

const periods = [
  { label: "Current Month", value: "current_month" },
  { label: "Last Month", value: "last_month" },
  { label: "Last 7 Days", value: "last_7_days" },
] as const;

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatDuration(seconds?: number | null): string {
  if (!seconds) return "--";
  const hours = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  if (hours > 0) return `${hours}h ${mins}m`;
  return `${mins}m`;
}

function SkeletonRow() {
  return (
    <tr className="border-b border-border/50">
      {Array.from({ length: 7 }).map((_, i) => (
        <td key={i} className="px-4 py-3">
          <div className="h-4 bg-bg-card-hover rounded animate-pulse w-16" />
        </td>
      ))}
    </tr>
  );
}

export function BillingDashboard() {
  const [period, setPeriod] = useState("current_month");
  const { data, error, isLoading, mutate } = useSWR<UsageResponse>(
    `/api/v1/billing/usage?period=${period}`,
    fetcher,
    { refreshInterval: 60000 }
  );

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <p className="type-ui-sm text-red-400">Failed to load billing data</p>
        <button
          onClick={() => mutate()}
          className="mt-3 type-ui-xs text-purple hover:text-purple-light transition-colors"
        >
          Retry
        </button>
      </div>
    );
  }

  const sessions = data?.sessions ?? [];
  const activeSessions = sessions.filter((s) => s.is_active);

  return (
    <div className="space-y-6">
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Total Cost */}
        <div className="bg-bg-card border border-border rounded-xl p-5">
          <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mb-2">
            Total Cost
          </p>
          {isLoading ? (
            <div className="h-10 bg-bg-card-hover rounded animate-pulse w-32" />
          ) : (
            <p className="type-h2 font-mono text-text">
              <span className="text-text-muted">$</span>
              {(data?.total_cost ?? 0).toFixed(2)}
            </p>
          )}
          {data?.currency && (
            <p className="type-ui-2xs text-text-dim mt-1">{data.currency}</p>
          )}
        </div>

        {/* Active Sessions */}
        <div className="bg-bg-card border border-border rounded-xl p-5">
          <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mb-2">
            Active Sessions
          </p>
          {isLoading ? (
            <div className="h-10 bg-bg-card-hover rounded animate-pulse w-12" />
          ) : (
            <p className="type-h2 font-mono text-green">
              {activeSessions.length}
            </p>
          )}
          <p className="type-ui-2xs text-text-dim mt-1">Currently running</p>
        </div>

        {/* Total Sessions */}
        <div className="bg-bg-card border border-border rounded-xl p-5">
          <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mb-2">
            Total Sessions
          </p>
          {isLoading ? (
            <div className="h-10 bg-bg-card-hover rounded animate-pulse w-12" />
          ) : (
            <p className="type-h2 font-mono text-text">{sessions.length}</p>
          )}
          <p className="type-ui-2xs text-text-dim mt-1">In this period</p>
        </div>
      </div>

      {/* Period Selector */}
      <div className="flex rounded-lg border border-border overflow-hidden w-fit">
        {periods.map((p) => (
          <button
            key={p.value}
            onClick={() => setPeriod(p.value)}
            className={cn(
              "px-4 py-2 type-ui-xs font-medium transition-colors",
              period === p.value
                ? "bg-purple-dim text-purple-light"
                : "text-text-muted hover:text-text hover:bg-bg-card"
            )}
          >
            {p.label}
          </button>
        ))}
      </div>

      {/* Sessions Table */}
      <div className="rounded-lg border border-border bg-bg-card/50 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border">
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  GPU
                </th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  Count
                </th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  Rate
                </th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  Started
                </th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  Ended
                </th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  Duration
                </th>
                <th className="type-ui-2xs text-right text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  Cost
                </th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 4 }).map((_, i) => (
                  <SkeletonRow key={i} />
                ))
              ) : sessions.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center">
                    <p className="type-ui-sm text-text-muted">
                      No billing sessions in this period
                    </p>
                    <p className="type-ui-2xs text-text-dim mt-1">
                      Launch an instance to start tracking usage.
                    </p>
                  </td>
                </tr>
              ) : (
                sessions.map((session) => (
                  <tr
                    key={session.id}
                    className="border-b border-border/50 hover:bg-bg-card transition-colors"
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        {session.is_active && (
                          <span className="h-1.5 w-1.5 rounded-full bg-green animate-pulse-dot shrink-0" />
                        )}
                        <span className="type-ui-sm text-text font-medium">
                          {session.gpu_type}
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className="type-ui-sm text-text font-mono">
                        x{session.gpu_count}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className="type-ui-sm text-text-muted font-mono">
                        ${session.price_per_hour.toFixed(2)}/hr
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className="type-ui-xs text-text-muted">
                        {formatDate(session.started_at)}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      {session.ended_at ? (
                        <span className="type-ui-xs text-text-muted">
                          {formatDate(session.ended_at)}
                        </span>
                      ) : (
                        <span className="type-ui-xs inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 bg-green-dim text-green font-medium">
                          <span className="h-1.5 w-1.5 rounded-full bg-green" />
                          Active
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span className="type-ui-sm text-text-muted font-mono">
                        {formatDuration(session.duration_seconds)}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <span className="type-ui-sm text-text font-mono">
                        $
                        {(
                          session.total_cost ??
                          session.estimated_cost ??
                          0
                        ).toFixed(2)}
                      </span>
                      {session.is_active && session.estimated_cost != null && (
                        <p className="type-ui-2xs text-text-dim">est.</p>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
