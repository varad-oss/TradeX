import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Real-Time Equity Trading Exchange",
  description: "Low-latency equity exchange",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="antialiased">{children}</body>
    </html>
  );
}
