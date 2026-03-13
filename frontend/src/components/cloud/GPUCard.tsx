"use client";

import { cn } from "@/lib/utils";
import { getDisplayName, classifyGPU } from "@/lib/gpu-categories";
import { getRegionDisplay } from "@/lib/regions";
import type { GPUCardData } from "@/lib/types";

interface GPUCardProps {
  card: GPUCardData;
  onLaunch: (card: GPUCardData) => void;
  featured?: boolean;
}

export function GPUCard({ card, onLaunch }: GPUCardProps) {
  function handleLaunch() {
    const available = card.offerings.filter(
      (o) => o.available_count > 0 && o.tier === "on_demand"
    );
    if (available.length === 0) return;
    onLaunch(card);
  }

  const generation = classifyGPU(card.gpu_model);

  return (
    <div
      className="bg-bg-card border border-border rounded-[10px] p-5 hover:border-border-light transition-all snake-border"
    >
      {/* Header: GPU name + VRAM inline + generation tag */}
      <div className="flex items-start justify-between mb-1">
        <div className="flex items-baseline gap-2 min-w-0">
          <h3 className="text-[15px] text-text font-semibold leading-tight tracking-[-0.01em] truncate">
            {getDisplayName(card.gpu_model)}
          </h3>
          <span className="type-ui-2xs text-text-dim whitespace-nowrap">
            {card.vram_gb} GB
          </span>
        </div>
        {/* Availability dot */}
        <span className="flex items-center gap-1.5 shrink-0 ml-2">
          <span
            className={cn(
              "w-1.5 h-1.5 rounded-full",
              card.total_available > 5
                ? "bg-green"
                : card.total_available > 0
                  ? "bg-amber-400"
                  : "bg-red-400"
            )}
          />
          <span className="type-ui-2xs text-text-dim font-mono">
            {card.total_available}
          </span>
        </span>
      </div>

      {/* Generation label */}
      <span className="type-ui-2xs text-text-dim/70 mb-4 block">
        {generation}
      </span>

      {/* Specs row */}
      <div className="flex items-center gap-4 mb-4 pb-4 border-b border-border/50">
        <div>
          <span className="type-ui-2xs text-text-dim uppercase block">CPU</span>
          <span className="type-ui-xs text-text-muted font-mono">
            {card.cpu_cores}c
          </span>
        </div>
        <div>
          <span className="type-ui-2xs text-text-dim uppercase block">RAM</span>
          <span className="type-ui-xs text-text-muted font-mono">
            {card.ram_gb} GB
          </span>
        </div>
        <div>
          <span className="type-ui-2xs text-text-dim uppercase block">Disk</span>
          <span className="type-ui-xs text-text-muted font-mono">
            {card.storage_gb} GB
          </span>
        </div>
      </div>

      {/* Region tags */}
      <div className="flex flex-wrap gap-1.5 mb-4">
        {card.regions.map((region) => {
          const r = getRegionDisplay(region);
          return (
            <span
              key={region}
              className="type-ui-2xs bg-bg-card-hover text-text-dim rounded px-1.5 py-0.5"
            >
              {r.flag} {r.label}
            </span>
          );
        })}
      </div>

      {/* Price + Launch */}
      <div className="flex items-center justify-between">
        <span className="text-[15px] font-mono text-text font-semibold">
          {card.on_demand_price != null && card.on_demand_price !== Infinity
            ? `$${card.on_demand_price.toFixed(2)}`
            : "--"}
          <span className="type-ui-2xs text-text-dim font-normal">/hr</span>
        </span>
        {card.total_available > 0 ? (
          <button
            onClick={handleLaunch}
            className="btn-primary px-4 py-1.5 rounded-md type-ui-xs font-medium"
          >
            Launch
          </button>
        ) : (
          <span className="type-ui-xs text-text-dim">Unavailable</span>
        )}
      </div>
    </div>
  );
}
