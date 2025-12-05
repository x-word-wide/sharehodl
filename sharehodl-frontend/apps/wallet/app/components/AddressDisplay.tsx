"use client";

import { useState } from "react";

// Simple inline implementation
function generateRandomHodlAddress(): string {
    const chars = '0123456789abcdef';
    let hexPart = '';
    
    for (let i = 0; i < 40; i++) {
        hexPart += chars[Math.floor(Math.random() * chars.length)];
    }
    
    return 'Hodl' + hexPart;
}

function formatHodlAddress(address: string): string {
    // Add spaces for better readability: Hodl 46d0 7236 46bc c9eb 6bf1 f382 871c 8b0f c321 54ad
    const prefix = address.substring(0, 4); // "Hodl"
    const hex = address.substring(4);
    const chunks = hex.match(/.{1,4}/g) || [];
    
    return `${prefix} ${chunks.join(' ')}`;
}

export function AddressDisplay() {
    const [address] = useState(() => generateRandomHodlAddress());
    const [copied, setCopied] = useState(false);

    const handleCopy = async () => {
        await navigator.clipboard.writeText(address);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="border rounded-lg p-6">
            <h3 className="font-semibold mb-4">Your Hodl Address</h3>
            <div className="space-y-3">
                <div className="p-3 bg-muted rounded border font-mono text-sm break-all">
                    {formatHodlAddress(address)}
                </div>
                <div className="flex gap-2">
                    <button 
                        onClick={handleCopy}
                        className="px-4 py-2 bg-blue-500 text-white rounded text-sm hover:bg-blue-600 transition-colors"
                    >
                        {copied ? 'Copied!' : 'Copy Address'}
                    </button>
                    <button className="px-4 py-2 bg-green-500 text-white rounded text-sm hover:bg-green-600 transition-colors">
                        Show QR Code
                    </button>
                </div>
                <p className="text-xs text-muted-foreground">
                    Use this address to receive ShareHODL tokens and assets. 
                    Compatible with Trust Wallet and other multi-chain wallets.
                </p>
            </div>
        </div>
    );
}