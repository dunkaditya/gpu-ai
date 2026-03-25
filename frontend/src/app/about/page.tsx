import { Navbar } from "@/components/landing/Navbar";
import { Footer } from "@/components/landing/Footer";
import { Container } from "@/components/ui/Container";
import { Button } from "@/components/ui/Button";
import Image from "next/image";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "About",
};

/* ── Values ── */
const VALUES = [
  {
    title: "Transparency",
    body: "No hidden fees, no opaque pricing tiers, no surprise bills. The rate you see is the rate you pay.",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="h-5 w-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
        <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
      </svg>
    ),
  },
  {
    title: "Speed",
    body: "Infrastructure should disappear. One command, one API call — GPU capacity deployed in seconds, not hours of Terraform and tickets.",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="h-5 w-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
      </svg>
    ),
  },
  {
    title: "Access",
    body: "The best hardware shouldn't be gatekept behind six-figure commitments. We bring data-center GPUs to teams of every size.",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="h-5 w-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 10.5V6.75a4.5 4.5 0 119 0v3.75M3.75 21.75h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H3.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
      </svg>
    ),
  },
];

/* ── Team ── */
interface TeamMember {
  name: string;
  role: string;
  photo: string | null;
  bio: string;
  linkedin?: string;
  previous?: { label: string; logo?: React.ReactNode; showPrefix?: boolean }[];
  photoClassName?: string;
}

const TEAM: TeamMember[] = [
  {
    name: "Ranbir Badwal",
    role: "Chief Executive Officer",
    photo: "/team/ranbir.jpg",
    bio: "Co-founded NovaCore, India's first GPU cloud with Blackwells.",
    linkedin: "https://www.linkedin.com/in/ranbir-badwal-216015204/",
    previous: [{ label: "NovaCore", logo: <NovacoreLogo />, showPrefix: false }],
  },
  {
    name: "Aryamaan Singhania",
    role: "Chief Financial Officer",
    photo: "/team/aryamaan.jpg",
    photoClassName: "scale-[1.15]",
    bio: "Co-founded NovaCore, India's first GPU cloud with Blackwells.",
    linkedin: "https://www.linkedin.com/in/aryamaansinghania/",
    previous: [
      { label: "NovaCore", logo: <NovacoreLogo />, showPrefix: false },
      { label: "Advantage Capital", logo: <AdvantageLogo />, showPrefix: false },
    ],
  },
  {
    name: "Aditya Reddy",
    role: "Chief Technology Officer",
    photo: "/team/aditya.jpg",
    bio: "Previously engineering at Amazon.",
    linkedin: "https://www.linkedin.com/in/aditya~reddy/",
    previous: [{ label: "Amazon", logo: <AmazonLogo />, showPrefix: false }],
  },
  {
    name: "John Nguyen",
    role: "Chief Operating Officer",
    photo: "/team/john.png",
    bio: "Co-founded BitSync, a US-based bare metal GPU cloud.",
    previous: [{ label: "BitSync", logo: <BitSyncLogo />, showPrefix: false }],
  },
  {
    name: "William Han",
    role: "Chief Product Officer",
    photo: "/team/william.jpg",
    bio: "Co-founded BitSync, a US-based bare metal GPU cloud.",
    linkedin: "https://www.linkedin.com/in/will-h-81b7671bb/",
    previous: [{ label: "BitSync", logo: <BitSyncLogo />, showPrefix: false }],
  },
];

/* ── Company Logos ── */
const logoBase = "w-auto grayscale brightness-200 mix-blend-lighten";

function AmazonLogo() {
  return <Image src="/logos/amazon.png" alt="Amazon" width={1280} height={387} className={`${logoBase} h-[20px]`} />;
}

function BitSyncLogo() {
  return <Image src="/logos/bitsync.png" alt="BitSync" width={443} height={120} className="h-[26px] w-auto grayscale invert opacity-50" />;
}

function NovacoreLogo() {
  return <Image src="/logos/novacore.png" alt="NovaCore" width={600} height={535} className="h-auto w-[100px] sm:w-[140px] -mr-4 sm:-mr-12 grayscale invert brightness-75" />;
}

function AdvantageLogo() {
  return <Image src="/logos/advantage.png" alt="Advantage Capital" width={6381} height={2086} className={`${logoBase} h-[26px]`} />;
}

function TotalEnergiesLogo() {
  return <Image src="/logos/totalenergies.png" alt="TotalEnergies" width={600} height={437} className={`${logoBase} h-[40px] sm:h-[56px]`} />;
}

function IndiaGovLogo() {
  return <Image src="/logos/india.png" alt="Government of India" width={800} height={687} className="h-[36px] sm:h-[50px] w-auto opacity-70" />;
}

function RashiLogo() {
  return <Image src="/logos/rashi.png" alt="Rashi Group" width={389} height={150} className={`${logoBase} h-[38px] sm:h-[54px]`} />;
}

/* ── Board Members ── */
const BOARD_MEMBERS: TeamMember[] = [
  {
    name: "Tuyen Nguyen",
    role: "Board Member",
    photo: "/team/tuyen.png",
    bio: "Founded Saigon Gas Holdings, acquired by TotalEnergies to form TotalEnergies LPG Vietnam.",
    previous: [{ label: "TotalEnergies", logo: <TotalEnergiesLogo />, showPrefix: false }],
  },
  {
    name: "Ashish Singhania",
    role: "Board Member",
    photo: "/team/ashish.png",
    bio: "25+ years in capital markets. Founded Rashi Group, now building hydropower infrastructure for India.",
    previous: [{ label: "Rashi Group", logo: <RashiLogo />, showPrefix: false }],
  },
];

/* ── LinkedIn Icon ── */
function LinkedInIcon() {
  return (
    <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="currentColor">
      <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 01-2.063-2.065 2.064 2.064 0 112.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z" />
    </svg>
  );
}

export default function AboutPage() {
  return (
    <div className="film-grain min-h-screen">
      <div className="stripe-lines" />
      <Navbar />

      {/* ═══ Hero ═══ */}
      <section className="relative pt-[88px]">
        <div className="radial-glow pointer-events-none absolute inset-0" />
        <Container className="relative z-10 pb-10 pt-24 md:pt-32">
          <div className="mx-auto max-w-[800px] text-center">
            <p className="animate-fade-up font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-purple-light">
              About GPU.ai
            </p>
            <h1
              className="type-display animate-fade-up mt-5 font-bold text-white"
              style={{ animationDelay: "0.08s" }}
            >
              GPU compute for{" "}
              <span className="gradient-text">everyone.</span>
            </h1>
            <p
              className="animate-fade-up mx-auto mt-6 max-w-[560px] text-[17px] font-normal leading-relaxed text-text-muted"
              style={{ animationDelay: "0.14s" }}
            >
              We aggregate GPU inventory from every major cloud provider into a
              single platform — so you always get the best price and the fastest
              deploy.
            </p>
          </div>
        </Container>
      </section>

      {/* ═══ Mission — The Problem ═══ */}
      <section>
        <Container className="py-24 md:py-32">
          <div className="grid grid-cols-1 gap-16 lg:grid-cols-12 lg:gap-20">
            {/* Left: headline */}
            <div className="lg:col-span-5">
              <p className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-purple-light">
                The Problem
              </p>
              <h2 className="mt-4 text-[28px] font-bold leading-[1.15] tracking-[-0.02em] text-white md:text-[36px]">
                GPU clouds are fragmented, expensive, and slow.
              </h2>
            </div>

            {/* Right: body with accent line */}
            <div className="flex lg:col-span-7">
              <div className="mr-6 hidden w-px shrink-0 bg-gradient-to-b from-purple/60 via-purple/20 to-transparent lg:block" />
              <div className="flex flex-col justify-center gap-5">
                <p className="text-[15px] font-normal leading-[1.75] text-text-muted">
                  Every provider has different APIs, different pricing models,
                  different availability windows. Teams waste days comparing
                  options, configuring networking, and overpaying because they
                  can&apos;t see the full market.
                </p>
                <p className="text-[15px] font-normal leading-[1.75] text-text-muted">
                  GPU.ai solves this with a single aggregation layer. We
                  continuously poll inventory across providers, normalize pricing
                  to per-GPU-hour, and let you deploy with one command.{" "}
                  <span className="font-semibold text-white">
                    Per-second billing, no minimums, no lock-in.
                  </span>
                </p>
              </div>
            </div>
          </div>
        </Container>
      </section>

      {/* ═══ Board Members ═══ */}
      <section className="border-t border-border">
        <Container className="py-24 md:py-32">
          <div className="mb-12 text-center md:mb-16">
            <p className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-purple-light">
              Board Members
            </p>
            <h2 className="type-h2 mt-3 font-bold text-white">
              Backed by industry leaders
            </h2>
          </div>

          <div className="mx-auto max-w-[640px] rounded-xl border border-border">
            <div className="grid grid-cols-1 sm:grid-cols-2">
              {BOARD_MEMBERS.map((t, i) => (
                <div
                  key={t.name}
                  className={`animate-fade-up group flex flex-col items-center px-6 py-10 text-center ${
                    i === 0 ? "border-b border-border sm:border-b-0 sm:border-r sm:border-border" : ""
                  }`}
                  style={{ animationDelay: `${0.1 + i * 0.07}s` }}
                >
                  <div className="relative mb-5 h-[120px] w-[120px] overflow-hidden rounded-full border border-border/60 transition-all duration-300 group-hover:border-purple/40 group-hover:shadow-[0_0_20px_rgba(124,107,240,0.12)]">
                    {t.photo ? (
                      <Image
                        src={t.photo}
                        alt={t.name}
                        width={400}
                        height={400}
                        className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
                      />
                    ) : (
                      <div className="flex h-full w-full items-center justify-center bg-purple-dim">
                        <span className="text-[26px] font-bold tracking-tight text-purple-light">
                          {t.name.split(" ").map((w) => w[0]).join("")}
                        </span>
                      </div>
                    )}
                  </div>

                  <h3 className="text-[15px] font-semibold text-white">{t.name}</h3>
                  <p className="mt-1 font-mono text-[12px] font-normal text-text-dim">{t.role}</p>

                  <div className="mt-3 flex items-center gap-2">
                    {t.previous && t.previous.length > 0 && (
                      <span className="flex items-center gap-1.5 font-mono text-[11px] font-normal text-text-dim">
                        {t.previous.map((p, pi) => (
                          <span key={p.label} className="flex items-center gap-1.5 text-text-muted">
                            {pi > 0 && <span className="text-text-dim/40">·</span>}
                            {p.logo ? p.logo : <span className="font-semibold">{p.label}</span>}
                          </span>
                        ))}
                      </span>
                    )}
                    {t.linkedin && (
                      <>
                        {t.previous && <span className="text-border-light">|</span>}
                        <a
                          href={t.linkedin}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex text-text-dim transition-colors hover:text-white"
                          aria-label={`${t.name} on LinkedIn`}
                        >
                          <LinkedInIcon />
                        </a>
                      </>
                    )}
                  </div>

                  <p className="mt-4 max-w-[220px] text-[12px] font-normal leading-relaxed text-text-dim">
                    {t.bio}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </Container>
      </section>

      {/* ═══ Team ═══ */}
      <section className="border-t border-border">
        <Container className="py-24 md:py-32">
          <div className="mb-12 text-center md:mb-16">
            <p className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-purple-light">
              Leadership
            </p>
            <h2 className="type-h2 mt-3 font-bold text-white">
              The team behind the compute layer
            </h2>
          </div>

          {/* Team grid */}
          <div className="mx-auto max-w-[960px] rounded-xl border border-border">
            <div className="flex flex-wrap justify-center py-8">
              {TEAM.map((t, i) => (
                <div
                  key={t.name}
                  className="animate-fade-up group flex w-full flex-col items-center px-6 py-5 text-center sm:w-1/2 lg:w-1/3"
                  style={{ animationDelay: `${0.1 + i * 0.07}s` }}
                >
                  <div className="relative mb-5 h-[120px] w-[120px] overflow-hidden rounded-full border border-border/60 transition-all duration-300 group-hover:border-purple/40 group-hover:shadow-[0_0_20px_rgba(124,107,240,0.12)]">
                    {t.photo ? (
                      <Image
                        src={t.photo}
                        alt={t.name}
                        width={400}
                        height={400}
                        className={`h-full w-full object-cover transition-transform duration-500 group-hover:scale-105 ${t.photoClassName ?? ""}`}
                      />
                    ) : (
                      <div className="flex h-full w-full items-center justify-center bg-purple-dim">
                        <span className="text-[26px] font-bold tracking-tight text-purple-light">
                          {t.name.split(" ").map((w) => w[0]).join("")}
                        </span>
                      </div>
                    )}
                  </div>

                  <h3 className="text-[15px] font-semibold text-white">{t.name}</h3>
                  <p className="mt-1 font-mono text-[12px] font-normal text-text-dim">{t.role}</p>

                  <div className="mt-3 flex items-center gap-2">
                    {t.previous && t.previous.length > 0 && (
                      <span className="flex items-center gap-1 font-mono text-[11px] font-normal text-text-dim">
                        {t.previous[0].showPrefix && (
                          <span className="text-text-dim/60">Previously</span>
                        )}
                        {t.previous.map((p) => (
                          <span key={p.label} className="flex items-center text-text-muted">
                            {p.logo ? p.logo : <span className="font-semibold">{p.label}</span>}
                          </span>
                        ))}
                      </span>
                    )}
                    {t.previous && t.linkedin && (
                      <span className="text-border-light">|</span>
                    )}
                    {t.linkedin && (
                      <a
                        href={t.linkedin}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex text-text-dim transition-colors hover:text-white"
                        aria-label={`${t.name} on LinkedIn`}
                      >
                        <LinkedInIcon />
                      </a>
                    )}
                  </div>

                  <p className="mt-4 max-w-[220px] text-[12px] font-normal leading-relaxed text-text-dim">
                    {t.bio}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </Container>
      </section>

      {/* ═══ Values ═══ */}
      <section className="border-t border-border">
        <Container className="py-24 md:py-32">
          <div className="mb-12 text-center md:mb-16">
            <p className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-purple-light">
              What We Believe
            </p>
            <h2 className="type-h2 mt-3 font-bold text-white">Our values</h2>
          </div>
          <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
            {VALUES.map((v, i) => (
              <div
                key={v.title}
                className="animate-fade-up snake-border rounded-xl border border-border bg-bg-card p-8 transition-all duration-200 hover:-translate-y-0.5 hover:border-transparent md:p-10"
                style={{ animationDelay: `${0.1 + i * 0.08}s` }}
              >
                <div className="mb-5 inline-flex h-10 w-10 items-center justify-center rounded-lg bg-purple-dim text-purple-light">
                  {v.icon}
                </div>
                <h3 className="text-[17px] font-semibold text-white">
                  {v.title}
                </h3>
                <p className="mt-3 text-[14px] font-normal leading-[1.7] text-text-muted">
                  {v.body}
                </p>
              </div>
            ))}
          </div>
        </Container>
      </section>

      {/* ═══ CTA ═══ */}
      <section className="relative overflow-hidden border-t border-border py-24 md:py-32">
        <div className="radial-glow pointer-events-none absolute inset-0" />
        <Container className="relative z-10 text-center">
          <h2 className="type-h2 font-bold text-white">
            Start deploying GPUs
            <br />
            in under 60 seconds
          </h2>
          <p className="mx-auto mt-3 max-w-[440px] text-[16px] font-normal text-text-muted">
            Try GPU.ai free — up to $100 in credits, no credit card required.
          </p>
          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Button href="/free-trial">$100 Free Trial</Button>
            <Button variant="secondary" href="/cloud/gpu-availability">
              Launch Instance
            </Button>
          </div>
        </Container>
      </section>

      <Footer />
    </div>
  );
}
