import type { Metadata } from "next";
import { MOCK_INSTANCES } from "@/lib/mock-data";
import { InstancesTable } from "@/components/cloud/InstancesTable";

export const metadata: Metadata = {
  title: "Instances",
};

export default function InstancesPage() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="type-h3 text-text">Instances</h1>
        <button className="gradient-btn px-4 py-2 rounded-lg type-ui-sm font-medium transition-all">
          Launch Instance
        </button>
      </div>

      {/* Table */}
      <div className="rounded-lg border border-border bg-bg-card/50 overflow-hidden">
        <InstancesTable instances={MOCK_INSTANCES} />
      </div>
    </div>
  );
}
