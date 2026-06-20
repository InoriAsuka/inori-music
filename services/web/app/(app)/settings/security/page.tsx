/**
 * settings/security — Change password page
 */
"use client";

import { useState, FormEvent } from "react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { Loader2, CheckCircle } from "lucide-react";

export default function ChangePasswordPage() {
  const token = useAuthStore((s) => s.token);
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    if (next !== confirm) { setError("Passwords do not match."); return; }
    if (next.length < 8) { setError("Password must be at least 8 characters."); return; }
    if (!token) return;

    setLoading(true);
    try {
      const { error: apiError } = await authedApi(token).POST("/api/v1/me/change-password", {
        body: { currentPassword: current, newPassword: next },
      });
      if (apiError) { setError("Incorrect current password."); return; }
      setSuccess(true);
      setCurrent(""); setNext(""); setConfirm("");
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto max-w-md space-y-6">
      <h1 className="text-2xl font-bold">Change Password</h1>

      <form onSubmit={handleSubmit} className="space-y-4 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-6">
        {[
          { id: "current", label: "Current password", value: current, set: setCurrent, autoComplete: "current-password" },
          { id: "next", label: "New password", value: next, set: setNext, autoComplete: "new-password" },
          { id: "confirm", label: "Confirm new password", value: confirm, set: setConfirm, autoComplete: "new-password" },
        ].map(({ id, label, value, set, autoComplete }) => (
          <div key={id} className="space-y-1.5">
            <label htmlFor={id} className="text-sm font-medium">{label}</label>
            <input
              id={id} type="password" autoComplete={autoComplete} required
              value={value} onChange={(e) => set(e.target.value)}
              className="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-background)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)] transition-colors"
            />
          </div>
        ))}

        {error && <p className="rounded-md bg-[var(--color-destructive)] bg-opacity-10 px-3 py-2 text-sm text-[var(--color-destructive)]">{error}</p>}
        {success && (
          <p className="flex items-center gap-1.5 rounded-md bg-green-500 bg-opacity-10 px-3 py-2 text-sm text-green-600">
            <CheckCircle size={14} /> Password updated successfully.
          </p>
        )}

        <button type="submit" disabled={loading} className="flex w-full items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] px-4 py-2 text-sm font-semibold text-[var(--color-primary-foreground)] hover:opacity-90 disabled:opacity-60 transition-opacity">
          {loading && <Loader2 size={14} className="animate-spin" />}
          Update password
        </button>
      </form>
    </div>
  );
}
