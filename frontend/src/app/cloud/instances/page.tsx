"use client";

import { useRouter } from "next/navigation";
import useSWR from "swr";
import { fetcher } from "@/lib/api";
import { InstancesTable } from "@/components/cloud/InstancesTable";
import type { InstanceResponse } from "@/lib/types";

export default function InstancesPage() {
  const router = useRouter();
  const { data, error, isLoading, mutate } = useSWR<{
    data: InstanceResponse[];
    has_more: boolean;
  }>("/api/v1/instances", fetcher, {
    refreshInterval: 10000,
  });

  return (
    <div className="space-y-6">
      {/* Content */}
      {error ? (
        <div className="rounded-[10px] border border-border bg-bg-card/50 overflow-hidden">
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <p className="type-ui-sm text-red-400">
              Failed to load instances
            </p>
            <button
              onClick={() => mutate()}
              className="mt-3 type-ui-xs text-text-muted hover:text-text transition-colors"
            >
              Retry
            </button>
          </div>
        </div>
      ) : isLoading ? (
        <div className="rounded-[10px] border border-border bg-bg-card/50 overflow-hidden">
          <div className="space-y-0">
            {Array.from({ length: 4 }).map((_, i) => (
              <div
                key={i}
                className="flex items-center gap-4 px-4 py-3 border-b border-border/50"
              >
                <div className="h-4 bg-bg-card-hover rounded animate-pulse w-28" />
                <div className="h-4 bg-bg-card-hover rounded animate-pulse w-20" />
                <div className="h-4 bg-bg-card-hover rounded animate-pulse w-16" />
                <div className="h-4 bg-bg-card-hover rounded animate-pulse w-16" />
                <div className="h-4 bg-bg-card-hover rounded animate-pulse w-16" />
                <div className="h-4 bg-bg-card-hover rounded animate-pulse w-16" />
              </div>
            ))}
          </div>
        </div>
      ) : (
        <div className="rounded-[10px] border border-border bg-bg-card/50 overflow-hidden">
          <InstancesTable
            instances={data?.data ?? []}
            onRefresh={() => mutate()}
            onLaunch={() => router.push("/cloud/gpu-availability")}
          />
        </div>
      )}
    </div>
  );
}
