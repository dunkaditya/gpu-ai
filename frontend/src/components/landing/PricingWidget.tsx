"use client";

import { useState, useEffect } from "react";
import useSWR from "swr";
import { fetcher } from "@/lib/api";
import { PRICING_FALLBACK } from "@/lib/constants";
import type { PricingComparisonResponse } from "@/lib/types";

interface PriceBarProps {
  name: string;
  price: number;
  maxP: number;
  isSelf: boolean;
  delay: number;
}

function PriceBar({ name, price, maxP, isSelf, delay }: PriceBarProps) {
  const [w, setW] = useState(0);

  useEffect(() => {
    const t = setTimeout(() => setW((price / maxP) * 100), 300 + delay);
    return () => clearTimeout(t);
  }, [price, maxP, delay]);

  return (
    <div className="mb-2.5">
      <div className="mb-1 flex justify-between type-ui-xs">
        <span className={isSelf ? "font-semibold text-green" : "text-text-muted"}>
          {name}
        </span>
        <span
          className={`type-ui-xs font-medium ${
            isSelf ? "text-green" : "text-text-dim"
          }`}
        >
          ${price.toFixed(2)}
        </span>
      </div>
      <div className="h-[5px] overflow-hidden rounded-[3px] bg-border/70">
        <div
          className={`h-full rounded-[3px] transition-[width] duration-[2700ms] ${
            isSelf ? "bg-green" : "bg-border-light"
          }`}
          style={{
            width: `${w}%`,
            transitionTimingFunction: "cubic-bezier(0.16, 1, 0.3, 1)",
          }}
        />
      </div>
    </div>
  );
}

// Select top 4 GPUs for the widget tabs.
const WIDGET_GPU_COUNT = 4;

export function PricingWidget() {
  const { data } = useSWR<PricingComparisonResponse>(
    "/api/v1/pricing/comparison",
    fetcher,
    { fallbackData: PRICING_FALLBACK, revalidateOnFocus: false },
  );
  const [selGpu, setSelGpu] = useState(0);

  const allGpus = data?.gpus ?? PRICING_FALLBACK.gpus;
  // Filter to GPUs that have a GPU.ai price, take top N.
  const gpus = allGpus
    .filter((g) => g.gpuai_price !== null)
    .slice(0, WIDGET_GPU_COUNT);

  if (gpus.length === 0) return null;

  const safeIdx = selGpu < gpus.length ? selGpu : 0;
  const gpu = gpus[safeIdx];
  const price = gpu.gpuai_price!;

  // Build competitor bars (only those with a price for this GPU).
  const compArr = gpu.competitors
    .filter((c): c is { name: string; price: number } => c.price !== null);
  const maxP = Math.max(price, ...compArr.map((c) => c.price));
  // Calculate savings vs the next highest competitor price (cheapest one above our price).
  const higherPrices = compArr.map((c) => c.price).filter((p) => p > price);
  const nextHighest = higherPrices.length > 0 ? Math.min(...higherPrices) : null;
  const savePct = nextHighest !== null ? Math.round(((nextHighest - price) / nextHighest) * 100) : null;

  return (
    <div className="relative w-full max-w-[400px]">
      {/* Outer glow */}
      <div
        className="absolute -inset-8 rounded-3xl opacity-40 blur-2xl"
        style={{
          background: "radial-gradient(ellipse at 50% 0%, rgba(124, 107, 240, 0.3), transparent 70%)",
        }}
      />

      {/* Gradient border */}
      <div
        className="absolute -inset-[1px] rounded-2xl"
        style={{
          background: "linear-gradient(160deg, rgba(255,255,255,0.18) 0%, rgba(124, 107, 240, 0.15) 40%, rgba(124, 107, 240, 0.06) 60%, rgba(255,255,255,0.1) 100%)",
        }}
      />

      <div
        className="relative overflow-hidden rounded-2xl"
        style={{
          background: "linear-gradient(165deg, rgba(20, 18, 40, 0.85) 0%, rgba(14, 13, 28, 0.92) 100%)",
          backdropFilter: "blur(40px) saturate(1.4)",
          WebkitBackdropFilter: "blur(40px) saturate(1.4)",
          boxShadow: [
            "inset 0 1px 0 0 rgba(255,255,255,0.07)",
            "inset 0 0 30px rgba(124, 107, 240, 0.04)",
            "0 25px 60px -12px rgba(0,0,0,0.5)",
            "0 0 1px rgba(255,255,255,0.05)",
          ].join(", "),
        }}
      >
        {/* Top refraction highlight */}
        <div
          className="pointer-events-none absolute inset-x-0 top-0 h-[1px]"
          style={{
            background: "linear-gradient(90deg, transparent 10%, rgba(255,255,255,0.12) 30%, rgba(255,255,255,0.2) 50%, rgba(255,255,255,0.12) 70%, transparent 90%)",
          }}
        />

        {/* GPU Tabs */}
        <div className="flex border-b border-white/[0.06]">
          {gpus.map((g, i) => (
            <button
              key={g.gpu_model}
              onClick={() => setSelGpu(i)}
              className={`relative flex-1 px-2 py-3 type-ui-2xs transition-all ${
                safeIdx === i
                  ? "font-semibold text-white"
                  : "text-text-dim hover:text-text-muted"
              }`}
            >
              {safeIdx === i && (
                <div
                  className="absolute inset-0"
                  style={{
                    background: "linear-gradient(to bottom, rgba(124, 107, 240, 0.1), transparent)",
                    borderBottom: "1.5px solid rgba(124, 107, 240, 0.6)",
                  }}
                />
              )}
              <span className="relative">{g.display_name.split(" ")[0]}</span>
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="px-6 py-5">
          <div className="mb-1.5 flex items-baseline justify-between">
            <span className="text-[16px] font-semibold text-white">{gpu.display_name}</span>
            <span className="type-ui-2xs text-text-dim">{gpu.vram_gb} GB</span>
          </div>

          <div className="mb-5 flex items-baseline gap-0.5">
            <span className="text-[32px] font-extrabold tracking-tight text-green">
              ${price.toFixed(2)}
            </span>
            <span className="type-ui-sm text-text-dim">/gpu/hr</span>
            {savePct !== null && savePct > 0 && (
              <span
                className="ml-auto inline-flex items-center rounded-[5px] px-2 py-0.5 type-ui-2xs font-semibold text-green"
                style={{
                  background: "rgba(34, 197, 94, 0.08)",
                  border: "1px solid rgba(34, 197, 94, 0.15)",
                  boxShadow: "inset 0 0 8px rgba(34, 197, 94, 0.06)",
                }}
              >
                -{savePct}%
              </span>
            )}
          </div>

          <div className="mb-2.5 text-[10px] font-semibold uppercase tracking-[1px] text-text-dim">
            vs. Competitors
          </div>

          <PriceBar name="GPU.ai" price={price} maxP={maxP} isSelf delay={0} />
          {compArr.map((c, i) => (
            <PriceBar
              key={c.name}
              name={c.name}
              price={c.price}
              maxP={maxP}
              isSelf={false}
              delay={(i + 1) * 60}
            />
          ))}

          <div
            className="mt-5 border-t border-white/[0.06] pt-3.5"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <div
                  className={`h-[7px] w-[7px] rounded-full ${
                    gpu.available_count > 30
                      ? "bg-green shadow-[0_0_8px_rgba(34,197,94,0.27)]"
                      : "bg-amber-400 shadow-[0_0_8px_rgba(251,191,36,0.27)]"
                  }`}
                />
                <span className="type-ui-xs text-text-muted">{gpu.available_count} available</span>
              </div>
              <span className="type-ui-2xs text-text-dim">On-demand</span>
            </div>
            <a
              href="/cloud/gpu-availability"
              className="mt-3.5 flex w-full items-center justify-center gap-2 rounded-lg py-3 text-[15px] font-semibold text-white transition-all"
              style={{
                background: "linear-gradient(135deg, #1a9d4a 0%, #16a34a 50%, #15803d 100%)",
                boxShadow: "0 0 20px rgba(34, 197, 94, 0.15), inset 0 1px 0 rgba(255,255,255,0.1)",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.boxShadow = "0 0 32px rgba(34, 197, 94, 0.3), inset 0 1px 0 rgba(255,255,255,0.15)";
                e.currentTarget.style.background = "linear-gradient(135deg, #1fb854 0%, #22c55e 50%, #16a34a 100%)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.boxShadow = "0 0 20px rgba(34, 197, 94, 0.15), inset 0 1px 0 rgba(255,255,255,0.1)";
                e.currentTarget.style.background = "linear-gradient(135deg, #1a9d4a 0%, #16a34a 50%, #15803d 100%)";
              }}
            >
              LAUNCH INSTANCE →
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}
