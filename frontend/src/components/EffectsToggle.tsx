"use client";

import { useState, useEffect } from "react";

export interface EffectsState {
  glow: boolean;
  cardGlow: boolean;
  cursorInteract: boolean;
  scrollIndicator: boolean;
  gradientButtons: boolean;
  nodes: boolean;
  meshGradient: boolean;
  containEffects: boolean;
}

const DEFAULTS: EffectsState = {
  glow: true,
  cardGlow: true,
  cursorInteract: true,
  scrollIndicator: false,
  gradientButtons: true,
  nodes: true,
  meshGradient: true,
  containEffects: false,
};

declare global {
  interface Window {
    __effects: EffectsState;
  }
}

function broadcast(state: EffectsState) {
  window.__effects = state;
  window.dispatchEvent(new CustomEvent("effects-change", { detail: state }));
}

export function EffectsToggle() {
  const [nodes, setNodes] = useState(DEFAULTS.nodes);
  const [mesh, setMesh] = useState(DEFAULTS.meshGradient);
  useEffect(() => {
    window.__effects = DEFAULTS;
  }, []);

  const toggleNodes = () => {
    setNodes((prev) => {
      const next = !prev;
      broadcast({ ...window.__effects, nodes: next });
      return next;
    });
  };

  const toggleMesh = () => {
    setMesh((prev) => {
      const next = !prev;
      broadcast({ ...window.__effects, meshGradient: next });
      return next;
    });
  };

  return (
    <div className="fixed bottom-4 right-4 z-[100] flex items-center gap-1 rounded-full border border-border-light bg-bg-card/95 px-1 py-1 shadow-2xl backdrop-blur-md">
      <button
        onClick={toggleMesh}
        className={`rounded-full px-3 py-1.5 text-[11px] font-medium transition-all ${
          mesh
            ? "bg-purple/15 text-purple-light"
            : "text-text-dim hover:text-text-muted"
        }`}
      >
        Mesh {mesh ? "ON" : "OFF"}
      </button>
      <div className="h-3 w-px bg-border-light" />
      <button
        onClick={toggleNodes}
        className={`rounded-full px-3 py-1.5 text-[11px] font-medium transition-all ${
          nodes
            ? "bg-purple/15 text-purple-light"
            : "text-text-dim hover:text-text-muted"
        }`}
      >
        Nodes {nodes ? "ON" : "OFF"}
      </button>
    </div>
  );
}
