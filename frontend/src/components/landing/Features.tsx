import { Section, SectionHeader } from "@/components/ui/Section";
import { Card } from "@/components/ui/Card";
import { Icon } from "@/components/ui/Icon";
import { FEATURES } from "@/lib/constants";
import type { IconName } from "@/components/ui/Icon";

export function Features() {
  return (
    <Section id="features" className="pt-20 md:pt-24">
      <SectionHeader label="Features" title="Built for teams that ship" />

      <div className="grid gap-6 md:grid-cols-3">
        {FEATURES.map((feature) => (
          <Card key={feature.title}>
            <div className="mb-5 inline-flex h-12 w-12 items-center justify-center rounded-lg bg-purple-dim text-purple-light">
              <Icon name={feature.icon as IconName} />
            </div>
            <h3 className="type-h5 mb-2 font-semibold text-white">
              {feature.title}
            </h3>
            <p className="type-body leading-[1.7] text-text-muted">
              {feature.description}
            </p>
          </Card>
        ))}
      </div>
    </Section>
  );
}
