import { Button, Container } from "@/components/ui";

export function FinalCTA() {
  return (
    <section className="py-24">
      <Container>
        <div className="text-center">
          <h2 className="text-3xl md:text-5xl font-bold text-white">
            Start building with GPU.ai
          </h2>
          <p className="mt-6 text-text-muted text-lg max-w-[500px] mx-auto">
            Deploy your first GPU instance in under 60 seconds. No contracts, no
            commitments, per-second billing.
          </p>
          <div className="flex items-center justify-center gap-4 mt-10">
            <Button variant="primary" size="lg" href="#">
              Launch GPU Instance
            </Button>
            <Button variant="secondary" size="lg" href="#">
              Talk to our team
            </Button>
          </div>
        </div>
      </Container>
    </section>
  );
}
