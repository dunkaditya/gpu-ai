"use client";

import { cn } from "@/lib/utils";
import type { InstanceResponse } from "@/lib/types";

type InstanceStatus = InstanceResponse["status"];

const statusConfig: Record<
  InstanceStatus,
  { label: string; dotClass: string; pillClass: string }
> = {
  running: {
    label: "Running",
    dotClass: "bg-green",
    pillClass: "bg-green-dim text-green",
  },
  starting: {
    label: "Starting",
    dotClass: "bg-purple animate-pulse-dot",
    pillClass: "bg-purple-dim text-purple",
  },
  stopping: {
    label: "Stopping",
    dotClass: "bg-amber-400",
    pillClass: "bg-amber-400/10 text-amber-400",
  },
  terminated: {
    label: "Terminated",
    dotClass: "bg-text-dim",
    pillClass: "bg-bg-card text-text-dim",
  },
  error: {
    label: "Error",
    dotClass: "bg-red-500",
    pillClass: "bg-red-500/10 text-red-400",
  },
};

export function StatusBadge({ status }: { status: InstanceStatus }) {
  const config = statusConfig[status];

  return (
    <span
      className={cn(
        "type-ui-xs inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 font-medium",
        config.pillClass
      )}
    >
      <span
        className={cn("h-1.5 w-1.5 rounded-full", config.dotClass)}
        aria-hidden="true"
      />
      {config.label}
    </span>
  );
}
