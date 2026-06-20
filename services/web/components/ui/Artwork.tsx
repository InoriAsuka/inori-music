"use client";

import { Play } from "lucide-react";
import { cn } from "@/lib/utils";

interface ArtworkProps {
  src?: string | null;
  alt: string;
  size?: "sm" | "md" | "lg";
  /** Show play button overlay on hover. */
  playable?: boolean;
  onPlay?: () => void;
  className?: string;
}

const SIZE_CLASSES = {
  sm: "h-10 w-10",
  md: "h-14 w-14",
  lg: "h-48 w-48",
};

export function Artwork({ src, alt, size = "md", playable, onPlay, className }: ArtworkProps) {
  return (
    <div
      className={cn(
        "relative shrink-0 overflow-hidden rounded-md bg-[var(--color-muted)]",
        SIZE_CLASSES[size],
        playable && "group cursor-pointer",
        className
      )}
      onClick={onPlay}
    >
      {src ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img src={src} alt={alt} className="h-full w-full object-cover" />
      ) : (
        <div className="flex h-full w-full items-center justify-center text-[var(--color-muted-foreground)]">
          <svg width="40%" height="40%" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
          </svg>
        </div>
      )}

      {/* Play overlay */}
      {playable && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/40 opacity-0 transition-opacity group-hover:opacity-100">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--color-primary)]">
            <Play size={14} fill="white" className="text-white ml-0.5" />
          </div>
        </div>
      )}
    </div>
  );
}
