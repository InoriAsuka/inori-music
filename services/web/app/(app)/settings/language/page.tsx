"use client";

import { useEffect, useState } from "react";
import { Check } from "lucide-react";
import { SUPPORTED_LANGS, setLanguage } from "@/lib/i18n";
import { cn } from "@/lib/utils";

export default function LanguagePage() {
  const [current, setCurrent] = useState("en");

  useEffect(() => {
    if (typeof window !== "undefined") {
      setCurrent(localStorage.getItem("inori_lang") ?? "en");
    }
  }, []);

  function pick(code: string) {
    setLanguage(code as "en" | "zh" | "ja");
    setCurrent(code);
  }

  return (
    <div className="mx-auto max-w-md space-y-6">
      <h1 className="text-2xl font-bold text-[var(--color-text)]">Language</h1>

      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] divide-y divide-[var(--color-border)]">
        {SUPPORTED_LANGS.map(({ code, label, nativeLabel }) => (
          <button
            key={code}
            onClick={() => pick(code)}
            className={cn(
              "flex w-full items-center justify-between px-4 py-3.5 text-sm transition-colors hover:bg-[var(--color-surface-raised)]",
              current === code ? "text-[var(--color-primary)]" : "text-[var(--color-text)]"
            )}
          >
            <span>
              <span className="font-medium">{nativeLabel}</span>
              <span className="ml-2 text-[var(--color-text-muted)]">{label}</span>
            </span>
            {current === code && <Check size={16} className="text-[var(--color-primary)]" />}
          </button>
        ))}
      </div>
    </div>
  );
}
