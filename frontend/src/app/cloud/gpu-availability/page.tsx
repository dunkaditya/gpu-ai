"use client";

import { GPUAvailabilityTable } from "@/components/cloud/GPUAvailabilityTable";

export default function GPUAvailabilityPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="type-h3 text-text">GPU Availability</h1>
      </div>
      <GPUAvailabilityTable />
    </div>
  );
}
