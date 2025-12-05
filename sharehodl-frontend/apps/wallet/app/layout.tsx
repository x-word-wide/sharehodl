import "@repo/ui/styles.css";
import "./globals.css";
import type { Metadata } from "next";
import { Geist } from "next/font/google";
import { WorkingHeader } from "../../explorer/app/components/WorkingHeader";

const geist = Geist({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "ShareHODL Wallet",
  description: "Manage your ShareHODL assets securely",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={geist.className}>
        <WorkingHeader appName="Wallet" />
        {children}
      </body>
    </html>
  );
}
