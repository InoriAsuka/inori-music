#!/usr/bin/env node
/**
 * scripts/gen-icons.mjs — Generate PWA icons (192 × 192, 512 × 512)
 * Requires: npm install --save-dev sharp
 *
 * Usage: node scripts/gen-icons.mjs
 *
 * Generates a simple Neon Shrine gradient icon — no external SVG required.
 * Replace the canvas drawing code with your actual artwork once available.
 */

import { createCanvas } from "canvas";
import { writeFileSync, mkdirSync } from "node:fs";
import { join } from "node:path";

const sizes = [192, 512];
const outDir = join(process.cwd(), "public", "icons");
mkdirSync(outDir, { recursive: true });

for (const size of sizes) {
  const canvas = createCanvas(size, size);
  const ctx = canvas.getContext("2d");

  // Background
  const bg = ctx.createLinearGradient(0, 0, size, size);
  bg.addColorStop(0, "#070711");
  bg.addColorStop(1, "#12122b");
  ctx.fillStyle = bg;
  ctx.beginPath();
  ctx.roundRect(0, 0, size, size, size * 0.2);
  ctx.fill();

  // Glow ring
  const ring = ctx.createRadialGradient(size / 2, size / 2, size * 0.1, size / 2, size / 2, size * 0.45);
  ring.addColorStop(0, "rgba(155, 92, 255, 0.25)");
  ring.addColorStop(1, "rgba(15, 212, 192, 0.05)");
  ctx.fillStyle = ring;
  ctx.fillRect(0, 0, size, size);

  // Music note icon (simple, no external font)
  const s = size * 0.45;
  const cx = size / 2 - s * 0.15;
  const cy = size / 2 + s * 0.05;

  ctx.strokeStyle = "#9b5cff";
  ctx.lineWidth = size * 0.055;
  ctx.lineCap = "round";

  // Stem
  ctx.beginPath();
  ctx.moveTo(cx + s * 0.38, cy - s * 0.55);
  ctx.lineTo(cx + s * 0.38, cy + s * 0.05);
  ctx.stroke();

  // Note head
  ctx.fillStyle = "#9b5cff";
  ctx.beginPath();
  ctx.ellipse(cx + s * 0.22, cy + s * 0.13, s * 0.2, s * 0.15, -0.4, 0, Math.PI * 2);
  ctx.fill();

  const buf = canvas.toBuffer("image/png");
  writeFileSync(join(outDir, `icon-${size}.png`), buf);
  console.log(`✓ ${size}×${size} → public/icons/icon-${size}.png`);
}
