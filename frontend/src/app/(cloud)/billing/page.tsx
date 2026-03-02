import { BillingDashboard } from "@/components/cloud/BillingDashboard";

export default function BillingPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="type-h3 text-text">Billing</h1>
      </div>
      <BillingDashboard />
    </div>
  );
}
