"use client";

import { Box } from "lucide-react";

export const SimpleHeader = ({ appName }: { appName: string }) => {
    return (
        <header className="sticky top-0 z-50 w-full border-b bg-background/95">
            <div className="container mx-auto flex h-16 items-center justify-between px-4">
                <div className="flex items-center gap-2 font-bold text-xl">
                    <Box className="h-5 w-5" />
                    <span>ShareHODL</span>
                    <span>/</span>
                    <span>{appName}</span>
                </div>
                <button className="bg-primary text-primary-foreground px-4 py-2 rounded">
                    Connect Wallet
                </button>
            </div>
        </header>
    );
};