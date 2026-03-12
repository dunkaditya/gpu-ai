"use client";

import { cn } from "@/lib/utils";
import type { GPUCardData, AvailableOffering } from "@/lib/types";

interface GPUCardProps {
  card: GPUCardData;
  onLaunch: (offering: AvailableOffering) => void;
}

export function GPUCard({ card, onLaunch }: GPUCardProps) {
  function handleLaunch() {
    // Pick cheapest available on-demand offering
    const available = card.offerings.filter(
      (o) => o.available_count > 0 && o.tier === "on_demand"
    );
    if (available.length === 0) return;
    const cheapest = available.reduce((a, b) =>
      a.price_per_hour < b.price_per_hour ? a : b
    );
    onLaunch(cheapest);
  }

  return (
    <div className="bg-bg-card border border-border rounded-[10px] p-5 hover:border-border-light transition-colors">
      {/* Header: GPU model + VRAM badge */}
      <div className="flex items-start justify-between mb-4">
        <h3 className="type-ui-sm text-text font-medium leading-tight">
          {card.gpu_model}
        </h3>
        <span className="type-ui-2xs bg-bg-card-hover text-text-muted rounded-full px-2 py-0.5 whitespace-nowrap ml-2">
          {card.vram_gb} GB
        </span>
      </div>

      {/* Specs row */}
      <div className="flex items-center gap-3 mb-4 pb-4 border-b border-border/50">
        <div>
          <span className="type-ui-2xs text-text-dim uppercase block">CPU</span>
          <span className="type-ui-sm text-text-muted font-mono">
            {card.cpu_cores} cores
          </span>
        </div>
        <div>
          <span className="type-ui-2xs text-text-dim uppercase block">RAM</span>
          <span className="type-ui-sm text-text-muted font-mono">
            {card.ram_gb} GB
          </span>
        </div>
        <div>
          <span className="type-ui-2xs text-text-dim uppercase block">
            Storage
          </span>
          <span className="type-ui-sm text-text-muted font-mono">
            {card.storage_gb} GB
          </span>
        </div>
      </div>

      {/* Price display */}
      <div className="mb-4">
        <span className="type-ui-sm font-mono text-text">
          {card.on_demand_price != null && card.on_demand_price !== Infinity
            ? `$${card.on_demand_price.toFixed(2)}/hr`
            : "--"}
        </span>
      </div>

      {/* Region tags */}
      <div className="flex flex-wrap gap-1.5 mb-4">
        {card.regions.map((region) => (
          <span
            key={region}
            className="type-ui-2xs bg-bg-card-hover text-text-dim rounded-full px-2 py-0.5"
          >
            {region}
          </span>
        ))}
      </div>

      {/* Availability + Launch */}
      <div className="flex items-center justify-between pt-3 border-t border-border/50">
        <span
          className={cn(
            "type-ui-sm font-mono",
            card.total_available > 5
              ? "text-green"
              : card.total_available > 0
                ? "text-amber-400"
                : "text-red-400"
          )}
        >
          {card.total_available} available
        </span>
        {card.total_available > 0 ? (
          <button
            onClick={handleLaunch}
            className="btn-primary px-4 py-1.5 rounded-lg type-ui-xs font-medium"
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
