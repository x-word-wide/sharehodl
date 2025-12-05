"use client";

import * as React from "react";

export function TestHeader({ appName }: { appName: string }) {
    return (
        <header className="bg-blue-500 text-white p-4">
            <h1>Test Header - {appName}</h1>
        </header>
    );
}