"use client";

import { useState } from "react";
import useSWR from "swr";
import { setSpendingLimit, deleteSpendingLimit } from "@/lib/api";
import { cn } from "@/lib/utils";
import { ConfirmDialog } from "@/components/cloud/ConfirmDialog";
import type { SpendingLimitResponse } from "@/lib/types";

function spendingLimitFetcher(url: string): Promise<SpendingLimitResponse | null> {
  return fetch(url).then((r) => {
    if (r.status === 404) return null;
    if (!r.ok) throw new Error(`API error: ${r.status}`);
    return r.json();
  });
}

export default function SettingsPage() {
  const { data: limitData, error, isLoading, mutate } = useSWR<SpendingLimitResponse | null>(
    "/api/v1/billing/spending-limit",
    spendingLimitFetcher,
  );

  const [limitInput, setLimitInput] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [showRemoveConfirm, setShowRemoveConfirm] = useState(false);
  const [removeLoading, setRemoveLoading] = useState(false);

  async function handleSetLimit(e: React.FormEvent) {
    e.preventDefault();
    const dollars = parseFloat(limitInput);
    if (isNaN(dollars) || dollars <= 0) {
      setSaveError("Please enter a valid amount greater than $0");
      return;
    }
    setSaveError(null);
    setSaving(true);
    try {
      await setSpendingLimit(dollars);
      await mutate();
      setLimitInput("");
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : "Failed to set limit");
    } finally {
      setSaving(false);
    }
  }

  async function handleRemoveLimit() {
    setRemoveLoading(true);
    try {
      await deleteSpendingLimit();
      await mutate();
      setShowRemoveConfirm(false);
    } catch {
      // Will reflect on next fetch
      setShowRemoveConfirm(false);
    } finally {
      setRemoveLoading(false);
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="type-h3 text-text">Settings</h1>
      </div>

      {/* Spending Limit Section */}
      <div className="bg-bg-card border border-border rounded-xl overflow-hidden">
        <div className="px-6 py-4 border-b border-border/60">
          <h2 className="type-ui-sm text-text font-medium">
            Monthly Spending Limit
          </h2>
          <p className="type-ui-2xs text-text-dim mt-1">
            Set a maximum monthly spend. Running instances will be stopped when
            the limit is reached.
          </p>
        </div>

        <div className="p-6 space-y-5">
          {error ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <p className="type-ui-sm text-red-400">
                Failed to load spending limit
              </p>
              <button
                onClick={() => mutate()}
                className="mt-3 type-ui-xs text-purple hover:text-purple-light transition-colors"
              >
                Retry
              </button>
            </div>
          ) : isLoading ? (
            <div className="space-y-3">
              <div className="h-6 bg-bg-card-hover rounded animate-pulse w-48" />
              <div className="h-4 bg-bg-card-hover rounded animate-pulse w-32" />
            </div>
          ) : limitData ? (
            <>
              {/* Current limit display */}
              <div className="bg-bg rounded-lg border border-border/60 p-5">
                <div className="flex items-start justify-between">
                  <div>
                    <p className="type-ui-2xs text-text-dim font-medium uppercase tracking-wider mb-1">
                      Current Limit
                    </p>
                    <p className="type-h3 font-mono text-text">
                      <span className="text-text-muted">$</span>
                      {limitData.monthly_limit_dollars.toFixed(2)}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="type-ui-2xs text-text-dim font-medium uppercase tracking-wider mb-1">
                      Current Spend
                    </p>
                    <p className="type-ui-sm font-mono text-text">
                      <span className="text-text-muted">$</span>
                      {limitData.current_month_spend_dollars.toFixed(2)}
                    </p>
                  </div>
                </div>

                {/* Progress bar */}
                <div className="mt-4">
                  <div className="w-full h-2 bg-bg-card rounded-full overflow-hidden">
                    <div
                      className={cn(
                        "h-full rounded-full transition-all",
                        limitData.percent_used > 90
                          ? "bg-red-500"
                          : limitData.percent_used > 70
                            ? "bg-yellow-500"
                            : "bg-purple"
                      )}
                      style={{
                        width: `${Math.min(limitData.percent_used, 100)}%`,
                      }}
                    />
                  </div>
                  <div className="flex justify-between mt-1.5">
                    <span className="type-ui-2xs text-text-dim">
                      {limitData.percent_used.toFixed(0)}% used
                    </span>
                    <span className="type-ui-2xs text-text-dim">
                      ${(limitData.monthly_limit_dollars - limitData.current_month_spend_dollars).toFixed(2)} remaining
                    </span>
                  </div>
                </div>

                {limitData.limit_reached_at && (
                  <div className="mt-3 bg-red-500/10 border border-red-500/30 rounded-lg px-3 py-2">
                    <p className="type-ui-xs text-red-400 font-medium">
                      Limit reached -- instances have been stopped
                    </p>
                  </div>
                )}
              </div>

              {/* Update form */}
              <form onSubmit={handleSetLimit} className="flex gap-3 items-end">
                <div className="flex-1">
                  <label className="type-ui-2xs text-text-dim font-medium block mb-1.5">
                    Update Limit
                  </label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 type-ui-sm text-text-dim">
                      $
                    </span>
                    <input
                      type="number"
                      step="0.01"
                      min="0.01"
                      value={limitInput}
                      onChange={(e) => setLimitInput(e.target.value)}
                      placeholder={limitData.monthly_limit_dollars.toFixed(2)}
                      className="w-full bg-bg border border-border rounded-lg pl-7 pr-4 py-2.5 type-ui-sm text-text font-mono placeholder:text-text-dim focus:outline-none focus:ring-2 focus:ring-purple/50 focus:border-purple/50 transition-all"
                    />
                  </div>
                </div>
                <button
                  type="submit"
                  disabled={saving || !limitInput}
                  className={cn(
                    "px-4 py-2.5 rounded-lg type-ui-sm font-medium transition-all whitespace-nowrap",
                    saving || !limitInput
                      ? "bg-purple/30 text-text-dim cursor-not-allowed"
                      : "gradient-btn"
                  )}
                >
                  {saving ? "Saving..." : "Update"}
                </button>
                <button
                  type="button"
                  onClick={() => setShowRemoveConfirm(true)}
                  className="px-4 py-2.5 rounded-lg type-ui-sm font-medium border border-red-500/30 text-red-400 hover:bg-red-500/10 hover:border-red-500/50 transition-colors whitespace-nowrap"
                >
                  Remove
                </button>
              </form>
            </>
          ) : (
            <>
              {/* No limit set */}
              <div className="bg-bg rounded-lg border border-border/60 p-5 text-center">
                <div className="w-10 h-10 rounded-full bg-bg-card flex items-center justify-center mx-auto mb-3">
                  <svg
                    width="18"
                    height="18"
                    viewBox="0 0 16 16"
                    fill="none"
                    className="text-text-dim"
                  >
                    <circle
                      cx="8"
                      cy="8"
                      r="6"
                      stroke="currentColor"
                      strokeWidth="1.5"
                    />
                    <path
                      d="M8 5v6M5 8h6"
                      stroke="currentColor"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                    />
                  </svg>
                </div>
                <p className="type-ui-sm text-text-muted">
                  No spending limit set
                </p>
                <p className="type-ui-2xs text-text-dim mt-1">
                  Set a limit to automatically stop instances when your monthly
                  spend exceeds it.
                </p>
              </div>

              {/* Set limit form */}
              <form onSubmit={handleSetLimit} className="flex gap-3 items-end">
                <div className="flex-1">
                  <label className="type-ui-2xs text-text-dim font-medium block mb-1.5">
                    Monthly Limit
                  </label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 -translate-y-1/2 type-ui-sm text-text-dim">
                      $
                    </span>
                    <input
                      type="number"
                      step="0.01"
                      min="0.01"
                      value={limitInput}
                      onChange={(e) => setLimitInput(e.target.value)}
                      placeholder="100.00"
                      className="w-full bg-bg border border-border rounded-lg pl-7 pr-4 py-2.5 type-ui-sm text-text font-mono placeholder:text-text-dim focus:outline-none focus:ring-2 focus:ring-purple/50 focus:border-purple/50 transition-all"
                    />
                  </div>
                </div>
                <button
                  type="submit"
                  disabled={saving || !limitInput}
                  className={cn(
                    "px-4 py-2.5 rounded-lg type-ui-sm font-medium transition-all whitespace-nowrap",
                    saving || !limitInput
                      ? "bg-purple/30 text-text-dim cursor-not-allowed"
                      : "gradient-btn"
                  )}
                >
                  {saving ? "Saving..." : "Set Limit"}
                </button>
              </form>
            </>
          )}

          {saveError && (
            <div className="bg-red-500/10 border border-red-500/30 rounded-lg px-4 py-3">
              <p className="type-ui-xs text-red-400">{saveError}</p>
            </div>
          )}

          {/* Warning */}
          <div className="bg-yellow-500/5 border border-yellow-500/20 rounded-lg px-4 py-3">
            <p className="type-ui-xs text-yellow-400/80">
              Instances will be stopped when the spending limit is reached.
              Stopped instances are automatically terminated after 72 hours.
            </p>
          </div>
        </div>
      </div>

      {/* Organization Section (placeholder) */}
      <div className="bg-bg-card border border-border rounded-xl overflow-hidden">
        <div className="px-6 py-4 border-b border-border/60">
          <h2 className="type-ui-sm text-text font-medium">Organization</h2>
        </div>
        <div className="p-6">
          <p className="type-ui-sm text-text-dim">
            Organization settings coming soon.
          </p>
        </div>
      </div>

      {/* Remove spending limit confirmation */}
      {showRemoveConfirm && (
        <ConfirmDialog
          title="Remove Spending Limit"
          message="Without a spending limit, instances will not be automatically stopped based on cost. You can always set a new limit later."
          confirmLabel="Remove Limit"
          confirmVariant="danger"
          onConfirm={handleRemoveLimit}
          onCancel={() => setShowRemoveConfirm(false)}
          loading={removeLoading}
        />
      )}
    </div>
  );
}
