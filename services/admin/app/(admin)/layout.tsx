import type { Metadata } from "next";
import { AdminAuthProvider } from "@/components/layout/AdminAuthProvider";
import { AdminShell } from "@/components/layout/AdminShell";
import "../globals.css";

export const metadata: Metadata = {
  title: { template: "%s — Inori Admin", default: "Inori Admin" },
};

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <AdminAuthProvider>
      <AdminShell>{children}</AdminShell>
    </AdminAuthProvider>
  );
}
