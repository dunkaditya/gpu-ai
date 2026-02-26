"use client";

import { useState } from "react";
import { USE_CASE_TABS } from "@/lib/constants";
import { Container } from "@/components/ui";
import { cn } from "@/lib/utils";

export function UseCaseTabs() {
  const [activeIndex, setActiveIndex] = useState(0);
  const activeTab = USE_CASE_TABS[activeIndex];

  return (
    <section className="py-24">
      <Container>
        <h2 className="text-3xl md:text-4xl font-bold text-white text-center mb-12">
          Built for every GPU workload
        </h2>

        {/* Tab row */}
        <div className="flex border-b border-white/10 justify-center overflow-x-auto">
          {USE_CASE_TABS.map((tab, index) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveIndex(index)}
              className={cn(
                "px-5 py-3 text-sm font-medium transition-colors whitespace-nowrap",
                index === activeIndex
                  ? "border-b-2 border-white text-white"
                  : "text-gray-400 hover:text-gray-200"
              )}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* Content panel */}
        <div className="mt-12">
          <h3 className="text-2xl md:text-3xl font-bold text-white">
            {activeTab.title}
          </h3>
          <p className="mt-4 text-text-muted text-lg max-w-[700px]">
            {activeTab.description}
          </p>
          <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-4">
            {activeTab.features.map((feature) => (
              <div key={feature} className="flex items-start gap-3">
                <svg
                  className="text-green-400 mt-1 shrink-0 w-4 h-4"
                  viewBox="0 0 16 16"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    d="M13.5 4.5L6 12L2.5 8.5"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
                <span className="text-sm text-text-muted">{feature}</span>
              </div>
            ))}
          </div>
        </div>
      </Container>
    </section>
  );
}
