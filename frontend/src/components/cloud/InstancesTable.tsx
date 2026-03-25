"use client";

import { useState, useRef, useEffect, useMemo } from "react";
import Link from "next/link";
import { cn } from "@/lib/utils";
import { StatusBadge } from "@/components/cloud/StatusBadge";
import { ConfirmDialog } from "@/components/cloud/ConfirmDialog";
import { EmptyState } from "@/components/cloud/EmptyState";
import { terminateInstance, renameInstance } from "@/lib/api";
import { getDisplayName } from "@/lib/gpu-categories";
import { getRegionDisplay } from "@/lib/regions";
import type { InstanceResponse } from "@/lib/types";

/* -- Utility Functions -- */

function formatUptime(instance: InstanceResponse): string {
  if (instance.status === "terminated") return "--";
  const start = instance.ready_at ?? instance.created_at;
  if (!start) return "--";
  const elapsed = Math.floor(
    (Date.now() - new Date(start).getTime()) / 1000
  );
  if (elapsed < 0) return "--";
  const hours = Math.floor(elapsed / 3600);
  const minutes = Math.floor((elapsed % 3600) / 60);
  if (hours >= 24) {
    const days = Math.floor(hours / 24);
    const remainingHours = hours % 24;
    return `${days}d ${remainingHours}h`;
  }
  return `${hours}h ${minutes}m`;
}

function displayName(instance: InstanceResponse) {
  return instance.name ?? instance.id.slice(0, 12);
}

/* -- Status Filter -- */

const ALL_STATUSES = ["running", "starting", "stopping", "terminated", "error"] as const;

const STATUS_DOT_COLORS: Record<string, string> = {
  running: "bg-green",
  starting: "bg-purple",
  stopping: "bg-amber-400",
  terminated: "bg-text-dim",
  error: "bg-red-500",
};

const STATUS_LABELS: Record<string, string> = {
  running: "Running",
  starting: "Starting",
  stopping: "Stopping",
  terminated: "Terminated",
  error: "Error",
};

function StatusFilter({
  selected,
  onChange,
}: {
  selected: Set<string>;
  onChange: (s: Set<string>) => void;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    function handleKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    if (open) {
      document.addEventListener("mousedown", handleClick);
      document.addEventListener("keydown", handleKey);
    }
    return () => {
      document.removeEventListener("mousedown", handleClick);
      document.removeEventListener("keydown", handleKey);
    };
  }, [open]);

  function toggle(status: string) {
    const next = new Set(selected);
    if (next.has(status)) next.delete(status);
    else next.add(status);
    onChange(next);
  }

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen(!open)}
        className={cn(
          "w-full flex items-center gap-2 px-3 py-1.5 rounded-lg border transition-colors type-ui-sm font-medium",
          open
            ? "border-border-light bg-bg-card text-text"
            : "border-border bg-bg-card/50 text-text-muted hover:border-border-light hover:text-text"
        )}
      >
        <span className="inline-flex items-center gap-1">
          {ALL_STATUSES.map((s) => (
            <span
              key={s}
              className={cn(
                "w-1.5 h-1.5 rounded-full transition-opacity",
                STATUS_DOT_COLORS[s],
                selected.has(s) ? "opacity-100" : "opacity-20"
              )}
            />
          ))}
        </span>
        <span>Status</span>
        <svg
          width="12"
          height="12"
          viewBox="0 0 12 12"
          fill="none"
          className={cn(
            "transition-transform",
            open && "rotate-180"
          )}
        >
          <path
            d="M3 4.5L6 7.5L9 4.5"
            stroke="currentColor"
            strokeWidth="1.25"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </button>

      {open && (
        <div className="absolute top-full left-0 mt-1.5 z-50 min-w-[180px] rounded-lg border border-border bg-bg-card shadow-lg py-1">
          {ALL_STATUSES.map((status) => (
            <button
              key={status}
              onClick={() => toggle(status)}
              className="flex items-center gap-2.5 w-full px-3 py-2.5 sm:py-1.5 text-left type-ui-sm hover:bg-bg-card-hover transition-colors"
            >
              <span
                className={cn(
                  "w-3.5 h-3.5 rounded border flex items-center justify-center transition-colors",
                  selected.has(status)
                    ? "border-text-muted bg-text-muted"
                    : "border-border-light bg-transparent"
                )}
              >
                {selected.has(status) && (
                  <svg width="10" height="10" viewBox="0 0 10 10" fill="none">
                    <path
                      d="M2 5L4 7L8 3"
                      stroke="currentColor"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      className="text-bg"
                    />
                  </svg>
                )}
              </span>
              <span className={cn("w-1.5 h-1.5 rounded-full", STATUS_DOT_COLORS[status])} />
              <span className={cn(
                "font-medium",
                selected.has(status) ? "text-text" : "text-text-muted"
              )}>
                {STATUS_LABELS[status]}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

/* -- Region Filter -- */

function RegionFilter({
  regions,
  selected,
  onChange,
}: {
  regions: string[];
  selected: Set<string>;
  onChange: (s: Set<string>) => void;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    function handleKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    if (open) {
      document.addEventListener("mousedown", handleClick);
      document.addEventListener("keydown", handleKey);
    }
    return () => {
      document.removeEventListener("mousedown", handleClick);
      document.removeEventListener("keydown", handleKey);
    };
  }, [open]);

  function toggle(region: string) {
    const next = new Set(selected);
    if (next.has(region)) next.delete(region);
    else next.add(region);
    onChange(next);
  }

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen(!open)}
        className={cn(
          "w-full flex items-center gap-2 px-3 py-1.5 rounded-lg border transition-colors type-ui-sm font-medium",
          open
            ? "border-border-light bg-bg-card text-text"
            : selected.size > 0
              ? "border-border-light bg-bg-card/50 text-text hover:border-border-light"
              : "border-border bg-bg-card/50 text-text-muted hover:border-border-light hover:text-text"
        )}
      >
        <span>Region{selected.size > 0 ? ` ${selected.size}/${regions.length}` : ""}</span>
        <svg
          width="12"
          height="12"
          viewBox="0 0 12 12"
          fill="none"
          className={cn("transition-transform", open && "rotate-180")}
        >
          <path
            d="M3 4.5L6 7.5L9 4.5"
            stroke="currentColor"
            strokeWidth="1.25"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </button>

      {open && (
        <div className="absolute top-full left-0 mt-1.5 z-50 min-w-[180px] rounded-lg border border-border bg-bg-card shadow-lg py-1">
          {selected.size > 0 && (
            <button
              onClick={() => onChange(new Set())}
              className="flex items-center w-full px-3 py-1.5 text-left type-ui-2xs text-text-dim hover:text-text hover:bg-bg-card-hover transition-colors"
            >
              Clear all
            </button>
          )}
          {regions.map((region) => (
            <button
              key={region}
              onClick={() => toggle(region)}
              className="flex items-center gap-2.5 w-full px-3 py-2.5 sm:py-1.5 text-left type-ui-sm hover:bg-bg-card-hover transition-colors"
            >
              <span
                className={cn(
                  "w-3.5 h-3.5 rounded border flex items-center justify-center transition-colors",
                  selected.has(region)
                    ? "border-text-muted bg-text-muted"
                    : "border-border-light bg-transparent"
                )}
              >
                {selected.has(region) && (
                  <svg width="10" height="10" viewBox="0 0 10 10" fill="none">
                    <path
                      d="M2 5L4 7L8 3"
                      stroke="currentColor"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      className="text-bg"
                    />
                  </svg>
                )}
              </span>
              <span className={cn(
                "font-medium",
                selected.has(region) ? "text-text" : "text-text-muted"
              )}>
                {(() => { const r = getRegionDisplay(region); return `${r.flag} ${r.label}`; })()}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

/* -- CopyButton -- */

function CopyButton({
  text,
  onClick,
}: {
  text: string;
  onClick?: (e: React.MouseEvent) => void;
}) {
  const [copied, setCopied] = useState(false);

  async function handleCopy(e: React.MouseEvent) {
    e.stopPropagation();
    e.preventDefault();
    onClick?.(e);
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard API may fail in non-secure contexts
    }
  }

  return (
    <button
      onClick={handleCopy}
      className={cn(
        "inline-flex items-center justify-center w-10 h-10 sm:w-7 sm:h-7 rounded transition-colors shrink-0",
        copied
          ? "text-green bg-green-dim"
          : "text-text-dim hover:text-text hover:bg-bg-card-hover"
      )}
      title={copied ? "Copied!" : "Copy to clipboard"}
    >
      {copied ? (
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <path
            d="M2.5 7.5L5.5 10.5L11.5 4.5"
            stroke="currentColor"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      ) : (
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <rect
            x="4.5"
            y="4.5"
            width="8"
            height="8"
            rx="1"
            stroke="currentColor"
            strokeWidth="1.25"
          />
          <path
            d="M9.5 4.5V2.5C9.5 1.95 9.05 1.5 8.5 1.5H2.5C1.95 1.5 1.5 1.95 1.5 2.5V8.5C1.5 9.05 1.95 9.5 2.5 9.5H4.5"
            stroke="currentColor"
            strokeWidth="1.25"
          />
        </svg>
      )}
    </button>
  );
}

/* -- EditableName -- */

function EditableName({
  instance,
  onRename,
}: {
  instance: InstanceResponse;
  onRename: (id: string, name: string) => Promise<void>;
}) {
  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState(displayName(instance));
  const [saving, setSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing) {
      inputRef.current?.focus();
      inputRef.current?.select();
    }
  }, [editing]);

  async function save() {
    const trimmed = value.trim();
    if (!trimmed || trimmed === displayName(instance)) {
      setEditing(false);
      setValue(displayName(instance));
      return;
    }
    setSaving(true);
    try {
      await onRename(instance.id, trimmed);
    } catch {
      setValue(displayName(instance));
    } finally {
      setSaving(false);
      setEditing(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") {
      e.preventDefault();
      save();
    } else if (e.key === "Escape") {
      e.preventDefault();
      setValue(displayName(instance));
      setEditing(false);
    }
  }

  if (editing) {
    return (
      <div className="flex flex-col" onClick={(e) => e.stopPropagation()}>
        <input
          ref={inputRef}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onBlur={save}
          onKeyDown={handleKeyDown}
          disabled={saving}
          className="type-ui-sm text-text font-medium bg-bg border border-border-light rounded px-2 py-0.5 focus:outline-none focus:ring-1 focus:ring-border-light w-40"
        />
      </div>
    );
  }

  return (
    <div
      className="group/name flex flex-col cursor-pointer min-h-[44px] sm:min-h-0 justify-center"
      onClick={(e) => {
        e.stopPropagation();
        e.preventDefault();
        setEditing(true);
      }}
    >
      <span className="text-[15px] text-text font-bold inline-flex items-center gap-1.5">
        {displayName(instance)}
        <svg
          width="12"
          height="12"
          viewBox="0 0 12 12"
          fill="none"
          className="opacity-0 group-hover/name:opacity-60 transition-opacity shrink-0"
        >
          <path
            d="M8.5 1.5L10.5 3.5M1 11H3L9.5 4.5L7.5 2.5L1 9V11Z"
            stroke="currentColor"
            strokeWidth="1.2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </span>
    </div>
  );
}

/* -- Desktop Table (Flexbox Rows) -- */
function DesktopTable({
  instances,
  onRefresh,
  terminatingId,
  setTerminatingId,
}: {
  instances: InstanceResponse[];
  onRefresh?: () => void;
  terminatingId: string | null;
  setTerminatingId: (id: string | null) => void;
}) {
  async function handleRename(id: string, name: string) {
    await renameInstance(id, name);
    onRefresh?.();
  }

  return (
    <div className="hidden md:block">
      {/* Header row */}
      <div className="flex items-center gap-4 px-4 py-2">
        <div className="flex-1 min-w-[180px] max-w-[240px] type-ui-2xs text-text-dim font-medium uppercase tracking-wider">Name</div>
        <div className="flex-none w-[180px] type-ui-2xs text-text-dim font-medium uppercase tracking-wider">Status</div>
        <div className="flex-none w-[120px] type-ui-2xs text-text-dim font-medium uppercase tracking-wider">Region</div>
        <div className="flex-none w-[100px] type-ui-2xs text-text-dim font-medium uppercase tracking-wider">Cost</div>
        <div className="flex-1 min-w-0 type-ui-2xs text-text-dim font-medium uppercase tracking-wider">SSH Command</div>
      </div>

      {/* Data rows */}
      {instances.map((instance) => (
        <Link
          key={instance.id}
          href={`/cloud/instances/${instance.id}`}
          className="flex items-center gap-4 px-4 py-2 border-t border-border hover:bg-bg-card-hover transition-colors"
        >
          {/* Name + GPU type */}
          <div className="flex-1 min-w-[180px] max-w-[240px]">
            <div className="flex flex-col">
              <EditableName
                instance={instance}
                onRename={handleRename}
              />
              <span className="type-ui-sm text-text-muted font-mono">
                {getDisplayName(instance.gpu_type)} x{instance.gpu_count}
              </span>
            </div>
          </div>
          {/* Status + Uptime */}
          <div className="flex-none w-[180px]">
            <div className="flex flex-col gap-0.5 -ml-2">
              <StatusBadge status={instance.status} />
              <span className="type-ui-sm text-text-dim font-mono pl-2">
                {formatUptime(instance)}
              </span>
            </div>
          </div>
          {/* Region */}
          <div className="flex-none w-[120px]">
            <span className="text-[14px] text-text-muted">
              {(() => { const r = getRegionDisplay(instance.region); return `${r.flag} ${r.label}`; })()}
            </span>
          </div>
          {/* Cost */}
          <div className="flex-none w-[100px]">
            <span className="text-[14px] text-text font-mono">
              ${instance.price_per_hour.toFixed(2)}/hr
            </span>
          </div>
          {/* SSH Command */}
          <div className="flex-1 min-w-0">
            {instance.connection ? (
              <div className="flex items-center gap-2 min-w-0">
                <code className="type-ui-sm text-text-muted font-mono bg-bg-card pl-0 pr-2 py-1 rounded truncate min-w-0">
                  {instance.connection.ssh_command}
                </code>
                <CopyButton
                  text={instance.connection.ssh_command}
                  onClick={(e) => {
                    e.stopPropagation();
                    e.preventDefault();
                  }}
                />
              </div>
            ) : (
              <span className="type-ui-sm text-text-dim">--</span>
            )}
          </div>
        </Link>
      ))}
    </div>
  );
}

/* -- Mobile Cards -- */
function MobileCards({
  instances,
  onRefresh,
  terminatingId,
  setTerminatingId,
}: {
  instances: InstanceResponse[];
  onRefresh?: () => void;
  terminatingId: string | null;
  setTerminatingId: (id: string | null) => void;
}) {
  async function handleRename(id: string, name: string) {
    await renameInstance(id, name);
    onRefresh?.();
  }

  return (
    <div className="md:hidden space-y-3">
      {instances.map((instance) => (
        <Link
          key={instance.id}
          href={`/cloud/instances/${instance.id}`}
          className="block bg-bg-card rounded-[10px] border border-border p-4 space-y-3 hover:border-border-light transition-colors"
        >
          <div className="flex items-center justify-between">
            <EditableName instance={instance} onRename={handleRename} />
            <StatusBadge status={instance.status} />
          </div>

          <div className="grid grid-cols-2 gap-2">
            <div>
              <p className="type-ui-2xs text-text-dim uppercase">GPU</p>
              <p className="type-ui-sm text-text font-mono">
                {getDisplayName(instance.gpu_type)} x{instance.gpu_count}
              </p>
            </div>
            <div>
              <p className="type-ui-2xs text-text-dim uppercase">Region</p>
              <p className="type-ui-sm text-text-muted">{(() => { const r = getRegionDisplay(instance.region); return `${r.flag} ${r.label}`; })()}</p>
            </div>
            <div>
              <p className="type-ui-2xs text-text-dim uppercase">Cost</p>
              <p className="type-ui-sm text-text font-mono">
                ${instance.price_per_hour.toFixed(2)}/hr
              </p>
            </div>
            <div>
              <p className="type-ui-2xs text-text-dim uppercase">Uptime</p>
              <p className="type-ui-sm text-text-muted font-mono">
                {formatUptime(instance)}
              </p>
            </div>
          </div>

          {instance.connection && (
            <div
              className="flex items-center gap-2 bg-bg rounded px-3 py-2 border border-border/50"
              onClick={(e) => e.stopPropagation()}
            >
              <code className="type-ui-2xs text-text-muted font-mono flex-1 truncate">
                {instance.connection.ssh_command}
              </code>
              <CopyButton text={instance.connection.ssh_command} />
            </div>
          )}

          {instance.error_reason && (
            <p className="type-ui-2xs text-red-400">
              {instance.error_reason}
            </p>
          )}

          {instance.status !== "terminated" &&
            instance.status !== "stopping" && (
              <div className="pt-1">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    e.preventDefault();
                    setTerminatingId(instance.id);
                  }}
                  className={cn(
                    "type-ui-2xs px-3 py-2.5 sm:px-2 sm:py-1 rounded border transition-colors font-medium",
                    "border-red-500/30 text-red-400 hover:bg-red-500/10 hover:border-red-500/50"
                  )}
                >
                  Terminate
                </button>
              </div>
            )}
        </Link>
      ))}
    </div>
  );
}

/* -- Main Component -- */
export function InstancesTable({
  instances,
  onRefresh,
  onLaunch,
}: {
  instances: InstanceResponse[];
  onRefresh?: () => void;
  onLaunch?: () => void;
}) {
  const [terminatingId, setTerminatingId] = useState<string | null>(null);
  const [terminateLoading, setTerminateLoading] = useState(false);
  const [selectedStatuses, setSelectedStatuses] = useState<Set<string>>(
    () => new Set(ALL_STATUSES.filter((s) => s !== "terminated"))
  );
  const [selectedRegions, setSelectedRegions] = useState<Set<string>>(() => new Set());
  const [searchQuery, setSearchQuery] = useState("");
  const [perPage, setPerPage] = useState(10);
  const [currentPage, setCurrentPage] = useState(0);

  const allRegions = useMemo(
    () => [...new Set(instances.map((i) => i.region))].sort(),
    [instances]
  );

  const filteredInstances = useMemo(() => {
    return instances.filter((i) => {
      if (!selectedStatuses.has(i.status)) return false;
      if (selectedRegions.size > 0 && !selectedRegions.has(i.region)) return false;
      if (searchQuery) {
        const q = searchQuery.toLowerCase();
        const name = (i.name ?? i.id.slice(0, 12)).toLowerCase();
        const gpu = getDisplayName(i.gpu_type).toLowerCase();
        if (!name.includes(q) && !gpu.includes(q)) return false;
      }
      return true;
    });
  }, [instances, selectedStatuses, selectedRegions, searchQuery]);

  const totalFiltered = filteredInstances.length;
  const totalPages = Math.max(1, Math.ceil(totalFiltered / perPage));
  const safePage = Math.min(currentPage, totalPages - 1);
  const paginatedInstances = filteredInstances.slice(safePage * perPage, (safePage + 1) * perPage);
  const rangeStart = totalFiltered === 0 ? 0 : safePage * perPage + 1;
  const rangeEnd = Math.min((safePage + 1) * perPage, totalFiltered);

  // Reset page when filters or perPage change
  useEffect(() => { setCurrentPage(0); }, [selectedStatuses, selectedRegions, searchQuery, perPage]);

  async function handleTerminate() {
    if (!terminatingId) return;
    setTerminateLoading(true);
    try {
      await terminateInstance(terminatingId);
      onRefresh?.();
    } catch {
      // Error visible via status change on refresh
    } finally {
      setTerminateLoading(false);
      setTerminatingId(null);
    }
  }

  if (instances.length === 0) {
    return (
      <EmptyState
        icon={
          <svg width="20" height="20" viewBox="0 0 16 16" fill="none">
            <rect
              x="1"
              y="2"
              width="14"
              height="4"
              rx="1"
              stroke="currentColor"
              strokeWidth="1.5"
            />
            <rect
              x="1"
              y="10"
              width="14"
              height="4"
              rx="1"
              stroke="currentColor"
              strokeWidth="1.5"
            />
          </svg>
        }
        title="No instances running"
        description="Launch your first GPU instance to get started."
        action={{ label: "Browse GPUs", href: "/cloud/gpu-availability" }}
      />
    );
  }

  return (
    <>
      {/* Filter row — column-aligned with table headers */}
      <div className="hidden md:flex items-center gap-4 px-4 py-2.5 border-b border-border">
        <div className="flex-1 min-w-[180px] max-w-[240px]">
          <div className="relative">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" className="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-dim pointer-events-none">
              <circle cx="6" cy="6" r="4.25" stroke="currentColor" strokeWidth="1.25" />
              <path d="M9.5 9.5L12.5 12.5" stroke="currentColor" strokeWidth="1.25" strokeLinecap="round" />
            </svg>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search instances..."
              className="w-full pl-[30px] pr-3 py-1.5 rounded-lg border border-border bg-bg-card/50 text-text type-ui-sm placeholder:text-text-dim focus:outline-none focus:border-border-light transition-colors"
            />
          </div>
        </div>
        <div className="flex-none w-[180px]">
          <StatusFilter selected={selectedStatuses} onChange={setSelectedStatuses} />
        </div>
        <div className="flex-none w-[120px]">
          <RegionFilter regions={allRegions} selected={selectedRegions} onChange={setSelectedRegions} />
        </div>
        <div className="flex-none w-[100px]" />
        <div className="flex-1 min-w-0 text-right">
          {onLaunch && (
            <button
              onClick={onLaunch}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-purple hover:bg-purple/80 text-white transition-colors type-ui-sm font-medium"
            >
              Launch Instance
            </button>
          )}
        </div>
      </div>
      {/* Mobile filter bar */}
      <div className="md:hidden px-4 py-2.5 border-b border-border space-y-2">
        <div className="relative">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none" className="absolute left-2 top-1/2 -translate-y-1/2 text-text-dim pointer-events-none">
            <circle cx="6" cy="6" r="4.25" stroke="currentColor" strokeWidth="1.25" />
            <path d="M9.5 9.5L12.5 12.5" stroke="currentColor" strokeWidth="1.25" strokeLinecap="round" />
          </svg>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search instances..."
            className="w-full pl-[30px] pr-3 py-1.5 rounded-lg border border-border bg-bg-card/50 text-text type-ui-sm placeholder:text-text-dim focus:outline-none focus:border-border-light transition-colors"
          />
        </div>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <StatusFilter selected={selectedStatuses} onChange={setSelectedStatuses} />
            <RegionFilter regions={allRegions} selected={selectedRegions} onChange={setSelectedRegions} />
          </div>
          {onLaunch && (
            <button
              onClick={onLaunch}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-purple hover:bg-purple/80 text-white transition-colors type-ui-sm font-medium"
            >
              Launch Instance
            </button>
          )}
        </div>
      </div>

      <DesktopTable
        instances={paginatedInstances}
        onRefresh={onRefresh}
        terminatingId={terminatingId}
        setTerminatingId={setTerminatingId}
      />
      <MobileCards
        instances={paginatedInstances}
        onRefresh={onRefresh}
        terminatingId={terminatingId}
        setTerminatingId={setTerminatingId}
      />

      {totalFiltered > 0 && (
        <div className="flex items-center justify-between px-4 py-2.5 border-t border-border">
          {/* Per-page toggler */}
          <div className="flex items-center gap-2">
            <span className="type-ui-xs text-text-dim">Show</span>
            <div className="inline-flex rounded-md border border-border overflow-hidden">
              {([10, 25, 50] as const).map((n) => (
                <button
                  key={n}
                  onClick={() => setPerPage(n)}
                  className={cn(
                    "px-2.5 py-1 type-ui-xs font-medium transition-colors",
                    "border-r border-border last:border-r-0",
                    perPage === n
                      ? "bg-bg-card-hover text-text"
                      : "text-text-dim hover:text-text hover:bg-bg-card-hover/50"
                  )}
                >
                  {n}
                </button>
              ))}
            </div>
          </div>

          {/* Page navigation */}
          <div className="flex items-center gap-3">
            <span className="type-ui-xs text-text-dim">
              Page {safePage + 1} of {totalPages}
            </span>
            <div className="flex items-center gap-1">
              <button
                onClick={() => setCurrentPage((p) => Math.max(0, p - 1))}
                disabled={safePage === 0}
                className={cn(
                  "inline-flex items-center justify-center w-10 h-10 sm:w-7 sm:h-7 rounded transition-colors",
                  safePage === 0
                    ? "text-text-dim/30 cursor-not-allowed"
                    : "text-text-dim hover:text-text hover:bg-bg-card-hover"
                )}
                aria-label="Previous page"
              >
                <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                  <path d="M8.5 3.5L5 7l3.5 3.5" stroke="currentColor" strokeWidth="1.25" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </button>
              <button
                onClick={() => setCurrentPage((p) => Math.min(totalPages - 1, p + 1))}
                disabled={safePage >= totalPages - 1}
                className={cn(
                  "inline-flex items-center justify-center w-10 h-10 sm:w-7 sm:h-7 rounded transition-colors",
                  safePage >= totalPages - 1
                    ? "text-text-dim/30 cursor-not-allowed"
                    : "text-text-dim hover:text-text hover:bg-bg-card-hover"
                )}
                aria-label="Next page"
              >
                <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                  <path d="M5.5 3.5L9 7l-3.5 3.5" stroke="currentColor" strokeWidth="1.25" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      )}

      {filteredInstances.length === 0 && (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="type-ui-sm text-text-dim">No instances match the selected filters</p>
          <button
            onClick={() => {
              setSelectedStatuses(new Set(ALL_STATUSES));
              setSelectedRegions(new Set());
              setSearchQuery("");
            }}
            className="mt-2 type-ui-xs text-text-muted hover:text-text transition-colors underline underline-offset-2"
          >
            Clear all filters
          </button>
        </div>
      )}

      {/* Terminate confirmation dialog */}
      {terminatingId && (
        <ConfirmDialog
          title="Terminate Instance"
          message="This will permanently destroy the instance and stop billing. This action cannot be undone."
          confirmLabel="Terminate"
          confirmVariant="danger"
          onConfirm={handleTerminate}
          onCancel={() => setTerminatingId(null)}
          loading={terminateLoading}
        />
      )}
    </>
  );
}
