"use client";

import { useEffect, useRef } from "react";
import { usePlayerStore } from "@/store/player";

/**
 * Lightweight 64-bar visualizer. It uses deterministic animation while playing;
 * future phase can wire this to Web Audio AnalyserNode once the audio element is shared.
 */
export function Visualizer() {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const status = usePlayerStore((s) => s.status);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    let raf = 0;
    let t = 0;

    function draw() {
      if (!canvas || !ctx) return;
      const dpr = window.devicePixelRatio || 1;
      const w = canvas.clientWidth;
      const h = canvas.clientHeight;
      canvas.width = Math.floor(w * dpr);
      canvas.height = Math.floor(h * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      ctx.clearRect(0, 0, w, h);

      const bars = 64;
      const gap = 2;
      const bw = (w - gap * (bars - 1)) / bars;
      const grad = ctx.createLinearGradient(0, 0, w, 0);
      grad.addColorStop(0, "#9b5cff");
      grad.addColorStop(0.5, "#ff5fa0");
      grad.addColorStop(1, "#0fd4c0");
      ctx.fillStyle = grad;

      for (let i = 0; i < bars; i++) {
        const wave = Math.sin(t * 0.04 + i * 0.35) * 0.5 + 0.5;
        const pulse = Math.sin(t * 0.09 + i * 0.17) * 0.5 + 0.5;
        const amp = status === "playing" ? Math.max(0.15, wave * 0.7 + pulse * 0.3) : 0.08;
        const bh = amp * h;
        ctx.globalAlpha = status === "playing" ? 0.8 : 0.25;
        ctx.fillRect(i * (bw + gap), h - bh, bw, bh);
      }
      t++;
      raf = requestAnimationFrame(draw);
    }

    draw();
    return () => cancelAnimationFrame(raf);
  }, [status]);

  return <canvas ref={canvasRef} className="h-8 w-full opacity-90" aria-hidden="true" />;
}
