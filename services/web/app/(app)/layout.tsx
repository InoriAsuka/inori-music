import type { Metadata } from "next";
import "../../globals.css";
import { AuthProvider } from "@/components/layout/AuthProvider";
import { AppShell } from "@/components/layout/AppShell";

export const metadata: Metadata = {
  title: { template: "%s — Inori Music", default: "Inori Music" },
};

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthProvider>
      <AppShell>{children}</AppShell>
    </AuthProvider>
  );
}
