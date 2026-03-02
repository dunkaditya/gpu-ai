import { Navbar } from "@/components/landing/Navbar";
import { Hero } from "@/components/landing/Hero";
import { TrustBar } from "@/components/landing/TrustBar";
import { PricingTable } from "@/components/landing/PricingTable";
import { HowItWorks } from "@/components/landing/HowItWorks";
import { Features } from "@/components/landing/Features";
import { CodeExample } from "@/components/landing/CodeExample";
import { FinalCTA } from "@/components/landing/FinalCTA";
import { Footer } from "@/components/landing/Footer";
import { EffectsToggle } from "@/components/EffectsToggle";

export default function LandingPage() {
  return (
    <div className="stripe-borders">
      {/* Stripe-style vertical border lines */}
      <div className="stripe-lines" aria-hidden="true" />
      <EffectsToggle />

      <Navbar />
      <main>
        <Hero />
        <PricingTable />
        <TrustBar />
        <HowItWorks />
        <Features />
        <CodeExample />
        <FinalCTA />
      </main>
      <Footer />
    </div>
  );
}
