"use client";

import * as React from "react";
import { Box, LayoutDashboard, ArrowRightLeft, Vote, Wallet, Building2, Menu, X, ExternalLink } from "lucide-react";

function cn(...inputs: any[]) {
    return inputs.filter(Boolean).join(' ');
}

export const WorkingHeader = ({ appName }: { appName: string }) => {
    const [isMobileMenuOpen, setIsMobileMenuOpen] = React.useState(false);

    const NAV_ITEMS = [
        { name: "Explorer", href: "https://scan.sharehodl.com", icon: LayoutDashboard, description: "Browse blocks and transactions" },
        { name: "Trading", href: "https://trade.sharehodl.com", icon: ArrowRightLeft, description: "Trade ShareHODL tokens" },
        { name: "Governance", href: "https://gov.sharehodl.com", icon: Vote, description: "Participate in governance" },
        { name: "Wallet", href: "https://wallet.sharehodl.com", icon: Wallet, description: "Manage your assets" },
        { name: "Business", href: "https://business.sharehodl.com", icon: Building2, description: "Business solutions" },
    ];

    const toggleMobileMenu = () => setIsMobileMenuOpen(!isMobileMenuOpen);

    return (
        <>
            <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur-md supports-[backdrop-filter]:bg-background/80 shadow-sm">
                <div className="container mx-auto flex h-16 items-center px-4">
                    <div className="flex items-center gap-2 font-bold text-xl flex-1 md:flex-none md:mr-8">
                        <a href="https://sharehodl.com" className="flex items-center gap-2 hover:text-primary transition-colors group">
                            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary/20 to-primary/10 text-primary group-hover:scale-105 transition-transform">
                                <Box className="h-5 w-5" />
                            </div>
                            <span className="hidden sm:inline-block">ShareHODL</span>
                        </a>
                        <span className="text-muted-foreground/30 font-thin hidden sm:inline">/</span>
                        <span className="text-foreground font-medium hidden sm:inline">{appName}</span>
                    </div>

                    {/* Desktop Navigation */}
                    <nav className="hidden md:flex items-center gap-1 flex-1 justify-center">
                        <div className="flex items-center p-1 rounded-full bg-muted/30 border border-border/40 backdrop-blur-sm shadow-sm">
                            {NAV_ITEMS.map((item) => {
                                const isActive = item.name === appName;
                                return (
                                    <a
                                        key={item.name}
                                        href={item.href}
                                        title={item.description}
                                        className={cn(
                                            "flex items-center gap-2 px-4 py-2 text-sm font-medium transition-all duration-300 rounded-full group relative",
                                            isActive
                                                ? "bg-background text-primary shadow-sm ring-1 ring-black/5 dark:ring-white/10 scale-105"
                                                : "text-muted-foreground hover:text-foreground hover:bg-muted/50 hover:scale-105"
                                        )}
                                    >
                                        <item.icon className={cn(
                                            "h-4 w-4 transition-colors", 
                                            isActive ? "text-primary" : "text-muted-foreground group-hover:text-foreground"
                                        )} />
                                        {item.name}
                                        {isActive && (
                                            <div className="absolute -bottom-1 left-1/2 h-0.5 w-4 -translate-x-1/2 bg-primary rounded-full" />
                                        )}
                                    </a>
                                );
                            })}
                        </div>
                    </nav>

                    {/* Desktop Actions */}
                    <div className="hidden md:flex items-center gap-4 ml-auto">
                        <button className="h-9 rounded-md px-3 hover:bg-accent hover:text-accent-foreground text-muted-foreground hover:text-foreground group inline-flex items-center justify-center whitespace-nowrap text-sm font-medium ring-offset-background transition-colors">
                            Docs
                            <ExternalLink className="ml-1 h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                        </button>
                        <button className="h-9 rounded-full px-6 shadow-lg shadow-primary/20 hover:shadow-primary/30 transition-all hover:scale-105 bg-gradient-to-r from-primary to-primary/90 bg-primary text-primary-foreground hover:bg-primary/90 inline-flex items-center justify-center whitespace-nowrap text-sm font-medium ring-offset-background transition-colors">
                            Connect Wallet
                        </button>
                    </div>

                    {/* Mobile Menu Button */}
                    <button
                        className="md:hidden h-9 w-9 rounded-md hover:bg-accent hover:text-accent-foreground inline-flex items-center justify-center text-sm font-medium ring-offset-background transition-colors ml-2 flex-shrink-0"
                        onClick={toggleMobileMenu}
                        aria-label={isMobileMenuOpen ? "Close menu" : "Open menu"}
                    >
                        {isMobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
                    </button>
                </div>
            </header>

            {/* Mobile Menu Overlay */}
            {isMobileMenuOpen && (
                <div 
                    className="fixed inset-0 z-40 bg-black/20 backdrop-blur-sm md:hidden"
                    onClick={() => setIsMobileMenuOpen(false)}
                />
            )}

            {/* Mobile Menu */}
            <div className={cn(
                "fixed top-16 left-0 right-0 z-50 bg-background border-b shadow-lg transform transition-transform duration-300 ease-in-out md:hidden",
                isMobileMenuOpen ? "translate-y-0" : "-translate-y-full"
            )}>
                <div className="container mx-auto px-4 py-6">
                    <nav className="space-y-4">
                        <div className="grid gap-2">
                            {NAV_ITEMS.map((item) => {
                                const isActive = item.name === appName;
                                return (
                                    <a
                                        key={item.name}
                                        href={item.href}
                                        onClick={() => setIsMobileMenuOpen(false)}
                                        className={cn(
                                            "flex items-center gap-3 p-3 rounded-lg transition-all duration-200 border",
                                            isActive
                                                ? "bg-primary/10 text-primary border-primary/20 shadow-sm"
                                                : "text-muted-foreground hover:text-foreground hover:bg-muted/50 border-transparent hover:border-border"
                                        )}
                                    >
                                        <item.icon className={cn(
                                            "h-5 w-5", 
                                            isActive ? "text-primary" : "text-muted-foreground"
                                        )} />
                                        <div className="flex-1">
                                            <div className={cn("font-medium", isActive ? "text-primary" : "text-foreground")}>
                                                {item.name}
                                            </div>
                                            <div className="text-xs text-muted-foreground">
                                                {item.description}
                                            </div>
                                        </div>
                                        {isActive && (
                                            <div className="h-2 w-2 bg-primary rounded-full" />
                                        )}
                                    </a>
                                );
                            })}
                        </div>
                        
                        <div className="pt-4 border-t space-y-3">
                            <button className="w-full justify-start text-muted-foreground hover:text-foreground hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2 inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors">
                                <ExternalLink className="mr-2 h-4 w-4" />
                                Documentation
                            </button>
                            <button className="w-full rounded-lg bg-gradient-to-r from-primary to-primary/90 shadow-lg hover:shadow-primary/20 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2 inline-flex items-center justify-center whitespace-nowrap text-sm font-medium ring-offset-background transition-colors">
                                Connect Wallet
                            </button>
                        </div>
                    </nav>
                </div>
            </div>
        </>
    );
};