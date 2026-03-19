import type { Metadata } from "next";
import { ClerkProvider } from "@clerk/nextjs";
import { dark } from "@clerk/themes";
import localFont from "next/font/local";
import "./globals.css";

const vremenaGrotesk = localFont({
  src: "../../public/fonts/vremena-grotesk.woff2",
  variable: "--font-vremena-grotesk",
  display: "swap",
});

const nectoMono = localFont({
  src: "../../public/fonts/necto-mono.woff2",
  variable: "--font-necto-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: {
    template: "%s | GPU.ai",
    default: "GPU.ai -- GPU Cloud at Unbeatable Prices",
  },
  description:
    "We aggregate GPU inventory from 12+ providers to find you the best price. Up to 30% cheaper than hyperscalers. Deploy in seconds.",
  openGraph: {
    title: "GPU.ai -- GPU Cloud at Unbeatable Prices",
    description:
      "Aggregate GPU cloud. Up to 30% cheaper. Deploy in seconds.",
    type: "website",
  },
};

const clerkKey = process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const inner = (
    <html lang="en" className="dark">
      <body
        className={`${vremenaGrotesk.variable} ${nectoMono.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );

  if (clerkKey) {
    return (
      <ClerkProvider
        localization={{
          signIn: {
            start: {
              title: "Sign in",
              subtitle: "",
            },
          },
          signUp: {
            start: {
              title: "Create your account",
              subtitle: "",
            },
          },
        }}
        appearance={{
          baseTheme: dark,
          layout: {
            logoPlacement: "none",
          },
          variables: {
            colorPrimary: "#7c6bf0",
            colorBackground: "#0a0a0a",
            colorInputBackground: "#111111",
            colorInputText: "#ededed",
            colorText: "#ededed",
            colorTextSecondary: "#9a9a9a",
            borderRadius: "10px",
            fontFamily: "var(--font-necto-mono), monospace",
          },
          elements: {
            card: {
              backgroundColor: "#0a0a0a",
              border: "1px solid #333333",
              boxShadow: "0 0 40px rgba(124, 107, 240, 0.08)",
            },
            headerTitle: {
              fontFamily: "var(--font-vremena-grotesk), sans-serif",
              letterSpacing: "-0.02em",
            },
            headerSubtitle: {
              color: "#9a9a9a",
            },
            socialButtonsBlockButton: {
              backgroundColor: "#111111",
              border: "1px solid #333333",
              color: "#ededed",
              "&:hover": {
                backgroundColor: "#1a1a1a",
                borderColor: "#444444",
              },
            },
            formFieldInput: {
              backgroundColor: "#111111",
              border: "1px solid #333333",
              color: "#ededed",
              "&:focus": {
                borderColor: "#7c6bf0",
                boxShadow: "0 0 0 2px rgba(124, 107, 240, 0.25)",
              },
            },
            formButtonPrimary: {
              background:
                "linear-gradient(135deg, #6b5ce0 0%, #5244c4 60%, #4a3ab8 100%)",
              "&:hover": {
                background:
                  "linear-gradient(135deg, #7c6bf0 0%, #6b5ce0 60%, #5244c4 100%)",
                boxShadow: "0 0 28px rgba(124, 107, 240, 0.35)",
              },
            },
            footerActionLink: {
              color: "#7c6bf0",
              "&:hover": {
                color: "#9a8cff",
              },
            },
            dividerLine: {
              backgroundColor: "#333333",
            },
            dividerText: {
              color: "#707070",
            },
            identityPreviewEditButton: {
              color: "#7c6bf0",
            },
          },
        }}
      >
        {inner}
      </ClerkProvider>
    );
  }

  return inner;
}
