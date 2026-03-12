"use client";

import { useMemo, useState } from "react";
import useSWR from "swr";
import { useDebouncedCallback } from "use-debounce";
import { fetcher } from "@/lib/api";
import { GPUCard } from "@/components/cloud/GPUCard";
import { LaunchInstanceForm } from "@/components/cloud/LaunchInstanceForm";
import { EmptyState } from "@/components/cloud/EmptyState";
import { GPU_CATEGORIES, classifyGPU } from "@/lib/gpu-categories";
import type { AvailableOffering, GPUCardData } from "@/lib/types";

const CATEGORY_LABELS = ["All", ...GPU_CATEGORIES.map((c) => c.label)];

function SkeletonCard() {
  return (
    <div className="bg-bg-card border border-border rounded-[10px] p-5 animate-pulse">
      <div className="flex items-start justify-between mb-4">
        <div className="h-6 bg-bg-card-hover rounded w-32" />
        <div className="h-5 bg-bg-card-hover rounded-full w-14" />
      </div>
      <div className="flex gap-3 mb-4 pb-4 border-b border-border/50">
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

  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [activeCategory, setActiveCategory] = useState("All");
  const [regionFilter, setRegionFilter] = useState("");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("asc");

  // Launch form state
  const [launchOffering, setLaunchOffering] = useState<
    AvailableOffering | undefined
  >();
  const [showLaunch, setShowLaunch] = useState(false);

  const debouncedSetQuery = useDebouncedCallback((value: string) => {
    setDebouncedQuery(value);
  }, 200);

  function handleSearchChange(e: React.ChangeEvent<HTMLInputElement>) {
    const value = e.target.value;
    setSearchQuery(value);
    debouncedSetQuery(value);
  }

  const offerings = data?.available ?? [];

  // Extract unique regions
  const regions = useMemo(
    () => [...new Set(offerings.map((o) => o.region))].sort(),
    [offerings]
  );

  const filtersActive =
    activeCategory !== "All" || debouncedQuery !== "" || regionFilter !== "";

  // Filter chain: on_demand -> category -> search -> region -> group -> sort
  const cards: GPUCardData[] = useMemo(() => {
    // 1. Start with on_demand offerings
    let result = offerings.filter((o) => o.tier === "on_demand");

    // 2. Category filter
    if (activeCategory !== "All") {
      result = result.filter(
        (o) => classifyGPU(o.gpu_model) === activeCategory
      );
    }

    // 3. Search filter (debounced)
    if (debouncedQuery) {
      const q = debouncedQuery.toLowerCase();
      result = result.filter(
        (o) =>
          o.gpu_model.toLowerCase().includes(q) ||
          o.region.toLowerCase().includes(q)
      );
    }

    // 4. Region filter
    if (regionFilter) {
      result = result.filter((o) => o.region === regionFilter);
    }

    // 5. Group into GPUCardData by gpu_model
    const map = new Map<string, AvailableOffering[]>();
    for (const o of result) {
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

    // 6. Sort by price
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
  }, [offerings, activeCategory, debouncedQuery, regionFilter, sortDir]);

  function handleLaunch(offering: AvailableOffering) {
    setLaunchOffering(offering);
    setShowLaunch(true);
  }

  function clearFilters() {
    setSearchQuery("");
    setDebouncedQuery("");
    setActiveCategory("All");
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
      {/* Filter toolbar */}
      <div className="flex flex-wrap items-center gap-3 mb-4">
        {/* Search input */}
        <div className="relative w-full sm:w-64">
          <svg
            className="absolute left-3 top-1/2 -translate-y-1/2 text-text-dim"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35" />
          </svg>
          <input
            type="text"
            value={searchQuery}
            onChange={handleSearchChange}
            placeholder="Search GPUs..."
            className="w-full bg-bg border border-border rounded-lg px-3 py-2 pl-9 type-ui-sm text-text placeholder:text-text-dim focus:outline-none focus:ring-1 focus:ring-border-light focus:border-border-light transition-all"
          />
        </div>

        {/* Category chips */}
        <div className="flex items-center gap-1 overflow-x-auto flex-1 min-w-0">
          {CATEGORY_LABELS.map((label) => (
            <button
              key={label}
              onClick={() => setActiveCategory(label)}
              className={
                label === activeCategory
                  ? "bg-bg-card-hover text-text rounded-md px-3 py-1.5 type-ui-xs font-medium whitespace-nowrap transition-colors"
                  : "text-text-dim hover:text-text-muted rounded-md px-3 py-1.5 type-ui-xs font-medium whitespace-nowrap transition-colors"
              }
            >
              {label}
            </button>
          ))}
        </div>

        {/* Region filter */}
        <select
          value={regionFilter}
          onChange={(e) => setRegionFilter(e.target.value)}
          className="bg-bg border border-border rounded-lg px-3 py-2 type-ui-xs text-text-muted focus:outline-none focus:ring-1 focus:ring-border-light focus:border-border-light transition-all"
        >
          <option value="">All Regions</option>
          {regions.map((r) => (
            <option key={r} value={r}>
              {r}
            </option>
          ))}
        </select>

        {/* Sort button */}
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

      {/* Results count */}
      {filtersActive && !isLoading && (
        <p className="type-ui-2xs text-text-dim mb-4">
          {cards.length} {cards.length === 1 ? "GPU" : "GPUs"} found
        </p>
      )}

      {/* Card grid */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      ) : cards.length === 0 ? (
        <EmptyState
          icon={
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <rect x="4" y="4" width="16" height="16" rx="2" />
              <path d="M9 9h6v6H9z" />
            </svg>
          }
          title="No GPUs match your filters"
          description="Try broadening your search or selecting a different category."
          action={{ label: "Clear filters", onClick: clearFilters }}
        />
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
