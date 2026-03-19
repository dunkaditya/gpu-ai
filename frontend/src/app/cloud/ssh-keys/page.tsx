import Link from "next/link";
import { SSHKeyManager } from "@/components/cloud/SSHKeyManager";

export default function SSHKeysPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="type-h3 text-text">SSH Keys</h1>
        <Link
          href="/docs/ssh-keys"
          target="_blank"
          className="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-purple/30 bg-purple-dim text-purple-light hover:bg-purple/15 hover:border-purple/50 transition-all type-ui-sm font-medium"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="8" cy="8" r="6.5" stroke="currentColor" strokeWidth="1.2" />
            <path d="M8 7v4.5" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
            <circle cx="8" cy="5" r="0.75" fill="currentColor" />
          </svg>
          How SSH keys work
        </Link>
      </div>
      <SSHKeyManager />
    </div>
  );
}
