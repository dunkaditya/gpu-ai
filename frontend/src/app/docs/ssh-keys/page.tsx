import Link from "next/link";
import type { Metadata } from "next";
import { ChipLogo } from "@/components/ui/ChipLogo";

export const metadata: Metadata = {
  title: "SSH Keys",
  description: "Learn how to generate and add SSH keys to GPU.ai",
};

function CodeBlock({ children }: { children: string }) {
  return (
    <pre className="bg-bg-alt border border-border rounded-lg px-5 py-4 overflow-x-auto">
      <code className="font-mono text-sm text-purple-light">{children}</code>
    </pre>
  );
}

export default function SSHKeysDocsPage() {
  return (
    <div className="min-h-screen bg-bg">
      {/* Nav */}
      <header className="border-b border-border">
        <div className="max-w-3xl mx-auto px-6 h-14 flex items-center justify-between">
          <Link href="/" className="flex items-center gap-0.5">
            <ChipLogo size={24} />
            <span className="font-sans text-[16px] font-bold tracking-[-0.5px]">
              <span className="text-white">gpu</span>
              <span className="gradient-text">.ai</span>
            </span>
            <span className="ml-2 type-ui-xs text-text-dim">/docs</span>
          </Link>
          <Link
            href="/cloud/ssh-keys"
            className="type-ui-xs text-purple hover:text-purple-light transition-colors"
          >
            Back to Dashboard
          </Link>
        </div>
      </header>

      {/* Content */}
      <main className="max-w-3xl mx-auto px-6 py-12">
        <article className="space-y-10">
          {/* Title */}
          <div className="space-y-3">
            <h1 className="type-h3 text-text font-sans">SSH Keys</h1>
            <p className="type-body-lg text-text-muted leading-relaxed">
              SSH keys let you securely connect to your GPU instances without a
              password. You need at least one SSH key before launching an
              instance.
            </p>
          </div>

          {/* Why */}
          <section className="space-y-4">
            <h2 className="type-h5 text-text font-sans">
              Why are SSH keys required?
            </h2>
            <p className="type-ui-sm text-text-muted leading-relaxed">
              When you launch a GPU instance, your public SSH keys are
              injected into the instance during setup. This is the only way
              to authenticate &mdash; there are no passwords. This approach is
              more secure, eliminates credential management, and lets you
              connect from any machine that has the matching private key.
            </p>
          </section>

          {/* Generate */}
          <section className="space-y-4">
            <h2 className="type-h5 text-text font-sans">
              Generate a new SSH key
            </h2>
            <p className="type-ui-sm text-text-muted leading-relaxed">
              If you don&apos;t already have an SSH key, generate one with:
            </p>
            <CodeBlock>ssh-keygen -t ed25519</CodeBlock>
            <p className="type-ui-sm text-text-muted leading-relaxed">
              Press Enter to accept the default file location{" "}
              <code className="font-mono text-text bg-bg-alt px-1.5 py-0.5 rounded text-xs">
                ~/.ssh/id_ed25519
              </code>
              . You can optionally set a passphrase for extra security.
            </p>
            <p className="type-ui-sm text-text-muted leading-relaxed">
              This creates two files:
            </p>
            <ul className="space-y-2 type-ui-sm text-text-muted">
              <li className="flex gap-3">
                <code className="font-mono text-text bg-bg-alt px-1.5 py-0.5 rounded text-xs shrink-0">
                  ~/.ssh/id_ed25519
                </code>
                <span>&mdash; your private key (never share this)</span>
              </li>
              <li className="flex gap-3">
                <code className="font-mono text-text bg-bg-alt px-1.5 py-0.5 rounded text-xs shrink-0">
                  ~/.ssh/id_ed25519.pub
                </code>
                <span>&mdash; your public key (this is what you add to GPU.ai)</span>
              </li>
            </ul>
          </section>

          {/* Copy */}
          <section className="space-y-4">
            <h2 className="type-h5 text-text font-sans">
              Copy your public key
            </h2>
            <div className="space-y-3">
              <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider">
                macOS
              </p>
              <CodeBlock>pbcopy &lt; ~/.ssh/id_ed25519.pub</CodeBlock>

              <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mt-4">
                Linux
              </p>
              <CodeBlock>cat ~/.ssh/id_ed25519.pub</CodeBlock>

              <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mt-4">
                Windows (PowerShell)
              </p>
              <CodeBlock>Get-Content ~/.ssh/id_ed25519.pub | Set-Clipboard</CodeBlock>
            </div>
          </section>

          {/* Add to GPU.ai */}
          <section className="space-y-4">
            <h2 className="type-h5 text-text font-sans">
              Add the key to GPU.ai
            </h2>
            <ol className="space-y-3 type-ui-sm text-text-muted list-decimal list-inside">
              <li>
                Go to{" "}
                <Link
                  href="/cloud/ssh-keys"
                  className="text-purple hover:text-purple-light transition-colors"
                >
                  SSH Keys
                </Link>{" "}
                in the dashboard
              </li>
              <li>Click <strong className="text-text">Add Key</strong></li>
              <li>Give it a name (e.g. &ldquo;work-laptop&rdquo;)</li>
              <li>Paste your public key and save</li>
            </ol>
            <p className="type-ui-sm text-text-muted leading-relaxed">
              All your SSH keys are automatically attached to new instances
              when you launch them.
            </p>
          </section>

          {/* Connect */}
          <section className="space-y-4">
            <h2 className="type-h5 text-text font-sans">
              Connect to an instance
            </h2>
            <p className="type-ui-sm text-text-muted leading-relaxed">
              Once your instance is running, use the SSH command shown in the
              instances table:
            </p>
            <CodeBlock>ssh -p 10000 root@your-server-ip</CodeBlock>
            <p className="type-ui-sm text-text-muted leading-relaxed">
              The exact command (with port and host) is shown in the SSH
              Command column and can be copied with one click.
            </p>
          </section>

          {/* Existing key */}
          <section className="space-y-4 border-t border-border pt-10">
            <h2 className="type-h5 text-text font-sans">
              Already have an SSH key?
            </h2>
            <p className="type-ui-sm text-text-muted leading-relaxed">
              If you already use SSH (e.g. with GitHub), you likely already
              have a key. Check with:
            </p>
            <CodeBlock>ls ~/.ssh/*.pub</CodeBlock>
            <p className="type-ui-sm text-text-muted leading-relaxed">
              If you see a file like{" "}
              <code className="font-mono text-text bg-bg-alt px-1.5 py-0.5 rounded text-xs">
                id_ed25519.pub
              </code>{" "}
              or{" "}
              <code className="font-mono text-text bg-bg-alt px-1.5 py-0.5 rounded text-xs">
                id_rsa.pub
              </code>
              , you can use that. Just copy its contents and add it to GPU.ai.
            </p>
          </section>
        </article>
      </main>
    </div>
  );
}
