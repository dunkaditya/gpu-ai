"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import { StatusBadge } from "@/components/cloud/StatusBadge";
import type { MockInstance } from "@/lib/mock-data";

function CopyButton({ text }: { text: string }) {
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

function formatTier(tier: MockInstance["tier"]) {
  return tier === "on_demand" ? "On-Demand" : "Spot";
}

function displayName(instance: MockInstance) {
  return instance.name ?? instance.id.slice(0, 12);
}

/* ── Desktop Table ── */
function DesktopTable({ instances }: { instances: MockInstance[] }) {
  return (
    <div className="hidden md:block overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-border">
            <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
              Name
            </th>
            <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
              GPU
            </th>
            <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
              Status
            </th>
            <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
              Region
            </th>
            <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
              Tier
            </th>
            <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
              Cost
            </th>
            <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
              SSH Command
            </th>
          </tr>
        </thead>
        <tbody>
          {instances.map((instance) => (
            <tr
              key={instance.id}
              className="border-b border-border/50 hover:bg-bg-card transition-colors"
            >
              <td className="px-4 py-3">
                <div className="flex flex-col">
                  <span className="type-ui-sm text-text font-medium">
                    {displayName(instance)}
                  </span>
                  {instance.name && (
                    <span className="type-ui-2xs text-text-dim font-mono">
                      {instance.id}
                    </span>
                  )}
                </div>
              </td>
              <td className="px-4 py-3">
                <span className="type-ui-sm text-text font-mono">
                  {instance.gpu_type} x{instance.gpu_count}
                </span>
              </td>
              <td className="px-4 py-3">
                <StatusBadge status={instance.status} />
              </td>
              <td className="px-4 py-3">
                <span className="type-ui-sm text-text-muted">
                  {instance.region}
                </span>
              </td>
              <td className="px-4 py-3">
                <span className="type-ui-sm text-text-muted">
                  {formatTier(instance.tier)}
                </span>
              </td>
              <td className="px-4 py-3">
                <span className="type-ui-sm text-text font-mono">
                  ${instance.price_per_hour.toFixed(2)}/hr
                </span>
              </td>
              <td className="px-4 py-3">
                {instance.connection ? (
                  <div className="flex items-center gap-2">
                    <code className="type-ui-2xs text-text-muted font-mono bg-bg-card px-2 py-1 rounded max-w-[260px] truncate">
                      {instance.connection.ssh_command}
                    </code>
                    <CopyButton text={instance.connection.ssh_command} />
                  </div>
                ) : (
                  <span className="type-ui-sm text-text-dim">--</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/* ── Mobile Cards ── */
function MobileCards({ instances }: { instances: MockInstance[] }) {
  return (
    <div className="md:hidden space-y-3">
      {instances.map((instance) => (
        <div
          key={instance.id}
          className="bg-bg-card rounded-lg border border-border p-4 space-y-3"
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="type-ui-sm text-text font-medium">
                {displayName(instance)}
              </p>
              {instance.name && (
                <p className="type-ui-2xs text-text-dim font-mono">
                  {instance.id}
                </p>
              )}
            </div>
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
              <p className="type-ui-2xs text-text-dim uppercase">Tier</p>
              <p className="type-ui-sm text-text-muted">
                {formatTier(instance.tier)}
              </p>
            </div>
            <div>
              <p className="type-ui-2xs text-text-dim uppercase">Cost</p>
              <p className="type-ui-sm text-text font-mono">
                ${instance.price_per_hour.toFixed(2)}/hr
              </p>
            </div>
          </div>

          {instance.connection && (
            <div className="flex items-center gap-2 bg-bg rounded px-3 py-2 border border-border/50">
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
        </div>
      ))}
    </div>
  );
}

/* ── Main Component ── */
export function InstancesTable({ instances }: { instances: MockInstance[] }) {
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
        <p className="type-ui-sm text-text-muted">No instances</p>
        <p className="type-ui-2xs text-text-dim mt-1">
          Launch your first GPU instance to get started.
        </p>
      </div>
    );
  }

  return (
    <>
      <DesktopTable instances={instances} />
      <MobileCards instances={instances} />
    </>
  );
}
