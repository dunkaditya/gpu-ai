import { FEATURE_PILLARS } from "@/lib/constants";
import { Container } from "@/components/ui";

export function FeaturePillars() {
  return (
    <section className="py-24">
      <Container>
        {FEATURE_PILLARS.map((pillar, index) => (
          <div
            key={pillar.id}
            className={
              "py-16 border-b border-white/[0.08]" +
              (index === FEATURE_PILLARS.length - 1 ? " border-b-0" : "")
            }
          >
            <div className="grid grid-cols-1 md:grid-cols-2 gap-12 items-start">
              {/* Left column: text */}
              <div>
                <span className="text-sm font-medium uppercase tracking-widest text-text-dim">
                  {pillar.title}
                </span>
                <h3 className="mt-4 text-3xl md:text-4xl font-bold text-white leading-tight">
                  {pillar.subtitle}
                </h3>
                <p className="mt-4 text-text-muted text-lg leading-relaxed">
                  {pillar.description}
                </p>
              </div>

              {/* Right column: features */}
              <div className="space-y-4">
                {pillar.features.map((feature) => (
                  <div key={feature} className="flex items-start gap-3">
                    <div className="w-2 h-2 rounded-full bg-white/20 mt-2 shrink-0" />
                    <span className="text-text-muted">{feature}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        ))}
      </Container>
    </section>
  );
}
