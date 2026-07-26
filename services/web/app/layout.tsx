import type { Metadata, Viewport } from "next";
import { I18nProvider } from "@/components/layout/I18nProvider";
import { AuthProvider } from "@/components/layout/AuthProvider";
import "./globals.css";

export const metadata: Metadata = {
  title: {
    template: "%s — Inori Music",
    default: "Inori Music",
  },
  description: "Your personal music library",
  manifest: "/manifest.json",
  appleWebApp: {
    capable: true,
    statusBarStyle: "default",
    title: "Inori Music",
  },
};

export const viewport: Viewport = {
  themeColor: "#fff7f2",
  width: "device-width",
  initialScale: 1,
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <I18nProvider>
          <AuthProvider>{children}</AuthProvider>
        </I18nProvider>
      </body>
    </html>
  );
}
