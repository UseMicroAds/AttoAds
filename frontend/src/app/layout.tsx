import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Link from "next/link";
import "./globals.css";
import { Providers } from "@/components/providers";
import { Navbar } from "@/components/navbar";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "AttoAds — Viral Comment Monetization",
  description:
    "A micro-sponsorship marketplace connecting brands with top YouTube commenters.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} min-h-screen antialiased`}
      >
        <Providers>
          <div className="flex min-h-screen flex-col">
            <Navbar />
            <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-8">
              {children}
            </main>
            <footer className="border-t">
              <div className="mx-auto flex max-w-7xl flex-col gap-4 px-4 py-6 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
                <p>AttoAds connects brands with top YouTube commenters.</p>
                <nav className="flex gap-4">
                  <Link
                    href="/privacy"
                    className="transition-colors hover:text-foreground"
                  >
                    Privacy Policy
                  </Link>
                  <Link
                    href="/terms"
                    className="transition-colors hover:text-foreground"
                  >
                    Terms of Service
                  </Link>
                </nav>
              </div>
            </footer>
          </div>
        </Providers>
      </body>
    </html>
  );
}
