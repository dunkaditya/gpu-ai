"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { useSWRConfig } from "swr";
import { cn } from "@/lib/utils";
import { StatusBadge } from "@/components/cloud/StatusBadge";
import { ConfirmDialog } from "@/components/cloud/ConfirmDialog";
import { renameInstance, terminateInstance } from "@/lib/api";
import type { InstanceResponse } from "@/lib/types";

/* ── Utility Functions ── */

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

function formatDateTime(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function calculateSessionCost(instance: InstanceResponse): string {
  if (instance.status === "terminated") return "--";
  const start = instance.ready_at ?? instance.created_at;
  if (!start) return "--";
  const elapsed = (Date.now() - new Date(start).getTime()) / 1000;
  if (elapsed < 0) return "$0.00";
  const cost = (instance.price_per_hour * elapsed) / 3600;
  return `$${cost.toFixed(2)}`;
}

function displayName(instance: InstanceResponse) {
  return instance.name ?? instance.id.slice(0, 12);
}

/* ── CopyButton ── */

function CopyButton({ text, label }: { text: string; label?: string }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
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
        "inline-flex items-center gap-1.5 px-3 py-2.5 sm:px-2 sm:py-1 rounded transition-colors type-ui-2xs font-medium",
        copied
          ? "text-green bg-green-dim"
          : "text-text-dim hover:text-text hover:bg-bg-card-hover"
      )}
      title={copied ? "Copied!" : "Copy to clipboard"}
    >
      {copied ? (
        <>
          <svg width="12" height="12" viewBox="0 0 14 14" fill="none">
            <path
              d="M2.5 7.5L5.5 10.5L11.5 4.5"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          {label ? "Copied!" : null}
        </>
      ) : (
        <>
          <svg width="12" height="12" viewBox="0 0 14 14" fill="none">
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
          {label ?? null}
        </>
      )}
    </button>
  );
}

/* ── Editable Name ── */

function EditableName({
  instance,
  onRename,
}: {
  instance: InstanceResponse;
  onRename: (name: string) => Promise<void>;
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
      await onRename(trimmed);
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
      <input
        ref={inputRef}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onBlur={save}
        onKeyDown={handleKeyDown}
        disabled={saving}
        className="type-h3 text-text font-medium bg-bg border border-purple/40 rounded-lg px-3 py-1 focus:outline-none focus:ring-2 focus:ring-purple/50 w-64"
      />
    );
  }

  return (
    <button
      onClick={() => setEditing(true)}
      className="group/name flex items-center gap-2"
    >
      <h1 className="type-h3 text-text font-medium">{displayName(instance)}</h1>
      <svg
        width="14"
        height="14"
        viewBox="0 0 12 12"
        fill="none"
        className="opacity-0 group-hover/name:opacity-60 transition-opacity shrink-0 text-text-dim"
      >
        <path
          d="M8.5 1.5L10.5 3.5M1 11H3L9.5 4.5L7.5 2.5L1 9V11Z"
          stroke="currentColor"
          strokeWidth="1.2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </button>
  );
}

/* ── Info Card ── */

function InfoCard({
  label,
  children,
  className,
}: {
  label: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "bg-bg-card border border-border rounded-[10px] p-5",
        className
      )}
    >
      <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mb-3">
        {label}
      </p>
      {children}
    </div>
  );
}

/* ── Main Component ── */

export function InstanceDetail({
  instance,
  onRefresh,
}: {
  instance: InstanceResponse;
  onRefresh: () => void;
}) {
  const { mutate: globalMutate } = useSWRConfig();
  const [showTerminate, setShowTerminate] = useState(false);
  const [terminateLoading, setTerminateLoading] = useState(false);

  const isTerminated = instance.status === "terminated";

  async function handleRename(name: string) {
    await renameInstance(instance.id, name);
    onRefresh();
    globalMutate("/api/v1/instances");
  }

  async function handleTerminate() {
    setTerminateLoading(true);
    try {
      await terminateInstance(instance.id);
      onRefresh();
      globalMutate("/api/v1/instances");
    } catch {
      // Error visible via status change
    } finally {
      setTerminateLoading(false);
      setShowTerminate(false);
    }
  }

  const tierLabel = instance.tier === "on_demand" ? "On-Demand" : "Spot";

  return (
    <div className="space-y-6">
      {/* Back link */}
      <Link
        href="/cloud/instances"
        className="inline-flex items-center gap-1.5 type-ui-sm text-text-dim hover:text-text transition-colors"
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 14 14"
          fill="none"
          className="shrink-0"
        >
          <path
            d="M8.5 3L4.5 7L8.5 11"
            stroke="currentColor"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
        Back to Instances
      </Link>

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-4">
        <EditableName instance={instance} onRename={handleRename} />
        <StatusBadge status={instance.status} />
        <div className="flex-1" />
        {!isTerminated && instance.status !== "stopping" && (
          <button
            onClick={() => setShowTerminate(true)}
            className="type-ui-sm px-4 py-2 rounded-lg border border-red-500/30 text-red-400 hover:bg-red-500/10 hover:border-red-500/50 transition-colors font-medium self-start"
          >
            Terminate
          </button>
        )}
      </div>

      {/* Error banner */}
      {instance.error_reason && (
        <div className="bg-red-500/10 border border-red-500/30 rounded-[10px] px-5 py-4">
          <p className="type-ui-xs text-red-400 font-medium uppercase tracking-wider mb-1">
            Error
          </p>
          <p className="type-ui-sm text-red-300">{instance.error_reason}</p>
        </div>
      )}

      {/* Info grid */}
      <div
        className={cn(
          "grid grid-cols-1 gap-4",
          isTerminated
            ? "md:grid-cols-2 lg:grid-cols-3"
            : "md:grid-cols-2 lg:grid-cols-3"
        )}
      >
        {/* GPU Configuration */}
        <InfoCard label="GPU Configuration">
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="type-ui-sm text-text-muted">Model</span>
              <span className="type-ui-sm text-text font-mono font-medium">
                {instance.gpu_type}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="type-ui-sm text-text-muted">Count</span>
              <span className="type-ui-sm text-text font-mono">
                x{instance.gpu_count}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="type-ui-sm text-text-muted">Tier</span>
              <span
                className={cn(
                  "type-ui-xs inline-flex items-center rounded-full px-2 py-0.5 font-medium",
                  instance.tier === "on_demand"
                    ? "bg-blue-500/10 text-blue-400"
                    : "bg-amber-400/10 text-amber-400"
                )}
              >
                {tierLabel}
              </span>
            </div>
          </div>
        </InfoCard>

        {/* Connection / SSH */}
        {!isTerminated && (
          <InfoCard label="Connection" className="md:col-span-1 lg:col-span-2">
            {instance.connection ? (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="type-ui-sm text-text-muted">Host</span>
                  <span className="type-ui-sm text-text font-mono">
                    {instance.connection.hostname}:{instance.connection.port}
                  </span>
                </div>
                <div>
                  <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mb-2">
                    SSH Command
                  </p>
                  <div className="bg-bg rounded-[10px] border border-border p-4 flex items-center gap-3">
                    <code className="type-ui-sm text-text font-mono flex-1 break-all">
                      {instance.connection.ssh_command}
                    </code>
                    <CopyButton
                      text={instance.connection.ssh_command}
                      label="Copy"
                    />
                  </div>
                </div>
              </div>
            ) : (
              <div className="flex items-center gap-2 text-text-dim">
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 14 14"
                  fill="none"
                  className="animate-spin"
                >
                  <circle
                    cx="7"
                    cy="7"
                    r="5.5"
                    stroke="currentColor"
                    strokeWidth="1.5"
                    strokeDasharray="20 12"
                  />
                </svg>
                <span className="type-ui-sm">
                  Waiting for connection details...
                </span>
              </div>
            )}
          </InfoCard>
        )}

        {/* Cost & Billing */}
        <InfoCard label="Cost & Billing">
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="type-ui-sm text-text-muted">Hourly Rate</span>
              <span className="type-ui-sm text-text font-mono font-medium">
                ${instance.price_per_hour.toFixed(2)}/hr
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="type-ui-sm text-text-muted">Uptime</span>
              <span className="type-ui-sm text-text font-mono">
                {formatUptime(instance)}
              </span>
            </div>
            {!isTerminated && (
              <div className="flex items-center justify-between pt-2 border-t border-border/50">
                <span className="type-ui-sm text-text-muted font-medium">
                  Est. Session Cost
                </span>
                <span className="type-ui-sm text-text font-mono font-medium">
                  {calculateSessionCost(instance)}
                </span>
              </div>
            )}
          </div>
        </InfoCard>

        {/* Metadata */}
        <InfoCard label="Metadata">
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="type-ui-sm text-text-muted">Region</span>
              <span className="type-ui-sm text-text">{instance.region}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="type-ui-sm text-text-muted">Created</span>
              <span className="type-ui-sm text-text">
                {formatDateTime(instance.created_at)}
              </span>
            </div>
            {instance.ready_at && (
              <div className="flex items-center justify-between">
                <span className="type-ui-sm text-text-muted">Ready</span>
                <span className="type-ui-sm text-text">
                  {formatDateTime(instance.ready_at)}
                </span>
              </div>
            )}
            {instance.terminated_at && (
              <div className="flex items-center justify-between">
                <span className="type-ui-sm text-text-muted">Terminated</span>
                <span className="type-ui-sm text-text">
                  {formatDateTime(instance.terminated_at)}
                </span>
              </div>
            )}
            <div className="flex items-center justify-between pt-2 border-t border-border/50">
              <span className="type-ui-sm text-text-muted">Instance ID</span>
              <div className="flex items-center gap-1">
                <span className="type-ui-2xs text-text-dim font-mono max-w-[120px] truncate">
                  {instance.id}
                </span>
                <CopyButton text={instance.id} />
              </div>
            </div>
          </div>
        </InfoCard>
      </div>

      {/* Terminated state CTA */}
      {isTerminated && (
        <div className="flex flex-col items-center justify-center py-8 text-center bg-bg-card border border-border rounded-[10px]">
          <p className="type-ui-sm text-text-dim mb-3">
            This instance has been terminated.
          </p>
          <Link
            href="/cloud/gpu-availability"
            className="btn-primary"
          >
            Launch New Instance
          </Link>
        </div>
      )}

      {/* Terminate confirmation dialog */}
      {showTerminate && (
        <ConfirmDialog
          title="Terminate Instance"
          message="This will permanently destroy the instance and stop billing. This action cannot be undone."
          confirmLabel="Terminate"
          confirmVariant="danger"
          onConfirm={handleTerminate}
          onCancel={() => setShowTerminate(false)}
          loading={terminateLoading}
        />
      )}
    </div>
  );
}
