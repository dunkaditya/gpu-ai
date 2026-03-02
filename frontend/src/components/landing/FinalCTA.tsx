import { Container } from "@/components/ui/Container";
import { Button } from "@/components/ui/Button";

export function FinalCTA() {
  return (
    <section className="relative border-t border-border overflow-hidden py-24 md:py-32">
      {/* Radial glow background */}
      <div className="radial-glow pointer-events-none absolute inset-0" />

      <Container className="relative z-10 text-center">
        <h2 className="type-h2 font-bold text-white">
          Start deploying GPUs
          <br />
          in under 60 seconds
        </h2>
        <p className="type-body-lg mx-auto mt-3 max-w-[440px] text-text-muted">
          No credit card required. Free tier available.
        </p>
        <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
          <Button href="/sign-up">Get Started Free</Button>
          <Button variant="secondary" href="/contact">
            Talk to Sales
          </Button>
        </div>
      </Container>
    </section>
  );
}
