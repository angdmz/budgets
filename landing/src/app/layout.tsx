import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Budget Management System - Take Control of Your Finances",
  description: "A powerful budget management system with multi-user support, group-based budgeting, and real-time expense tracking. Start managing your finances better today.",
  keywords: ["budget", "finance", "expense tracking", "money management", "budgeting app"],
  authors: [{ name: "Budget Management System" }],
  openGraph: {
    title: "Budget Management System - Take Control of Your Finances",
    description: "A powerful budget management system with multi-user support, group-based budgeting, and real-time expense tracking.",
    type: "website",
    locale: "en_US",
  },
  twitter: {
    card: "summary_large_image",
    title: "Budget Management System",
    description: "Take control of your finances with our powerful budget management system.",
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={inter.className}>{children}</body>
    </html>
  );
}
