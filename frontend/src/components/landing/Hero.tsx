import { HERO_CONTENT } from "@/lib/constants";
import { Button, Container } from "@/components/ui";

export function Hero() {
  return (
    <section className="relative pt-32 pb-24 overflow-visible">
      <Container>
        {/* Content */}
        <div className="text-center mx-auto max-w-[800px]">
          <h1 className="text-5xl md:text-7xl font-bold tracking-tight text-white leading-[1.1]">
            {HERO_CONTENT.headline}
          </h1>
          <p className="mt-6 text-lg md:text-xl text-text-muted max-w-[600px] mx-auto">
            {HERO_CONTENT.subtitle}
          </p>

          {/* CTA buttons */}
          <div className="flex items-center justify-center gap-4 mt-10">
            <Button variant="primary" size="lg" href="#">
              {HERO_CONTENT.primaryCTA}
            </Button>
            <Button variant="secondary" size="lg" href="#">
              {HERO_CONTENT.secondaryCTA}
            </Button>
          </div>

          {/* Metrics row */}
          <div className="flex items-center justify-center gap-8 md:gap-12 mt-16">
            {HERO_CONTENT.metrics.map((metric) => (
              <div key={metric.label} className="text-center">
                <div className="text-2xl md:text-3xl font-bold text-white">
                  {metric.value}
                </div>
                <div className="mt-1 text-sm text-text-muted">{metric.label}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Triangle + Rainbow glow */}
        <div className="relative mt-20 flex justify-center">
          <div className="absolute inset-0 flex justify-center">
            <div className="rainbow-glow w-[500px] h-[300px] md:w-[700px] md:h-[400px]" />
          </div>
          <svg
            className="relative z-10 w-[120px] h-[120px] md:w-[180px] md:h-[180px]"
            viewBox="0 0 1024 1024"
            fill="white"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path d="M512 128L896 832H128L512 128Z" />
          </svg>
        </div>
      </Container>
    </section>
  );
}
