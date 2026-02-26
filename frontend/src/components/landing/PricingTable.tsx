import { Section, SectionHeader } from "@/components/ui/Section";
import { PRICING_DATA, COMPETITOR_NAMES } from "@/lib/constants";

export function PricingTable() {
  return (
    <Section id="pricing">
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
              {COMPETITOR_NAMES.map((name) => (
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
            {PRICING_DATA.map((row) => (
              <tr
                key={row.gpu}
                className="border-b border-border last:border-0 transition-colors hover:bg-bg-card-hover"
              >
                <td className="px-6 py-5 text-[15px] font-medium text-white">
                  {row.gpu}
                </td>
                <td className="type-ui px-6 py-5 text-text-muted">
                  {row.vram}
                </td>
                <td className="bg-purple-dim px-6 py-5 text-[16px] font-semibold text-white">
                  ${row.gpuai.toFixed(2)}
                  <span className="type-ui-xs font-normal text-text-dim">
                    /hr
                  </span>
                </td>
                {row.competitors.map((c) => (
                  <td
                    key={c.name}
                    className="type-ui px-6 py-5 text-text-dim"
                  >
                    ${c.price.toFixed(2)}
                    <span className="type-ui-2xs">/hr</span>
                  </td>
                ))}
                <td className="px-6 py-5">
                  <span className="type-ui-sm inline-flex items-center rounded-full bg-green-dim px-3 py-1 font-semibold text-green">
                    {row.savings}%
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Mobile Cards */}
      <div className="flex flex-col gap-4 md:hidden">
        {PRICING_DATA.map((row) => (
          <div
            key={row.gpu}
            className="rounded-xl border border-border bg-bg-card p-5"
          >
            <div className="mb-4 flex items-baseline justify-between">
              <div>
                <h3 className="text-[18px] font-semibold text-white">
                  {row.gpu}
                </h3>
                <span className="type-ui-sm text-text-dim">{row.vram}</span>
              </div>
              <span className="type-ui-sm inline-flex items-center rounded-full bg-green-dim px-3 py-1 font-semibold text-green">
                Save {row.savings}%
              </span>
            </div>
            <div className="mb-3 rounded-lg bg-purple-dim p-3">
              <span className="type-ui-xs text-text-muted">GPU.ai</span>
              <div className="text-[24px] font-bold text-white">
                ${row.gpuai.toFixed(2)}
                <span className="type-ui font-normal text-text-dim">/hr</span>
              </div>
            </div>
            <div className="flex gap-3">
              {row.competitors.map((c) => (
                <div
                  key={c.name}
                  className="flex-1 rounded-lg bg-bg-alt p-2 text-center"
                >
                  <span className="type-ui-2xs block text-text-dim">
                    {c.name}
                  </span>
                  <span className="type-ui text-text-muted line-through decoration-text-dim/30">
                    ${c.price.toFixed(2)}
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
