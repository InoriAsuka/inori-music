import type { Metadata, Viewport } from "next";
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
      <body>{children}</body>
    </html>
  );
}
