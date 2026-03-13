"use client";

import { useEffect, useCallback, useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { createInstance } from "@/lib/api";
import { getDisplayName } from "@/lib/gpu-categories";
import { getRegionDisplay } from "@/lib/regions";
import type { CreateInstanceRequest, AvailableOffering } from "@/lib/types";

interface LaunchInstanceFormProps {
  onClose: () => void;
  onSuccess: () => void;
  defaultGPU?: string;
  defaultRegion?: string;
  offerings?: AvailableOffering[];
}

export function LaunchInstanceForm({
  onClose,
  onSuccess,
  defaultGPU,
  defaultRegion,
  offerings,
}: LaunchInstanceFormProps) {
  const router = useRouter();

  // Available regions from offerings
  const availableRegions = useMemo(() => {
    if (!offerings) return [];
    const regionSet = new Set<string>();
    for (const o of offerings) {
      if (o.available_count > 0 && o.tier === "on_demand") {
        regionSet.add(o.region);
      }
    }
    return [...regionSet].sort();
  }, [offerings]);

  // Region selection state
  const initialRegion =
    defaultRegion && availableRegions.includes(defaultRegion)
      ? defaultRegion
      : availableRegions[0] ?? "";
  const [selectedRegion, setSelectedRegion] = useState(initialRegion);

  // Derive active offering from selected region (cheapest on-demand in that region)
  const activeOffering = useMemo(() => {
    if (!offerings) return undefined;
    const inRegion = offerings.filter(
      (o) =>
        o.region === selectedRegion &&
        o.available_count > 0 &&
        o.tier === "on_demand"
    );
    if (inRegion.length === 0) return undefined;
    return inRegion.reduce((a, b) =>
      a.price_per_hour < b.price_per_hour ? a : b
    );
  }, [offerings, selectedRegion]);

  // Manual mode state
  const [gpuType, setGpuType] = useState(defaultGPU ?? "");
  const [region, setRegion] = useState(defaultRegion ?? "");
  // Shared state
  const [gpuCount, setGpuCount] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isPreFilled = !!offerings && offerings.length > 0;
  const hourlyPrice = activeOffering
    ? activeOffering.price_per_hour * gpuCount
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

    const req: CreateInstanceRequest =
      isPreFilled && activeOffering
        ? {
            gpu_type: activeOffering.gpu_model,
            gpu_count: gpuCount,
            region: activeOffering.region,
            tier: "on_demand",
          }
        : {
            gpu_type: gpuType,
            gpu_count: gpuCount,
            region,
            tier: "on_demand",
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

  const canSubmit = isPreFilled ? !!activeOffering : gpuType && region;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-bg/80 backdrop-blur-sm"
        onClick={loading ? undefined : onClose}
      />

      {/* Modal */}
      <div className="relative w-full max-w-lg mx-4 bg-bg-card border border-border rounded-[10px] shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <h2 className="type-ui-sm font-medium text-text">Launch Instance</h2>
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
          {isPreFilled && activeOffering ? (
            /* ---- Pre-filled mode: confirmation card ---- */
            <>
              {/* GPU model name */}
              <div className="text-center mb-2">
                <h3 className="type-h4 font-sans text-text">
                  {getDisplayName(activeOffering.gpu_model)}
                </h3>
              </div>

              {/* Specs grid */}
              <div className="grid grid-cols-2 gap-3 bg-bg rounded-[10px] border border-border p-4">
                <div>
                  <span className="type-ui-2xs text-text-dim uppercase block">
                    VRAM
                  </span>
                  <span className="type-ui-sm text-text font-mono">
                    {activeOffering.vram_gb} GB
                  </span>
                </div>
                <div>
                  <span className="type-ui-2xs text-text-dim uppercase block">
                    CPU Cores
                  </span>
                  <span className="type-ui-sm text-text font-mono">
                    {activeOffering.cpu_cores}
                  </span>
                </div>
                <div>
                  <span className="type-ui-2xs text-text-dim uppercase block">
                    RAM
                  </span>
                  <span className="type-ui-sm text-text font-mono">
                    {activeOffering.ram_gb} GB
                  </span>
                </div>
                <div>
                  <span className="type-ui-2xs text-text-dim uppercase block">
                    Storage
                  </span>
                  <span className="type-ui-sm text-text font-mono">
                    {activeOffering.storage_gb} GB
                  </span>
                </div>
              </div>

              {/* Region selector */}
              {availableRegions.length > 1 ? (
                <div className="space-y-2">
                  <label className="type-ui-xs text-text-muted font-medium uppercase tracking-wider">
                    Region
                  </label>
                  <div className="relative">
                    <select
                      value={selectedRegion}
                      onChange={(e) => setSelectedRegion(e.target.value)}
                      className="w-full appearance-none bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text focus:outline-none focus:ring-1 focus:ring-border-light focus:border-border-light transition-all cursor-pointer"
                    >
                      {availableRegions.map((r) => {
                        const display = getRegionDisplay(r);
                        const offeringsInRegion = offerings!.filter(
                          (o) =>
                            o.region === r &&
                            o.available_count > 0 &&
                            o.tier === "on_demand"
                        );
                        const cheapest =
                          offeringsInRegion.length > 0
                            ? Math.min(
                                ...offeringsInRegion.map((o) => o.price_per_hour)
                              )
                            : 0;
                        return (
                          <option key={r} value={r}>
                            {display.flag} {display.label} — ${cheapest.toFixed(2)}/hr
                          </option>
                        );
                      })}
                    </select>
                    {/* Chevron icon */}
                    <svg
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-text-dim pointer-events-none"
                      width="14"
                      height="14"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    >
                      <path d="M6 9l6 6 6-6" />
                    </svg>
                  </div>
                </div>
              ) : (
                <div className="flex items-center gap-3">
                  <span className="type-ui-2xs bg-bg-card-hover text-text-muted rounded-full px-2.5 py-0.5">
                    {(() => { const r = getRegionDisplay(selectedRegion); return `${r.flag} ${r.label}`; })()}
                  </span>
                  <span className="type-ui-2xs bg-bg-card-hover text-text-muted rounded-full px-2.5 py-0.5">
                    On-Demand
                  </span>
                </div>
              )}

              {/* Price confirmation */}
              <div className="text-center py-3 border-t border-b border-border/50">
                <span className="type-h4 font-mono text-text">
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
                  className="w-full bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text font-mono focus:outline-none focus:ring-1 focus:ring-border-light focus:border-border-light transition-all"
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
                  className="w-full bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text font-mono placeholder:text-text-dim focus:outline-none focus:ring-1 focus:ring-border-light focus:border-border-light transition-all"
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
                    className="w-full bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text font-mono focus:outline-none focus:ring-1 focus:ring-border-light focus:border-border-light transition-all"
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
                    className="w-full bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text font-mono placeholder:text-text-dim focus:outline-none focus:ring-1 focus:ring-border-light focus:border-border-light transition-all"
                  />
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
              className="btn-secondary flex-1 py-2.5 rounded-lg type-ui-sm font-medium"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !canSubmit}
              className={cn(
                "flex-1 py-2.5 rounded-lg type-ui-sm font-medium transition-all",
                loading || !canSubmit
                  ? "bg-bg-card-hover text-text-dim cursor-not-allowed"
                  : "btn-primary"
              )}
            >
              {loading
                ? "Launching..."
                : isPreFilled && hourlyPrice
                  ? `Launch Instance - $${hourlyPrice.toFixed(2)}/hr`
                  : "Launch Instance"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
