import "@repo/ui/styles.css";
import "./globals.css";
import type { Metadata } from "next";
import { Geist } from "next/font/google";
import { WorkingHeader } from "../../explorer/app/components/WorkingHeader";

const geist = Geist({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "ShareHODL Governance",
  description: "Participate in ShareHODL governance decisions",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={geist.className}>
        <WorkingHeader appName="Governance" />
        {children}
      </body>
    </html>
  );
}
