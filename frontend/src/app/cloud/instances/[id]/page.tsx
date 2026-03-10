"use client";

import { useParams } from "next/navigation";
import useSWR from "swr";
import { fetcher } from "@/lib/api";
import { InstanceDetail } from "@/components/cloud/InstanceDetail";
import Link from "next/link";
import type { InstanceResponse } from "@/lib/types";

function DetailSkeleton() {
  const pulseClass = "h-4 bg-bg-card-hover rounded animate-pulse";

  return (
    <div className="space-y-6">
      {/* Back link skeleton */}
      <div className={`${pulseClass} w-36`} />

      {/* Header skeleton */}
      <div className="flex items-center gap-4">
        <div className={`${pulseClass} w-48 h-6`} />
        <div className={`${pulseClass} w-20`} />
      </div>

      {/* Info grid skeleton */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div
            key={i}
            className="bg-bg-card border border-border rounded-xl p-5 space-y-3"
          >
            <div className={`${pulseClass} w-24`} />
            <div className={`${pulseClass} w-full h-6`} />
            <div className={`${pulseClass} w-3/4`} />
          </div>
        ))}
      </div>
    </div>
  );
}

export default function InstanceDetailPage() {
  const params = useParams<{ id: string }>();
  const {
    data: instance,
    error,
    isLoading,
    mutate,
  } = useSWR<InstanceResponse>(`/api/v1/instances/${params.id}`, fetcher, {
    refreshInterval: 10000,
  });

  if (isLoading) {
    return <DetailSkeleton />;
  }

  if (error || !instance) {
    return (
      <div className="space-y-6">
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="w-12 h-12 rounded-full bg-red-500/10 flex items-center justify-center mb-4">
            <svg
              width="20"
              height="20"
              viewBox="0 0 16 16"
              fill="none"
              className="text-red-400"
            >
              <circle
                cx="8"
                cy="8"
                r="6.5"
                stroke="currentColor"
                strokeWidth="1.5"
              />
              <path
                d="M8 5V9"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
              />
              <circle cx="8" cy="11.5" r="0.75" fill="currentColor" />
            </svg>
          </div>
          <p className="type-ui-sm text-text-muted">Instance not found</p>
          <p className="type-ui-2xs text-text-dim mt-1 mb-4">
            This instance may have been removed or the ID is invalid.
          </p>
          <Link
            href="/cloud/instances"
            className="type-ui-sm text-purple hover:text-purple-light transition-colors"
          >
            Back to Instances
          </Link>
        </div>
      </div>
    );
  }

  return <InstanceDetail instance={instance} onRefresh={() => mutate()} />;
}
