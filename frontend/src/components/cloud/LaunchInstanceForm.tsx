"use client";

import { useEffect, useCallback, useState } from "react";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { createInstance } from "@/lib/api";
import type { CreateInstanceRequest, AvailableOffering } from "@/lib/types";

interface LaunchInstanceFormProps {
  onClose: () => void;
  onSuccess: () => void;
  defaultGPU?: string;
  defaultRegion?: string;
  offering?: AvailableOffering;
}

export function LaunchInstanceForm({
  onClose,
  onSuccess,
  defaultGPU,
  defaultRegion,
  offering,
}: LaunchInstanceFormProps) {
  const router = useRouter();

  // Manual mode state
  const [gpuType, setGpuType] = useState(defaultGPU ?? "");
  const [region, setRegion] = useState(defaultRegion ?? "");
  const [tier, setTier] = useState<"spot" | "on_demand">(
    offering?.tier ?? "spot"
  );

  // Shared state
  const [gpuCount, setGpuCount] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isPreFilled = !!offering;
  const hourlyPrice = offering
    ? offering.price_per_hour * gpuCount
    : undefined;
  const monthlyEstimate = hourlyPrice ? hourlyPrice * 730 : undefined;

  // Escape key closes modal
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === "Escape" && !loading) {
        onClose();
      }
    },
    [onClose, loading]
  );

  useEffect(() => {
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    const req: CreateInstanceRequest = isPreFilled
      ? {
          gpu_type: offering.gpu_model,
          gpu_count: gpuCount,
          region: offering.region,
          tier: offering.tier,
        }
      : {
          gpu_type: gpuType,
          gpu_count: gpuCount,
          region,
          tier,
        };

    try {
      await createInstance(req);
      onSuccess();
      onClose();
      router.push("/cloud/instances");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to launch instance"
      );
    } finally {
      setLoading(false);
    }
  }

  const canSubmit = isPreFilled ? true : gpuType && region;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-bg/80 backdrop-blur-sm"
        onClick={loading ? undefined : onClose}
      />

      {/* Modal */}
      <div className="relative w-full max-w-lg mx-4 bg-bg-card border border-border rounded-xl shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <h2 className="type-h4 font-sans text-text">Launch Instance</h2>
          <button
            onClick={onClose}
            disabled={loading}
            className="text-text-dim hover:text-text transition-colors p-1 disabled:opacity-50"
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <path
                d="M4 4L12 12M12 4L4 12"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
              />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {isPreFilled ? (
            /* ---- Pre-filled mode: confirmation card ---- */
            <>
              {/* GPU model name */}
              <div className="text-center mb-2">
                <h3 className="type-h4 font-sans text-text">
                  {offering.gpu_model}
                </h3>
              </div>

              {/* Specs grid */}
              <div className="grid grid-cols-2 gap-3 bg-bg rounded-lg border border-border p-4">
                <div>
                  <span className="type-ui-2xs text-text-dim uppercase block">
                    VRAM
                  </span>
                  <span className="type-ui-sm text-text font-mono">
                    {offering.vram_gb} GB
                  </span>
                </div>
                <div>
                  <span className="type-ui-2xs text-text-dim uppercase block">
                    CPU Cores
                  </span>
                  <span className="type-ui-sm text-text font-mono">
                    {offering.cpu_cores}
                  </span>
                </div>
                <div>
                  <span className="type-ui-2xs text-text-dim uppercase block">
                    RAM
                  </span>
                  <span className="type-ui-sm text-text font-mono">
                    {offering.ram_gb} GB
                  </span>
                </div>
                <div>
                  <span className="type-ui-2xs text-text-dim uppercase block">
                    Storage
                  </span>
                  <span className="type-ui-sm text-text font-mono">
                    {offering.storage_gb} GB
                  </span>
                </div>
              </div>

              {/* Region + Tier */}
              <div className="flex items-center gap-3">
                <span className="type-ui-2xs bg-bg-card-hover text-text-muted rounded-full px-2.5 py-0.5">
                  {offering.region}
                </span>
                <span
                  className={cn(
                    "type-ui-2xs rounded-full px-2.5 py-0.5 font-medium",
                    offering.tier === "spot"
                      ? "bg-purple-dim text-purple-light"
                      : "bg-bg-card-hover text-text-muted"
                  )}
                >
                  {offering.tier === "on_demand" ? "On-Demand" : "Spot"}
                </span>
              </div>

              {/* Price confirmation */}
              <div className="text-center py-3 border-t border-b border-border/50">
                <span className="type-h3 font-mono text-purple-light">
                  ${hourlyPrice?.toFixed(2)}/hr
                </span>
                <p className="type-ui-2xs text-text-dim mt-1">
                  Estimated monthly cost: ~$
                  {monthlyEstimate?.toFixed(2)}
                </p>
              </div>

              {/* GPU count selector */}
              <div className="space-y-2">
                <label className="type-ui-xs text-text-muted font-medium uppercase tracking-wider">
                  GPU Count
                </label>
                <input
                  type="number"
                  min={1}
                  max={8}
                  value={gpuCount}
                  onChange={(e) =>
                    setGpuCount(
                      Math.max(1, Math.min(8, Number(e.target.value)))
                    )
                  }
                  className="w-full bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text font-mono focus:outline-none focus:ring-2 focus:ring-purple/50 focus:border-purple/50 transition-all"
                />
              </div>

              {/* Auto-SSH note */}
              <p className="type-ui-2xs text-text-dim text-center">
                All your SSH keys will be automatically attached
              </p>
            </>
          ) : (
            /* ---- Manual mode: free-text inputs ---- */
            <>
              {/* GPU Type */}
              <div className="space-y-2">
                <label className="type-ui-xs text-text-muted font-medium uppercase tracking-wider">
                  GPU Type
                </label>
                <input
                  type="text"
                  value={gpuType}
                  onChange={(e) => setGpuType(e.target.value)}
                  placeholder="e.g. H100 SXM, A100 SXM, RTX 4090"
                  required
                  className="w-full bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text font-mono placeholder:text-text-dim focus:outline-none focus:ring-2 focus:ring-purple/50 focus:border-purple/50 transition-all"
                />
              </div>

              {/* GPU Count + Region */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="type-ui-xs text-text-muted font-medium uppercase tracking-wider">
                    GPU Count
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={8}
                    value={gpuCount}
                    onChange={(e) => setGpuCount(Number(e.target.value))}
                    className="w-full bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text font-mono focus:outline-none focus:ring-2 focus:ring-purple/50 focus:border-purple/50 transition-all"
                  />
                </div>
                <div className="space-y-2">
                  <label className="type-ui-xs text-text-muted font-medium uppercase tracking-wider">
                    Region
                  </label>
                  <input
                    type="text"
                    value={region}
                    onChange={(e) => setRegion(e.target.value)}
                    placeholder="e.g. us-east"
                    required
                    className="w-full bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text font-mono placeholder:text-text-dim focus:outline-none focus:ring-2 focus:ring-purple/50 focus:border-purple/50 transition-all"
                  />
                </div>
              </div>

              {/* Tier Toggle */}
              <div className="space-y-2">
                <label className="type-ui-xs text-text-muted font-medium uppercase tracking-wider">
                  Tier
                </label>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => setTier("spot")}
                    className={cn(
                      "flex-1 py-2.5 rounded-lg type-ui-sm font-medium border transition-all",
                      tier === "spot"
                        ? "bg-purple-dim border-purple/40 text-purple-light"
                        : "bg-bg border-border text-text-muted hover:text-text hover:border-border-light"
                    )}
                  >
                    Spot
                  </button>
                  <button
                    type="button"
                    onClick={() => setTier("on_demand")}
                    className={cn(
                      "flex-1 py-2.5 rounded-lg type-ui-sm font-medium border transition-all",
                      tier === "on_demand"
                        ? "bg-purple-dim border-purple/40 text-purple-light"
                        : "bg-bg border-border text-text-muted hover:text-text hover:border-border-light"
                    )}
                  >
                    On-Demand
                  </button>
                </div>
              </div>
            </>
          )}

          {/* Error */}
          {error && (
            <div className="bg-red-500/10 border border-red-500/30 rounded-lg px-4 py-3">
              <p className="type-ui-xs text-red-400">{error}</p>
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="flex-1 py-2.5 rounded-lg type-ui-sm font-medium border border-border text-text-muted hover:text-text hover:bg-bg-card-hover transition-all disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !canSubmit}
              className={cn(
                "flex-1 py-2.5 rounded-lg type-ui-sm font-medium transition-all",
                loading || !canSubmit
                  ? "bg-purple/30 text-text-dim cursor-not-allowed"
                  : "gradient-btn"
              )}
            >
              {loading
                ? "Launching..."
                : isPreFilled
                  ? `Launch Instance - $${hourlyPrice?.toFixed(2)}/hr`
                  : "Launch Instance"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
