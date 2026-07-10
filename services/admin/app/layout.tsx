import type { Metadata, Viewport } from "next";
import { AdminI18nProvider } from "@/components/layout/AdminI18nProvider";
import "./globals.css";

export const metadata: Metadata = {
  title: { template: "%s — Inori Admin", default: "Inori Admin" },
  description: "Inori Music administration console",
};

export const viewport: Viewport = {
  themeColor: "#070711",
  width: "device-width",
  initialScale: 1,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <link
          href="https://fonts.googleapis.com/css2?family=Orbitron:wght@700;900&family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400&family=Noto+Sans+JP:wght@400;700&display=swap"
          rel="stylesheet"
        />
      </head>
      <body>
        <AdminI18nProvider>{children}</AdminI18nProvider>
      </body>
    </html>
  );
}
