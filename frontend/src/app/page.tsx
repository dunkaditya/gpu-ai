import { Navbar } from "@/components/landing/Navbar";
import { Hero } from "@/components/landing/Hero";
import { UseCaseTabs } from "@/components/landing/UseCaseTabs";
import { FeaturePillars } from "@/components/landing/FeaturePillars";
import { CLIDemo } from "@/components/landing/CLIDemo";
import { FinalCTA } from "@/components/landing/FinalCTA";
import { Footer } from "@/components/landing/Footer";

export default function LandingPage() {
  return (
    <div className="grid-background">
      <Navbar />
      <main>
        <Hero />
        <UseCaseTabs />
        <FeaturePillars />
        <CLIDemo />
        <FinalCTA />
      </main>
      <Footer />
    </div>
  );
}
