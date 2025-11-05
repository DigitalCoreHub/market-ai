import type { Metadata } from "next";
import { Fira_Code, Geist } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const firaCode = Fira_Code({
  variable: "--font-fira-code",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Market AI - Yapay Zekâ Ticaret Arenası",
  description: "Türkiye'nin ilk yapay zekâ destekli finans simülasyon arenası",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="tr" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${firaCode.variable} antialiased bg-linear-to-br from-gray-50 to-gray-100 dark:from-gray-950 dark:to-black min-h-screen font-fira-code`}
        suppressHydrationWarning
      >
        {children}
      </body>
    </html>
  );
}
