import { Navbar } from "@/components/landing/Navbar";
import { Hero } from "@/components/landing/Hero";
import { TrustBar } from "@/components/landing/TrustBar";
import { PricingTable } from "@/components/landing/PricingTable";
import { HowItWorks } from "@/components/landing/HowItWorks";
import { Features } from "@/components/landing/Features";
import { FinalCTA } from "@/components/landing/FinalCTA";
import { Footer } from "@/components/landing/Footer";

export default function LandingPage() {
  return (
    <div className="stripe-borders film-grain">
      {/* Stripe-style vertical border lines */}
      <div className="stripe-lines" aria-hidden="true" />

      <Navbar />
      <main>
        <Hero />
        <TrustBar />
        <Features />
        <PricingTable />
        <HowItWorks />
        <FinalCTA />
      </main>
      <Footer />
    </div>
  );
}
