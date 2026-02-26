"use client";

import { useEffect } from "react";

export interface EffectsState {
  glow: boolean;
  cardGlow: boolean;
  cursorInteract: boolean;
  scrollIndicator: boolean;
  gradientButtons: boolean;
}

const STATE: EffectsState = {
  glow: true,
  cardGlow: true,
  cursorInteract: true,
  scrollIndicator: false,
  gradientButtons: true,
};

declare global {
  interface Window {
    __effects: EffectsState;
  }
}

/** Initializes global effects state — no UI needed anymore */
export function EffectsToggle() {
  useEffect(() => {
    window.__effects = STATE;
  }, []);

  return null;
}
