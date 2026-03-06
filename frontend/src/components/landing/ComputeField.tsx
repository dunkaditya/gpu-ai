"use client";

import { useEffect, useRef } from "react";

interface Node {
  x: number;
  y: number;
  vx: number;
  vy: number;
  radius: number;
  color: string;
  pulse: number;
  pulseSpeed: number;
  depth: number; // 0 = far back, 1 = foreground
}

const NAV_HEIGHT = 88;

export function ComputeField() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    ).matches;
    if (prefersReduced) return;

    let animId: number;
    let scrollY = 0;
    let mouseX = -9999;
    let mouseY = -9999;
    const isMobile = window.innerWidth < 768;
    const nodeCount = isMobile ? 80 : 240;

    // Depth-aware colors: far nodes are dimmer
    const baseColors = [
      [154, 140, 255],
      [139, 122, 255],
      [100, 80, 220],
      [120, 107, 240],
      [180, 168, 255],
    ];

    const getBorders = (): { left: number; right: number } => {
      return { left: 0, right: window.innerWidth };
    };

    // Spread nodes across a large area — they just float freely
    const SPREAD = 2000;
    const nodes: Node[] = Array.from({ length: nodeCount }, () => {
      const depth = Math.random(); // 0 = far, 1 = near
      const rgb = baseColors[Math.floor(Math.random() * baseColors.length)];
      // Opacity scales with depth: far = 0.15–0.3, near = 0.7–0.95
      const alpha = 0.15 + depth * 0.8;
      return {
        x: (Math.random() - 0.5) * SPREAD + window.innerWidth / 2,
        y: (Math.random() - 0.5) * SPREAD + window.innerHeight / 2,
        // Far nodes move slower, near nodes faster
        vx: (Math.random() - 0.5) * 0.3 * (0.3 + depth * 0.7),
        vy: (Math.random() - 0.5) * 0.3 * (0.3 + depth * 0.7),
        // Far nodes smaller, near nodes larger
        radius: 0.5 + depth * 2.5,
        color: `rgba(${rgb[0]}, ${rgb[1]}, ${rgb[2]}, ${alpha.toFixed(2)})`,
        pulse: Math.random() * Math.PI * 2,
        pulseSpeed: 0.005 + depth * 0.02,
        depth,
      };
    });

    // Sort so far nodes draw first (behind)
    nodes.sort((a, b) => a.depth - b.depth);

    const connectionDist = isMobile ? 120 : 180;
    const cursorDist = 120;

    const resize = () => {
      const dpr = Math.min(window.devicePixelRatio, 2);
      canvas.width = window.innerWidth * dpr;
      canvas.height = window.innerHeight * dpr;
      canvas.style.width = `${window.innerWidth}px`;
      canvas.style.height = `${window.innerHeight}px`;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };

    const onScroll = () => {
      scrollY = window.scrollY;
    };

    const onMouse = (e: MouseEvent) => {
      mouseX = e.clientX;
      mouseY = e.clientY + scrollY;
    };

    const onMouseLeave = () => {
      mouseX = -9999;
      mouseY = -9999;
    };

    resize();
    window.addEventListener("resize", resize);
    window.addEventListener("scroll", onScroll, { passive: true });
    window.addEventListener("mousemove", onMouse, { passive: true });
    document.addEventListener("mouseleave", onMouseLeave);

    const draw = () => {
      const w = window.innerWidth;
      const h = window.innerHeight;
      ctx.clearRect(0, 0, w, h);

      const borders = getBorders();
      const visibleTop = scrollY + NAV_HEIGHT;
      const visibleBottom = scrollY + h;
      const cursorInteract = true;

      // Parallax: each node's rendered Y shifts based on scroll and depth
      // Far nodes (depth≈0) move less with scroll, near nodes (depth≈1) move more
      const parallaxOffset = (depth: number) => scrollY * (depth * 0.15);

      const screenY = (node: Node) =>
        node.y - scrollY + parallaxOffset(node.depth);

      const isVisible = (node: Node) => {
        const sy = screenY(node);
        return (
          node.x >= borders.left &&
          node.x <= borders.right &&
          sy >= NAV_HEIGHT &&
          sy <= h
        );
      };

      // Viewport-relative mouse position
      const mxView = mouseX;
      const myView = mouseY;

      // Draw connections between visible nodes (only connect same-ish depth)
      for (let i = 0; i < nodes.length; i++) {
        if (!isVisible(nodes[i])) continue;
        const sy1 = screenY(nodes[i]);
        for (let j = i + 1; j < nodes.length; j++) {
          if (!isVisible(nodes[j])) continue;
          // Only connect nodes within similar depth (±0.3)
          if (Math.abs(nodes[i].depth - nodes[j].depth) > 0.3) continue;
          const sy2 = screenY(nodes[j]);
          const dx = nodes[i].x - nodes[j].x;
          const dy = sy1 - sy2;
          const dist = Math.sqrt(dx * dx + dy * dy);

          if (dist < connectionDist) {
            const avgDepth = (nodes[i].depth + nodes[j].depth) / 2;
            const alpha = (1 - dist / connectionDist) * (0.15 + avgDepth * 0.55);
            ctx.beginPath();
            ctx.moveTo(nodes[i].x, sy1);
            ctx.lineTo(nodes[j].x, sy2);
            ctx.strokeStyle = `rgba(154, 140, 255, ${alpha})`;
            ctx.lineWidth = 0.3 + avgDepth * 0.7;
            ctx.stroke();
          }
        }
      }

      // Draw cursor connections when enabled
      if (cursorInteract && mouseX > 0) {
        for (const node of nodes) {
          if (!isVisible(node)) continue;
          const sy = screenY(node);
          const dx = node.x - mxView;
          const dy = node.y - myView;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < cursorDist) {
            const alpha = (1 - dist / cursorDist) * 0.2 * node.depth;
            ctx.beginPath();
            ctx.moveTo(node.x, sy);
            ctx.lineTo(mxView, mouseY - scrollY);
            ctx.strokeStyle = `rgba(154, 140, 255, ${alpha})`;
            ctx.lineWidth = 0.5;
            ctx.stroke();
          }
        }
      }

      // Move all nodes, wrap if too far off-screen, only draw visible ones
      const margin = 200;
      for (const node of nodes) {
        // Cursor repulsion when enabled (stronger for near nodes)
        if (cursorInteract && mouseX > 0) {
          const dx = node.x - mxView;
          const dy = node.y - myView;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < cursorDist && dist > 0) {
            const force = (1 - dist / cursorDist) * 0.008 * node.depth;
            node.vx += (dx / dist) * force;
            node.vy += (dy / dist) * force;
          }
        }

        // Small random drift — slower for far nodes
        const driftScale = 0.005 + node.depth * 0.015;
        node.vx += (Math.random() - 0.5) * driftScale;
        node.vy += (Math.random() - 0.5) * driftScale;

        // Dampen velocity
        node.vx *= 0.998;
        node.vy *= 0.998;

        node.x += node.vx;
        node.y += node.vy;

        // Wrap nodes that drift too far off-screen back to the opposite side
        if (node.x < -margin) node.x = w + margin;
        if (node.x > w + margin) node.x = -margin;
        if (node.y < -margin) node.y = h + margin;
        if (node.y > h + margin) node.y = -margin;

        if (!isVisible(node)) continue;

        const sy = screenY(node);

        node.pulse += node.pulseSpeed;
        const pulseScale = 1 + Math.sin(node.pulse) * 0.3;
        const r = node.radius * pulseScale;

        // Cursor proximity glow boost
        let glowMulti = 1;
        if (cursorInteract && mouseX > 0) {
          const dx = node.x - mxView;
          const dy = node.y - myView;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < cursorDist) {
            glowMulti = 1 + (1 - dist / cursorDist) * 0.6 * node.depth;
          }
        }

        // Glow — larger for foreground nodes
        const glowRadius = r * (2 + node.depth * 3) * glowMulti;
        const glowAlpha = (0.05 + node.depth * 0.15) * glowMulti;
        ctx.beginPath();
        ctx.arc(node.x, sy, glowRadius, 0, Math.PI * 2);
        ctx.fillStyle = node.color.replace(/[\d.]+\)$/, `${glowAlpha})`);
        ctx.fill();

        // Core
        ctx.beginPath();
        ctx.arc(node.x, sy, r * Math.min(glowMulti, 1.5), 0, Math.PI * 2);
        ctx.fillStyle = node.color;
        ctx.fill();
      }

      animId = requestAnimationFrame(draw);
    };

    draw();

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener("resize", resize);
      window.removeEventListener("scroll", onScroll);
      window.removeEventListener("mousemove", onMouse);
      document.removeEventListener("mouseleave", onMouseLeave);
    };
  }, []);

  return (
    <div className="pointer-events-none absolute inset-0 z-[43] overflow-hidden">
      <canvas ref={canvasRef} className="absolute inset-0" />
    </div>
  );
}
