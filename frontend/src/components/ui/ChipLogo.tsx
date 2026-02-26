"use client";

import { useEffect, useState } from "react";

export function ChipLogo({ size = 48 }: { size?: number }) {
  const [spin, setSpin] = useState(true);

  useEffect(() => {
    const t = setTimeout(() => setSpin(false), 3000);
    return () => clearTimeout(t);
  }, []);

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 48 48"
      fill="none"
      style={spin ? { animation: "spin-down 3s ease-out forwards" } : undefined}
    >
      {[16, 24, 32].map((v) => (
        <g key={v}>
          <rect x={v - 2} y="7" width="4" height="4.5" rx="1.5" fill="#a99bff" opacity="0.55" />
          <rect x={v - 2} y="36.5" width="4" height="4.5" rx="1.5" fill="#a99bff" opacity="0.55" />
          <rect x="7" y={v - 2} width="4.5" height="4" rx="1.5" fill="#a99bff" opacity="0.55" />
          <rect x="36.5" y={v - 2} width="4.5" height="4" rx="1.5" fill="#a99bff" opacity="0.55" />
        </g>
      ))}
      <rect x="12.5" y="12.5" width="23" height="23" rx="5" fill="#8b7aff" opacity="0.8" />
      <rect x="17.5" y="17.5" width="13" height="13" rx="2.5" fill="#c5bcff" opacity="0.5" />
      <rect x="21" y="21" width="6" height="6" rx="1.5" fill="#fff" opacity="0.85" />
    </svg>
  );
}
