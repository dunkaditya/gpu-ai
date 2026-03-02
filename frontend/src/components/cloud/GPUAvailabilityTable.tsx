"use client";

import { useMemo, useState } from "react";
import useSWR from "swr";
import { cn } from "@/lib/utils";
import { fetcher } from "@/lib/api";
import { LaunchInstanceForm } from "@/components/cloud/LaunchInstanceForm";
import type { AvailableOffering } from "@/lib/types";

function SkeletonRow() {
  return (
    <tr className="border-b border-border/50">
      {Array.from({ length: 7 }).map((_, i) => (
        <td key={i} className="px-4 py-3">
          <div className="h-4 bg-bg-card-hover rounded animate-pulse w-20" />
        </td>
      ))}
    </tr>
  );
}

export function GPUAvailabilityTable() {
  const { data, error, isLoading, mutate } = useSWR<{
    available: AvailableOffering[];
  }>("/api/v1/gpu/available", fetcher, { refreshInterval: 30000 });

  const [gpuFilter, setGpuFilter] = useState("");
  const [regionFilter, setRegionFilter] = useState("");
  const [tierFilter, setTierFilter] = useState("");
  const [sortField, setSortField] = useState<"price" | "vram" | "available">(
    "price"
  );
  const [sortDir, setSortDir] = useState<"asc" | "desc">("asc");

  // Launch form state
  const [launchGPU, setLaunchGPU] = useState<string | undefined>();
  const [launchRegion, setLaunchRegion] = useState<string | undefined>();
  const [showLaunch, setShowLaunch] = useState(false);

  const offerings = data?.available ?? [];

  // Extract unique filter values from data
  const gpuModels = useMemo(
    () => [...new Set(offerings.map((o) => o.gpu_model))].sort(),
    [offerings]
  );
  const regions = useMemo(
    () => [...new Set(offerings.map((o) => o.region))].sort(),
    [offerings]
  );

  // Filtered and sorted
  const filtered = useMemo(() => {
    let result = offerings;
    if (gpuFilter) result = result.filter((o) => o.gpu_model === gpuFilter);
    if (regionFilter) result = result.filter((o) => o.region === regionFilter);
    if (tierFilter) result = result.filter((o) => o.tier === tierFilter);

    return [...result].sort((a, b) => {
      const mul = sortDir === "asc" ? 1 : -1;
      if (sortField === "price")
        return (a.price_per_hour - b.price_per_hour) * mul;
      if (sortField === "vram") return (a.vram_gb - b.vram_gb) * mul;
      return (a.available_count - b.available_count) * mul;
    });
  }, [offerings, gpuFilter, regionFilter, tierFilter, sortField, sortDir]);

  function toggleSort(field: "price" | "vram" | "available") {
    if (sortField === field) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortField(field);
      setSortDir("asc");
    }
  }

  function SortIcon({ field }: { field: "price" | "vram" | "available" }) {
    if (sortField !== field)
      return <span className="text-text-dim ml-1 opacity-0 group-hover:opacity-50">^</span>;
    return (
      <span className="text-purple ml-1">
        {sortDir === "asc" ? "\u2191" : "\u2193"}
      </span>
    );
  }

  function handleLaunch(gpu: string, region: string) {
    setLaunchGPU(gpu);
    setLaunchRegion(region);
    setShowLaunch(true);
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <p className="type-ui-sm text-red-400">Failed to load GPU availability</p>
        <button
          onClick={() => mutate()}
          className="mt-3 type-ui-xs text-purple hover:text-purple-light transition-colors"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <>
      {/* Filters */}
      <div className="flex flex-wrap gap-3 mb-4">
        <select
          value={gpuFilter}
          onChange={(e) => setGpuFilter(e.target.value)}
          className="bg-bg border border-border rounded-lg px-3 py-2 type-ui-xs text-text-muted focus:outline-none focus:ring-2 focus:ring-purple/50 focus:border-purple/50 transition-all"
        >
          <option value="">All GPUs</option>
          {gpuModels.map((m) => (
            <option key={m} value={m}>
              {m}
            </option>
          ))}
        </select>

        <select
          value={regionFilter}
          onChange={(e) => setRegionFilter(e.target.value)}
          className="bg-bg border border-border rounded-lg px-3 py-2 type-ui-xs text-text-muted focus:outline-none focus:ring-2 focus:ring-purple/50 focus:border-purple/50 transition-all"
        >
          <option value="">All Regions</option>
          {regions.map((r) => (
            <option key={r} value={r}>
              {r}
            </option>
          ))}
        </select>

        <div className="flex rounded-lg border border-border overflow-hidden">
          {[
            { label: "All", value: "" },
            { label: "Spot", value: "spot" },
            { label: "On-Demand", value: "on_demand" },
          ].map((opt) => (
            <button
              key={opt.value}
              onClick={() => setTierFilter(opt.value)}
              className={cn(
                "px-3 py-2 type-ui-xs font-medium transition-colors",
                tierFilter === opt.value
                  ? "bg-purple-dim text-purple-light"
                  : "text-text-muted hover:text-text hover:bg-bg-card"
              )}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </div>

      {/* Table */}
      <div className="rounded-lg border border-border bg-bg-card/50 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border">
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  GPU Model
                </th>
                <th
                  className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3 cursor-pointer group"
                  onClick={() => toggleSort("vram")}
                >
                  VRAM
                  <SortIcon field="vram" />
                </th>
                <th
                  className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3 cursor-pointer group"
                  onClick={() => toggleSort("price")}
                >
                  Price
                  <SortIcon field="price" />
                </th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  Region
                </th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  Tier
                </th>
                <th
                  className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3 cursor-pointer group"
                  onClick={() => toggleSort("available")}
                >
                  Available
                  <SortIcon field="available" />
                </th>
                <th className="type-ui-2xs text-right text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                  Action
                </th>
              </tr>
            </thead>
            <tbody>
              {isLoading
                ? Array.from({ length: 6 }).map((_, i) => (
                    <SkeletonRow key={i} />
                  ))
                : filtered.length === 0
                  ? (
                    <tr>
                      <td colSpan={7} className="px-4 py-12 text-center">
                        <p className="type-ui-sm text-text-muted">
                          No GPU offerings match your filters
                        </p>
                        <button
                          onClick={() => {
                            setGpuFilter("");
                            setRegionFilter("");
                            setTierFilter("");
                          }}
                          className="mt-2 type-ui-xs text-purple hover:text-purple-light transition-colors"
                        >
                          Clear filters
                        </button>
                      </td>
                    </tr>
                  )
                  : filtered.map((offering, idx) => (
                    <tr
                      key={`${offering.gpu_model}-${offering.region}-${offering.tier}-${idx}`}
                      className="border-b border-border/50 hover:bg-bg-card transition-colors"
                    >
                      <td className="px-4 py-3">
                        <span className="type-ui-sm text-text font-medium">
                          {offering.gpu_model}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className="type-ui-sm text-text font-mono">
                          {offering.vram_gb}GB
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className="type-ui-sm text-text font-mono">
                          ${offering.price_per_hour.toFixed(2)}/hr
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className="type-ui-sm text-text-muted">
                          {offering.region}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={cn(
                            "type-ui-xs inline-flex items-center rounded-full px-2 py-0.5 font-medium",
                            offering.tier === "spot"
                              ? "bg-purple-dim text-purple-light"
                              : "bg-bg-card-hover text-text-muted"
                          )}
                        >
                          {offering.tier === "on_demand" ? "On-Demand" : "Spot"}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={cn(
                            "type-ui-sm font-mono",
                            offering.available_count > 5
                              ? "text-green"
                              : offering.available_count > 0
                                ? "text-amber-400"
                                : "text-red-400"
                          )}
                        >
                          {offering.available_count}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-right">
                        {offering.available_count > 0 ? (
                          <button
                            onClick={() =>
                              handleLaunch(offering.gpu_model, offering.region)
                            }
                            className="gradient-btn px-3 py-1.5 rounded-lg type-ui-xs font-medium transition-all"
                          >
                            Launch
                          </button>
                        ) : (
                          <span className="type-ui-xs text-text-dim">
                            Unavailable
                          </span>
                        )}
                      </td>
                    </tr>
                  ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Launch modal */}
      {showLaunch && (
        <LaunchInstanceForm
          onClose={() => setShowLaunch(false)}
          onSuccess={() => mutate()}
          defaultGPU={launchGPU}
          defaultRegion={launchRegion}
        />
      )}
    </>
  );
}
