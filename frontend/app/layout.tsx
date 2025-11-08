import type { Metadata } from "next";
import { Fira_Code, Geist } from "next/font/google";
import Script from "next/script";
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
      <head>
        <Script id="dark-mode-script" strategy="beforeInteractive">
          {`
            (function() {
              try {
                const stored = localStorage.getItem('darkMode');
                const isDark = stored ? JSON.parse(stored) : true;
                const html = document.documentElement;
                if (isDark) {
                  html.classList.add('dark');
                  html.classList.remove('light');
                } else {
                  html.classList.add('light');
                  html.classList.remove('dark');
                }
              } catch (e) {
                console.error('Dark mode script error:', e);
              }
            })();
          `}
        </Script>
      </head>
      <body
        className={`${geistSans.variable} ${firaCode.variable} antialiased bg-linear-to-br from-gray-50 to-gray-100 dark:from-gray-950 dark:to-black min-h-screen font-fira-code`}
        suppressHydrationWarning
      >
        {children}
      </body>
    </html>
  );
}
