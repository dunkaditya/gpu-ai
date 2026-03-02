import { Section, SectionHeader } from "@/components/ui/Section";
import { Card } from "@/components/ui/Card";
import { HOW_IT_WORKS } from "@/lib/constants";

export function HowItWorks() {
  return (
    <Section>
      <SectionHeader
        label="How it works"
        title="Three steps. Under a minute."
      />

      <div className="grid gap-6 md:grid-cols-3">
        {HOW_IT_WORKS.map((item) => (
          <Card
            key={item.step}
            className="group relative overflow-hidden hover:bg-bg-card-hover"
          >
            {/* Large background number */}
            <span className="pointer-events-none absolute -right-2 -top-4 text-[120px] font-bold leading-none text-white/[0.03] select-none">
              {item.step}
            </span>

            <span className="type-ui mb-5 inline-block font-semibold text-purple-light">
              0{item.step}
            </span>
            <h3 className="type-h5 mb-2 font-semibold text-white">
              {item.title}
            </h3>
            <p className="type-body leading-[1.7] text-text-muted">
              {item.description}
            </p>
          </Card>
        ))}
      </div>
    </Section>
  );
}
