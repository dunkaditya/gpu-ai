"use client";

import useSWR from "swr";
import { Section, SectionHeader } from "@/components/ui/Section";
import { fetcher } from "@/lib/api";
import { PRICING_FALLBACK } from "@/lib/constants";
import type { PricingComparisonResponse } from "@/lib/types";

export function PricingTable() {
  const { data } = useSWR<PricingComparisonResponse>(
    "/api/v1/pricing/comparison",
    fetcher,
    { fallbackData: PRICING_FALLBACK, revalidateOnFocus: false },
  );

  const gpus = data?.gpus ?? PRICING_FALLBACK.gpus;
  const competitorNames = data?.competitor_names ?? PRICING_FALLBACK.competitor_names;

  return (
    <Section id="pricing" border={false} className="pt-16 md:pt-20">
      <SectionHeader
        label="Pricing"
        title="Compare and save"
        description="Real prices. No hidden fees. See how GPU.ai stacks up."
      />

      {/* Desktop Table */}
      <div className="hidden overflow-hidden rounded-xl border border-border md:block">
        <table className="w-full border-collapse text-left">
          <thead>
            <tr className="border-b border-border bg-bg-alt">
              <th className="type-ui-sm px-6 py-4 font-medium uppercase tracking-[0.05em] text-text-dim">
                GPU Model
              </th>
              <th className="type-ui-sm px-6 py-4 font-medium uppercase tracking-[0.05em] text-text-dim">
                VRAM
              </th>
              <th className="type-ui-sm bg-purple-dim px-6 py-4 font-semibold uppercase tracking-[0.05em] text-purple-light">
                GPU.ai
              </th>
              {competitorNames.map((name) => (
                <th
                  key={name}
                  className="type-ui-sm px-6 py-4 font-medium uppercase tracking-[0.05em] text-text-dim"
                >
                  {name}
                </th>
              ))}
              <th className="type-ui-sm px-6 py-4 font-medium uppercase tracking-[0.05em] text-text-dim">
                Savings
              </th>
            </tr>
          </thead>
          <tbody>
            {gpus.map((row) => (
              <tr
                key={row.gpu_model}
                className="border-b border-border last:border-0 transition-colors hover:bg-bg-card-hover"
              >
                <td className="px-6 py-5 text-[15px] font-medium text-white">
                  {row.display_name}
                </td>
                <td className="type-ui px-6 py-5 text-text-muted">
                  {row.vram_gb} GB
                </td>
                <td className="bg-purple-dim px-6 py-5 text-[16px] font-semibold text-white">
                  {row.gpuai_price !== null ? (
                    <>
                      ${row.gpuai_price.toFixed(2)}
                      <span className="type-ui-xs font-normal text-text-dim">
                        /hr
                      </span>
                    </>
                  ) : (
                    <span className="type-ui-xs text-text-dim">—</span>
                  )}
                </td>
                {competitorNames.map((compName) => {
                  const comp = row.competitors.find((c) => c.name === compName);
                  const price = comp?.price;
                  return (
                    <td
                      key={compName}
                      className="type-ui px-6 py-5 text-text-dim"
                    >
                      {price !== null && price !== undefined ? (
                        <>
                          ${price.toFixed(2)}
                          <span className="type-ui-2xs">/hr</span>
                        </>
                      ) : (
                        "—"
                      )}
                    </td>
                  );
                })}
                <td className="px-6 py-5">
                  {row.savings_pct !== null && row.savings_pct > 0 ? (
                    <span className="type-ui-sm inline-flex items-center rounded-full bg-green-dim px-3 py-1 font-semibold text-green">
                      {row.savings_pct}%
                    </span>
                  ) : (
                    <span className="type-ui-xs text-text-dim">—</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Mobile Cards */}
      <div className="flex flex-col gap-4 md:hidden">
        {gpus.map((row) => (
          <div
            key={row.gpu_model}
            className="rounded-xl border border-border bg-bg-card p-5"
          >
            <div className="mb-4 flex items-baseline justify-between">
              <div>
                <h3 className="text-[18px] font-semibold text-white">
                  {row.display_name}
                </h3>
                <span className="type-ui-sm text-text-dim">{row.vram_gb} GB</span>
              </div>
              {row.savings_pct !== null && row.savings_pct > 0 && (
                <span className="type-ui-sm inline-flex items-center rounded-full bg-green-dim px-3 py-1 font-semibold text-green">
                  Save {row.savings_pct}%
                </span>
              )}
            </div>
            <div className="mb-3 rounded-lg bg-purple-dim p-3">
              <span className="type-ui-xs text-text-muted">GPU.ai</span>
              <div className="text-[24px] font-bold text-white">
                {row.gpuai_price !== null ? (
                  <>
                    ${row.gpuai_price.toFixed(2)}
                    <span className="type-ui font-normal text-text-dim">/hr</span>
                  </>
                ) : (
                  <span className="type-ui text-text-dim">—</span>
                )}
              </div>
            </div>
            <div className="flex gap-3">
              {row.competitors
                .filter((c) => c.price !== null)
                .map((c) => (
                  <div
                    key={c.name}
                    className="min-w-0 flex-1 rounded-lg bg-bg-alt p-2 text-center"
                  >
                    <span className="type-ui-2xs block truncate text-text-dim">
                      {c.name}
                    </span>
                    <span className="type-ui text-text-muted line-through decoration-text-dim/30">
                      ${c.price!.toFixed(2)}
                    </span>
                  </div>
                ))}
            </div>
          </div>
        ))}
      </div>
    </Section>
  );
}
