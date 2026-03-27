import { Navbar } from "@/components/landing/Navbar";
import { Footer } from "@/components/landing/Footer";
import { Container } from "@/components/ui/Container";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Privacy Policy",
};

export default function PrivacyPage() {
  return (
    <div className="film-grain">
      <div className="stripe-lines" />
      <Navbar />

      <section className="pt-[88px]">
        <Container className="py-20 md:py-28">
          <h1 className="type-h1 mb-4 font-bold text-white">Privacy Policy</h1>
          <p className="type-ui-sm mb-12 text-text-dim">
            Last updated: March 27, 2026
          </p>

          <div className="prose-gpu mx-auto max-w-[700px] space-y-8">
            <section>
              <h2 className="type-h5 mb-3 font-semibold text-white">
                Information We Collect
              </h2>
              <p className="type-body leading-[1.8] text-text-muted">
                We collect information you provide directly, such as your name,
                email address, and payment information when you create an account
                or use our services. We also collect usage data including GPU
                instance metrics, API calls, and session information to improve
                our platform.
              </p>
            </section>

            <section>
              <h2 className="type-h5 mb-3 font-semibold text-white">
                How We Use Your Information
              </h2>
              <p className="type-body leading-[1.8] text-text-muted">
                We use your information to provide and maintain our GPU cloud
                services, process payments, send service communications, and
                improve our platform. We do not sell your personal information to
                third parties.
              </p>
            </section>

            <section>
              <h2 className="type-h5 mb-3 font-semibold text-white">
                Data Security
              </h2>
              <p className="type-body leading-[1.8] text-text-muted">
                We implement industry-standard security measures to protect your
                data, including encryption in transit and at rest. Access to
                personal data is restricted to authorized personnel only.
              </p>
            </section>

            <section>
              <h2 className="type-h5 mb-3 font-semibold text-white">
                Contact
              </h2>
              <p className="type-body leading-[1.8] text-text-muted">
                For questions about this privacy policy, contact us at{" "}
                <a
                  href="mailto:privacy@gpu.ai"
                  className="text-purple-light transition-colors hover:text-white"
                >
                  privacy@gpu.ai
                </a>
                .
              </p>
            </section>
          </div>
        </Container>
      </section>

      <Footer />
    </div>
  );
}
