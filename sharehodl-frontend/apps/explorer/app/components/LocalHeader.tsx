"use client";

export function LocalHeader({ appName }: { appName: string }) {
    return (
        <header className="bg-gray-800 text-white p-4">
            <h1>Local Header - {appName}</h1>
        </header>
    );
}