import { Container } from "@/components/ui/Container";
import { Counter } from "@/components/ui/Counter";
import { ComputeField } from "./ComputeField";
import { LightFixture } from "./LightFixture";
import { PricingWidget } from "./PricingWidget";
import { HERO_STATS } from "@/lib/constants";

export function Hero() {
  return (
    <section className="relative pt-[88px]">
      {/* Black opaque layer + mesh zone — same clip path */}
      {/* Black bg with hard clip edge */}
      {/* Black opaque layer — full width, covers both stripe lines in hero area */}
      <div className="absolute inset-0 z-[41] bg-bg" />
      {/* Mesh effects with soft fade along the same diagonal */}
      <div
        className="absolute -bottom-[10vh] left-0 right-0 top-0 z-[42]"
        style={{
          WebkitMaskImage: "linear-gradient(192deg, black 40%, transparent 50%)",
          maskImage: "linear-gradient(192deg, black 40%, transparent 50%)",
        }}
      >
        <ComputeField />
        <LightFixture />
      </div>

      <Container className="relative z-[44] flex min-h-[calc(100vh-480px)] flex-col justify-center py-10 pb-10">
        <div className="flex w-full flex-col items-center gap-16 lg:flex-row lg:gap-20">
          {/* Left — copy */}
          <div className="flex-1 text-center lg:text-left">
            {/* Headline */}
            <h1
              className="type-display animate-fade-up font-bold"
              style={{ animationDelay: "0.1s" }}
            >
              <span className="text-white">GPUs shouldn&apos;t</span>
              <br />
              <span className="text-white">cost so much.</span>
            </h1>

            <p
              className="type-h2 animate-fade-up mt-3 font-bold"
              style={{ animationDelay: "0.15s" }}
            >
              <span className="gradient-text">Now they don&apos;t.</span>
            </p>

            <p
              className="type-body-lg animate-fade-up mt-7 max-w-[600px] font-normal tracking-[-0.08em] text-text"
              style={{ animationDelay: "0.2s" }}
            >
              We built the infrastructure layer that GPU clouds should have built years ago. Source NVIDIA hardware globally, deploy instantly, a fraction of the cost.
            </p>

            {/* Stats */}
            <div
              className="animate-fade-up mt-8 flex flex-wrap justify-center gap-12 lg:justify-start lg:gap-14"
              style={{ animationDelay: "0.3s" }}
            >
              {HERO_STATS.map((stat) => (
                <div key={stat.label} className="flex flex-col items-center gap-1 lg:items-start">
                  <span className="type-h3 font-bold text-white">
                    <Counter
                      value={stat.value}
                      prefix={"prefix" in stat ? stat.prefix : undefined}
                      suffix={stat.suffix}
                    />
                  </span>
                  <span className="type-body font-normal text-text">{stat.label}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Right — interactive pricing widget */}
          <div
            className="animate-fade-up w-full flex-shrink-0 lg:w-[400px]"
            style={{ animationDelay: "0.3s" }}
          >
            <PricingWidget />
          </div>
        </div>
      </Container>
    </section>
  );
}
