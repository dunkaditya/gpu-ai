import { SSHKeyManager } from "@/components/cloud/SSHKeyManager";

export default function SSHKeysPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="type-h3 text-text">SSH Keys</h1>
      </div>
      <SSHKeyManager />
    </div>
  );
}
