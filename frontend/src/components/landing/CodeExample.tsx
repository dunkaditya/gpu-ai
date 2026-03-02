"use client";

import { useState } from "react";
import { Section } from "@/components/ui/Section";
import { SectionLabel } from "@/components/ui/SectionLabel";
import { Button } from "@/components/ui/Button";
import { CODE_EXAMPLE } from "@/lib/constants";

function SyntaxHighlight({ code }: { code: string }) {
  const lines = code.split("\n");

  const highlight = (line: string) => {
    return line
      // Comments
      .replace(/(#.*)$/g, '<span class="code-comment">$1</span>')
      // Strings
      .replace(
        /("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')/g,
        '<span class="code-string">$1</span>',
      )
      // Keywords
      .replace(
        /\b(from|import|print)\b/g,
        '<span class="code-keyword">$1</span>',
      )
      // Function/method calls
      .replace(
        /\.(\w+)\(/g,
        '.<span class="code-function">$1</span>(',
      )
      // Named params
      .replace(
        /(\w+)=/g,
        '<span class="code-param">$1</span>=',
      )
      // Numbers
      .replace(
        /\b(\d+)\b/g,
        '<span class="code-number">$1</span>',
      );
  };

  return (
    <pre className="overflow-x-auto text-[13px] leading-[1.8] md:text-[14px]">
      <code>
        {lines.map((line, i) => (
          <div
            key={i}
            className="whitespace-pre"
            dangerouslySetInnerHTML={{ __html: highlight(line) || "\u00A0" }}
          />
        ))}
      </code>
    </pre>
  );
}

export function CodeExample() {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(CODE_EXAMPLE);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Section>
      <div className="grid items-center gap-12 md:grid-cols-[2fr_3fr] md:gap-16">
        {/* Left description */}
        <div>
          <SectionLabel>Developer First</SectionLabel>
          <h2 className="type-h2 mt-3 font-bold text-white">
            One API call to deploy
          </h2>
          <p className="type-body-lg mt-3 text-text-muted">
            Our Python SDK and REST API let you provision GPUs
            programmatically. Integrate into your CI/CD pipeline, training
            scripts, or internal tooling in minutes.
          </p>
          <div className="mt-8">
            <Button href="/docs" variant="secondary">
              Read the docs <span className="ml-1">→</span>
            </Button>
          </div>
        </div>

        {/* Right code block */}
        <div className="overflow-hidden rounded-xl border border-border bg-bg-card">
          {/* Mac chrome */}
          <div className="flex items-center justify-between border-b border-border px-4 py-3">
            <div className="flex items-center gap-3">
              <div className="flex gap-1.5">
                <span className="h-3 w-3 rounded-full bg-[#ff5f56]" />
                <span className="h-3 w-3 rounded-full bg-[#ffbd2e]" />
                <span className="h-3 w-3 rounded-full bg-[#27c93f]" />
              </div>
              <span className="type-ui-xs text-text-dim">deploy.py</span>
            </div>
            <button
              onClick={handleCopy}
              className="type-ui-2xs rounded-md px-3 py-1 text-text-dim transition-colors hover:bg-bg-card-hover hover:text-text-muted"
            >
              {copied ? "Copied!" : "Copy"}
            </button>
          </div>

          {/* Code */}
          <div className="p-5 text-text">
            <SyntaxHighlight code={CODE_EXAMPLE} />
          </div>
        </div>
      </div>
    </Section>
  );
}
