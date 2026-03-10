"use client";

import { useState } from "react";
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
  const [gpuType, setGpuType] = useState(defaultGPU ?? "");
  const [gpuCount, setGpuCount] = useState(1);
  const [region, setRegion] = useState(defaultRegion ?? "");
  const [tier, setTier] = useState<"spot" | "on_demand">("spot");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    const req: CreateInstanceRequest = {
      gpu_type: gpuType,
      gpu_count: gpuCount,
      region,
      tier,
    };

    try {
      await createInstance(req);
      onSuccess();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to launch instance");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-bg/80 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative w-full max-w-lg mx-4 bg-bg-card border border-border rounded-xl shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <h2 className="type-h4 font-sans text-text">Launch Instance</h2>
          <button
            onClick={onClose}
            className="text-text-dim hover:text-text transition-colors p-1"
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

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-5">
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
              className="flex-1 py-2.5 rounded-lg type-ui-sm font-medium border border-border text-text-muted hover:text-text hover:bg-bg-card-hover transition-all"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !gpuType || !region}
              className={cn(
                "flex-1 py-2.5 rounded-lg type-ui-sm font-medium transition-all",
                loading || !gpuType || !region
                  ? "bg-purple/30 text-text-dim cursor-not-allowed"
                  : "gradient-btn"
              )}
            >
              {loading ? "Launching..." : "Launch Instance"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
