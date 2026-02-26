"use client";

import { useEffect, useState } from "react";
import Image from "next/image";

const CUSTOMERS = [
  { name: "DeepMotion", logo: "/logos/deepmotion.png", w: 160, h: 22 },
  { name: "Exabits", logo: "/logos/exabits.png", w: 140, h: 82 },
  { name: "Cornell University", logo: "/logos/cornell.png", w: 140, h: 38 },
  { name: "NYU", logo: "/logos/nyu.png", w: 40, h: 40 },
  { name: "US Dept. of Energy", logo: "/logos/doe.png", w: 150, h: 38 },
  { name: "MIT", logo: "/logos/mit.png", w: 44, h: 44 },
  { name: "DeepInfra", logo: "/logos/deepinfra.png", w: 140, h: 38 },
] as const;

export function TrustBar() {
  // Duplicate enough times for a seamless long loop
  const items = [...CUSTOMERS, ...CUSTOMERS, ...CUSTOMERS, ...CUSTOMERS];
  const [inset, setInset] = useState(80);

  useEffect(() => {
    const readInset = () => {
      const el = document.querySelector(".stripe-lines");
      if (el) {
        setInset(el.getBoundingClientRect().left);
      }
    };
    readInset();
    window.addEventListener("resize", readInset);
    return () => window.removeEventListener("resize", readInset);
  }, []);

  return (
    <section className="border-y border-border">
      <div
        className="relative h-[88px]"
        style={{ marginLeft: inset, marginRight: inset }}
      >
        <div className="absolute inset-0 overflow-hidden">
          <div className="scroll-track flex h-full w-max items-center gap-12">
            {items.map((customer, i) => (
              <Image
                key={`${customer.name}-${i}`}
                src={customer.logo}
                alt={customer.name}
                width={customer.w}
                height={customer.h}
                className="shrink-0 object-contain opacity-90 transition-opacity hover:opacity-100"
                style={{ width: customer.w, height: customer.h }}
              />
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
