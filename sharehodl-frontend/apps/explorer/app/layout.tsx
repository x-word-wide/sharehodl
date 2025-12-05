import "@repo/ui/styles.css";
import "./globals.css";
import type { Metadata } from "next";
import { Geist } from "next/font/google";
import { WorkingHeader } from "./components/WorkingHeader";

const geist = Geist({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "ShareHODL Explorer",
  description: "Explore the ShareHODL blockchain",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={geist.className}>
        <WorkingHeader appName="Explorer" />
        {children}
      </body>
    </html>
  );
}
