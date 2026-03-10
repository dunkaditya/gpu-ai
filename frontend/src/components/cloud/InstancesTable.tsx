"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { cn } from "@/lib/utils";
import { StatusBadge } from "@/components/cloud/StatusBadge";
import { ConfirmDialog } from "@/components/cloud/ConfirmDialog";
import { terminateInstance, renameInstance } from "@/lib/api";
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

function displayName(instance: InstanceResponse) {
  return instance.name ?? instance.id.slice(0, 12);
}

/* ── CopyButton ── */

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
        "inline-flex items-center justify-center w-7 h-7 rounded transition-colors shrink-0",
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

/* ── EditableName ── */

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
          className="type-ui-sm text-text font-medium bg-bg border border-purple/40 rounded px-2 py-0.5 focus:outline-none focus:ring-1 focus:ring-purple/50 w-40"
        />
        {instance.name && (
          <span className="type-ui-2xs text-text-dim font-mono mt-0.5">
            {instance.id.slice(0, 12)}
          </span>
        )}
      </div>
    );
  }

  return (
    <div
      className="group/name flex flex-col cursor-pointer"
      onClick={(e) => {
        e.stopPropagation();
        e.preventDefault();
        setEditing(true);
      }}
    >
      <span className="type-ui-sm text-text font-medium inline-flex items-center gap-1.5">
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
      {instance.name && (
        <span className="type-ui-2xs text-text-dim font-mono">
          {instance.id.slice(0, 12)}
        </span>
      )}
    </div>
  );
}

/* ── Desktop Table ── */
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

  const thClass =
    "type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3";

  return (
    <div className="hidden md:block overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-border">
            <th className={thClass}>Name</th>
            <th className={thClass}>GPU</th>
            <th className={thClass}>Status</th>
            <th className={thClass}>Region</th>
            <th className={thClass}>Cost</th>
            <th className={thClass}>Uptime</th>
            <th className={thClass}>SSH Command</th>
            <th className={cn(thClass, "text-right")}>Actions</th>
          </tr>
        </thead>
        <tbody>
          {instances.map((instance) => (
            <tr key={instance.id} className="border-b border-border/50 group">
              <td colSpan={8} className="p-0">
                <Link
                  href={`/cloud/instances/${instance.id}`}
                  className="flex items-center hover:bg-bg-card transition-colors cursor-pointer"
                >
                  {/* Name */}
                  <div className="px-4 py-3 w-[180px] shrink-0">
                    <EditableName
                      instance={instance}
                      onRename={handleRename}
                    />
                  </div>
                  {/* GPU */}
                  <div className="px-4 py-3 shrink-0">
                    <span className="type-ui-sm text-text font-mono">
                      {instance.gpu_type} x{instance.gpu_count}
                    </span>
                  </div>
                  {/* Status */}
                  <div className="px-4 py-3 shrink-0">
                    <StatusBadge status={instance.status} />
                  </div>
                  {/* Region */}
                  <div className="px-4 py-3 shrink-0">
                    <span className="type-ui-sm text-text-muted">
                      {instance.region}
                    </span>
                  </div>
                  {/* Cost */}
                  <div className="px-4 py-3 shrink-0">
                    <span className="type-ui-sm text-text font-mono">
                      ${instance.price_per_hour.toFixed(2)}/hr
                    </span>
                  </div>
                  {/* Uptime */}
                  <div className="px-4 py-3 shrink-0">
                    <span className="type-ui-sm text-text-muted font-mono">
                      {formatUptime(instance)}
                    </span>
                  </div>
                  {/* SSH Command */}
                  <div className="px-4 py-3 flex-1 min-w-0">
                    {instance.connection ? (
                      <div className="flex items-center gap-2">
                        <code className="type-ui-2xs text-text-muted font-mono bg-bg-card px-2 py-1 rounded max-w-[260px] truncate">
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
                  {/* Actions */}
                  <div className="px-4 py-3 shrink-0 text-right">
                    {instance.status !== "terminated" &&
                      instance.status !== "stopping" && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            e.preventDefault();
                            setTerminatingId(instance.id);
                          }}
                          className={cn(
                            "type-ui-2xs px-2 py-1 rounded border transition-colors font-medium",
                            "border-red-500/30 text-red-400 hover:bg-red-500/10 hover:border-red-500/50"
                          )}
                        >
                          Terminate
                        </button>
                      )}
                  </div>
                </Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/* ── Mobile Cards ── */
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
          className="block bg-bg-card rounded-lg border border-border p-4 space-y-3 hover:border-border-light transition-colors"
        >
          <div className="flex items-center justify-between">
            <EditableName instance={instance} onRename={handleRename} />
            <StatusBadge status={instance.status} />
          </div>

          <div className="grid grid-cols-2 gap-2">
            <div>
              <p className="type-ui-2xs text-text-dim uppercase">GPU</p>
              <p className="type-ui-sm text-text font-mono">
                {instance.gpu_type} x{instance.gpu_count}
              </p>
            </div>
            <div>
              <p className="type-ui-2xs text-text-dim uppercase">Region</p>
              <p className="type-ui-sm text-text-muted">{instance.region}</p>
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
                    "type-ui-2xs px-2 py-1 rounded border transition-colors font-medium",
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

/* ── Main Component ── */
export function InstancesTable({
  instances,
  onRefresh,
}: {
  instances: InstanceResponse[];
  onRefresh?: () => void;
}) {
  const [terminatingId, setTerminatingId] = useState<string | null>(null);
  const [terminateLoading, setTerminateLoading] = useState(false);

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
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <div className="w-12 h-12 rounded-full bg-bg-card flex items-center justify-center mb-4">
          <svg width="20" height="20" viewBox="0 0 16 16" fill="none">
            <rect
              x="1"
              y="2"
              width="14"
              height="4"
              rx="1"
              stroke="currentColor"
              strokeWidth="1.5"
              className="text-text-dim"
            />
            <rect
              x="1"
              y="10"
              width="14"
              height="4"
              rx="1"
              stroke="currentColor"
              strokeWidth="1.5"
              className="text-text-dim"
            />
          </svg>
        </div>
        <p className="type-ui-sm text-text-muted">No instances running</p>
        <p className="type-ui-2xs text-text-dim mt-1 mb-4">
          Launch your first GPU instance to get started.
        </p>
        <Link
          href="/cloud/gpu-availability"
          className="gradient-btn px-4 py-2 rounded-lg type-ui-sm font-medium transition-all inline-block"
        >
          Browse GPU Availability
        </Link>
      </div>
    );
  }

  return (
    <>
      <DesktopTable
        instances={instances}
        onRefresh={onRefresh}
        terminatingId={terminatingId}
        setTerminatingId={setTerminatingId}
      />
      <MobileCards
        instances={instances}
        onRefresh={onRefresh}
        terminatingId={terminatingId}
        setTerminatingId={setTerminatingId}
      />

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
