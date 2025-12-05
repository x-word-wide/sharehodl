"use client";

export const BasicHeader = ({ appName }: { appName: string }) => {
    return (
        <header className="sticky top-0 z-50 w-full border-b bg-white">
            <div className="container mx-auto flex h-16 items-center justify-between px-4">
                <div className="flex items-center gap-2 font-bold text-xl">
                    <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500/20 to-blue-500/10 text-blue-600">
                        <div className="h-5 w-5 border-2 border-current rounded"></div>
                    </div>
                    <span>ShareHODL</span>
                    <span>/</span>
                    <span>{appName}</span>
                </div>
                <div className="flex items-center gap-4">
                    <a href="http://localhost:3001" className="px-3 py-2">Explorer</a>
                    <a href="http://localhost:3002" className="px-3 py-2">Trading</a>
                    <a href="http://localhost:3003" className="px-3 py-2 bg-blue-100">Governance</a>
                    <a href="http://localhost:3004" className="px-3 py-2">Wallet</a>
                    <a href="http://localhost:3005" className="px-3 py-2">Business</a>
                    <button className="bg-blue-500 text-white px-4 py-2 rounded">
                        Connect Wallet
                    </button>
                </div>
            </div>
        </header>
    );
};