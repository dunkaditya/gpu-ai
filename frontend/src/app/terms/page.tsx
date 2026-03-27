import { Navbar } from "@/components/landing/Navbar";
import { Footer } from "@/components/landing/Footer";
import { Container } from "@/components/ui/Container";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Terms of Service",
};

export default function TermsPage() {
  return (
    <div className="film-grain">
      <div className="stripe-lines" />
      <Navbar />

      <section className="pt-[88px]">
        <Container className="py-20 md:py-28">
          <h1 className="type-h1 mb-4 font-bold text-white">
            Terms of Service
          </h1>
          <p className="type-ui-sm mb-12 text-text-dim">
            Last updated: March 27, 2026
          </p>

          <div className="prose-gpu mx-auto max-w-[700px] space-y-8">
            <section>
              <h2 className="type-h5 mb-3 font-semibold text-white">
                Acceptance of Terms
              </h2>
              <p className="type-body leading-[1.8] text-text-muted">
                By accessing or using GPU.ai services, you agree to be bound by
                these Terms of Service. If you do not agree, you may not use our
                services.
              </p>
            </section>

            <section>
              <h2 className="type-h5 mb-3 font-semibold text-white">
                Service Description
              </h2>
              <p className="type-body leading-[1.8] text-text-muted">
                GPU.ai provides on-demand GPU cloud computing services, including
                instance provisioning, management, and billing. We aggregate GPU
                inventory from multiple providers to offer competitive pricing.
              </p>
            </section>

            <section>
              <h2 className="type-h5 mb-3 font-semibold text-white">
                Billing and Payment
              </h2>
              <p className="type-body leading-[1.8] text-text-muted">
                Usage is billed per second. You are responsible for all charges
                incurred under your account. We reserve the right to suspend
                services for non-payment.
              </p>
            </section>

            <section>
              <h2 className="type-h5 mb-3 font-semibold text-white">
                Acceptable Use
              </h2>
              <p className="type-body leading-[1.8] text-text-muted">
                You agree not to use our services for any unlawful purpose or in
                violation of any applicable laws. We reserve the right to
                terminate accounts that violate these terms.
              </p>
            </section>

            <section>
              <h2 className="type-h5 mb-3 font-semibold text-white">
                Contact
              </h2>
              <p className="type-body leading-[1.8] text-text-muted">
                For questions about these terms, contact us at{" "}
                <a
                  href="mailto:legal@gpu.ai"
                  className="text-purple-light transition-colors hover:text-white"
                >
                  legal@gpu.ai
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
