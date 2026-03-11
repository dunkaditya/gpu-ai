"use client";

import { useMemo, useState } from "react";
import useSWR from "swr";
import { cn } from "@/lib/utils";
import { fetcher } from "@/lib/api";
import { GPUCard } from "@/components/cloud/GPUCard";
import { LaunchInstanceForm } from "@/components/cloud/LaunchInstanceForm";
import type { AvailableOffering, GPUCardData } from "@/lib/types";

function SkeletonCard() {
  return (
    <div className="bg-bg-card border border-border rounded-xl p-5 animate-pulse">
      <div className="flex items-start justify-between mb-4">
        <div className="h-6 bg-bg-card-hover rounded w-32" />
        <div className="h-5 bg-bg-card-hover rounded-full w-14" />
      </div>
      <div className="flex gap-4 mb-4 pb-4 border-b border-border/50">
        <div className="space-y-1">
          <div className="h-3 bg-bg-card-hover rounded w-8" />
          <div className="h-4 bg-bg-card-hover rounded w-16" />
        </div>
        <div className="space-y-1">
          <div className="h-3 bg-bg-card-hover rounded w-8" />
          <div className="h-4 bg-bg-card-hover rounded w-12" />
        </div>
        <div className="space-y-1">
          <div className="h-3 bg-bg-card-hover rounded w-10" />
          <div className="h-4 bg-bg-card-hover rounded w-14" />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-3 mb-4">
        <div className="space-y-1">
          <div className="h-3 bg-bg-card-hover rounded w-8" />
          <div className="h-4 bg-bg-card-hover rounded w-20" />
        </div>
        <div className="space-y-1">
          <div className="h-3 bg-bg-card-hover rounded w-14" />
          <div className="h-4 bg-bg-card-hover rounded w-20" />
        </div>
      </div>
      <div className="flex gap-1.5 mb-4">
        <div className="h-5 bg-bg-card-hover rounded-full w-16" />
        <div className="h-5 bg-bg-card-hover rounded-full w-12" />
      </div>
      <div className="flex items-center justify-between pt-3 border-t border-border/50">
        <div className="h-4 bg-bg-card-hover rounded w-20" />
        <div className="h-8 bg-bg-card-hover rounded-lg w-16" />
      </div>
    </div>
  );
}

export function GPUAvailabilityTable() {
  const { data, error, isLoading, mutate } = useSWR<{
    available: AvailableOffering[];
  }>("/api/v1/gpu/available", fetcher, { refreshInterval: 30000 });

  const [regionFilter, setRegionFilter] = useState("");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("asc");

  // Launch form state
  const [launchOffering, setLaunchOffering] = useState<
    AvailableOffering | undefined
  >();
  const [showLaunch, setShowLaunch] = useState(false);

  const offerings = data?.available ?? [];

  // Extract unique filter values
  const regions = useMemo(
    () => [...new Set(offerings.map((o) => o.region))].sort(),
    [offerings]
  );

  // Filter to on-demand only (beta), then apply user filters
  const filteredOfferings = useMemo(() => {
    let result = offerings.filter((o) => o.tier === "on_demand");
    if (regionFilter)
      result = result.filter((o) => o.region === regionFilter);
    return result;
  }, [offerings, regionFilter]);

  // Group filtered offerings into cards by gpu_model
  const cards: GPUCardData[] = useMemo(() => {
    const map = new Map<string, AvailableOffering[]>();
    for (const o of filteredOfferings) {
      const existing = map.get(o.gpu_model) ?? [];
      existing.push(o);
      map.set(o.gpu_model, existing);
    }
    const grouped = Array.from(map.entries()).map(([model, items]) => {
      const spotPrices = items
        .filter((i) => i.tier === "spot")
        .map((i) => i.price_per_hour);
      const onDemandPrices = items
        .filter((i) => i.tier === "on_demand")
        .map((i) => i.price_per_hour);
      return {
        gpu_model: model,
        vram_gb: items[0].vram_gb,
        cpu_cores: items[0].cpu_cores,
        ram_gb: items[0].ram_gb,
        storage_gb: items[0].storage_gb,
        regions: [...new Set(items.map((i) => i.region))],
        spot_price:
          spotPrices.length > 0 ? Math.min(...spotPrices) : undefined,
        on_demand_price:
          onDemandPrices.length > 0
            ? Math.min(...onDemandPrices)
            : undefined,
        total_available: items.reduce(
          (sum, i) => sum + i.available_count,
          0
        ),
        offerings: items,
      };
    });

    // Sort by price (use lowest available price from either tier)
    return grouped.sort((a, b) => {
      const priceA = Math.min(
        a.spot_price ?? Infinity,
        a.on_demand_price ?? Infinity
      );
      const priceB = Math.min(
        b.spot_price ?? Infinity,
        b.on_demand_price ?? Infinity
      );
      return sortDir === "asc" ? priceA - priceB : priceB - priceA;
    });
  }, [filteredOfferings, sortDir]);

  function handleLaunch(offering: AvailableOffering) {
    setLaunchOffering(offering);
    setShowLaunch(true);
  }

  function clearFilters() {
    setRegionFilter("");
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <p className="type-ui-sm text-red-400">
          Failed to load GPU availability
        </p>
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
      <div className="flex flex-wrap items-center gap-3 mb-6">
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

        <button
          onClick={() => setSortDir((d) => (d === "asc" ? "desc" : "asc"))}
          className="flex items-center gap-1.5 bg-bg border border-border rounded-lg px-3 py-2 type-ui-xs text-text-muted hover:text-text hover:border-border-light transition-all"
        >
          Price
          <span className="text-purple">
            {sortDir === "asc" ? "\u2191" : "\u2193"}
          </span>
        </button>
      </div>

      {/* Card grid */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      ) : cards.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <p className="type-ui-sm text-text-muted">
            No GPUs match your filters
          </p>
          <button
            onClick={clearFilters}
            className="mt-3 type-ui-xs text-purple hover:text-purple-light transition-colors"
          >
            Clear filters
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {cards.map((card) => (
            <GPUCard
              key={card.gpu_model}
              card={card}
              onLaunch={handleLaunch}
            />
          ))}
        </div>
      )}

      {/* Launch modal */}
      {showLaunch && (
        <LaunchInstanceForm
          onClose={() => setShowLaunch(false)}
          onSuccess={() => mutate()}
          offering={launchOffering}
          defaultGPU={launchOffering?.gpu_model}
          defaultRegion={launchOffering?.region}
        />
      )}
    </>
  );
}
