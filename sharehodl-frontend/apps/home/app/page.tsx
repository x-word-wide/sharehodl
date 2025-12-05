"use client";

import { useState } from "react";

export default function Home() {
  const [activeSection, setActiveSection] = useState("platform");

  const platformStats = {
    listedCompanies: 4,
    totalMarketCap: "8.3T",
    dailyVolume: "45.2M",
    activeValidators: 127,
    settlementTime: "6 seconds",
    uptime: "99.97%"
  };

  const features = [
    {
      title: "Professional Trading",
      description: "FOK/IOC orders, circuit breakers, and institutional-grade trading features",
      link: "https://trade.sharehodl.com"
    },
    {
      title: "Blockchain Explorer",
      description: "Real-time network monitoring, block explorer, and transaction tracking",
      link: "https://scan.sharehodl.com"
    },
    {
      title: "Corporate Governance",
      description: "On-chain voting, proposal management, and shareholder democracy",
      link: "https://gov.sharehodl.com"
    },
    {
      title: "Digital Wallet",
      description: "Secure asset management, portfolio tracking, and transaction history",
      link: "https://wallet.sharehodl.com"
    },
    {
      title: "Business Portal",
      description: "IPO applications, validator registration, and enterprise services",
      link: "https://business.sharehodl.com"
    }
  ];

  const advantages = [
    {
      traditional: "T+2 Settlement (48-72 hours)",
      sharehodl: "6 Second Settlement",
      improvement: "28,800x faster"
    },
    {
      traditional: "8hr/day, 5days/week Trading",
      sharehodl: "24/7/365 Global Trading",
      improvement: "7.5x more access"
    },
    {
      traditional: "$5-15+ Trading Fees",
      sharehodl: "$0.005 Trading Fees",
      improvement: "1000-3000x cheaper"
    },
    {
      traditional: "Full Share Price Minimum",
      sharehodl: "$1 Minimum (Fractional)",
      improvement: "100x more accessible"
    },
    {
      traditional: "Beneficial Ownership (Broker)",
      sharehodl: "Direct Custody",
      improvement: "True ownership"
    }
  ];

  return (
    <div className="min-h-screen bg-background">
      <main className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="text-center mb-8 md:mb-12">
          <h1 className="text-3xl sm:text-4xl md:text-5xl font-bold mb-4 bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            ShareHODL
          </h1>
          <p className="text-lg sm:text-xl text-muted-foreground mb-4 max-w-3xl mx-auto px-4">
            The Future of Equity Markets
          </p>
          <p className="text-sm sm:text-base md:text-lg text-muted-foreground max-w-4xl mx-auto px-4">
            Professional blockchain infrastructure for institutional-grade equity trading, 
            corporate governance, and business services. Combining the security of traditional 
            markets with the efficiency of modern blockchain technology.
          </p>
        </div>

        {/* Navigation Menu */}
        <div className="grid gap-3 sm:gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 mb-8 md:mb-12">
          {features.map((feature, index) => (
            <a
              key={index}
              href={feature.link}
              className="group border rounded-lg p-4 sm:p-6 text-center hover:shadow-lg transition-all duration-300 hover:border-blue-500"
            >
              <h3 className="font-semibold mb-2 text-base sm:text-lg group-hover:text-blue-600">
                {feature.title}
              </h3>
              <p className="text-xs sm:text-sm text-muted-foreground">
                {feature.description}
              </p>
            </a>
          ))}
        </div>

        {/* Platform Stats */}
        <div className="grid gap-3 sm:gap-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 mb-8 md:mb-12">
          <div className="border rounded-lg p-3 sm:p-4 text-center">
            <div className="text-xl sm:text-2xl font-bold text-blue-600">{platformStats.listedCompanies}</div>
            <p className="text-xs sm:text-sm text-muted-foreground">Listed Companies</p>
          </div>
          <div className="border rounded-lg p-3 sm:p-4 text-center">
            <div className="text-xl sm:text-2xl font-bold text-green-600">${platformStats.totalMarketCap}</div>
            <p className="text-xs sm:text-sm text-muted-foreground">Total Market Cap</p>
          </div>
          <div className="border rounded-lg p-3 sm:p-4 text-center">
            <div className="text-xl sm:text-2xl font-bold text-purple-600">${platformStats.dailyVolume}</div>
            <p className="text-xs sm:text-sm text-muted-foreground">Daily Volume</p>
          </div>
          <div className="border rounded-lg p-3 sm:p-4 text-center">
            <div className="text-xl sm:text-2xl font-bold text-orange-600">{platformStats.activeValidators}</div>
            <p className="text-xs sm:text-sm text-muted-foreground">Active Validators</p>
          </div>
          <div className="border rounded-lg p-3 sm:p-4 text-center">
            <div className="text-xl sm:text-2xl font-bold text-red-600">{platformStats.settlementTime}</div>
            <p className="text-xs sm:text-sm text-muted-foreground">Settlement Time</p>
          </div>
          <div className="border rounded-lg p-3 sm:p-4 text-center">
            <div className="text-xl sm:text-2xl font-bold text-indigo-600">{platformStats.uptime}</div>
            <p className="text-xs sm:text-sm text-muted-foreground">Network Uptime</p>
          </div>
        </div>

        {/* Platform Advantages */}
        <div className="mb-8 md:mb-12">
          <h2 className="text-2xl sm:text-3xl font-bold text-center mb-6 sm:mb-8 px-4">
            ShareHODL vs Traditional Markets
          </h2>
          <div className="overflow-x-auto -mx-4 px-4">
            <table className="w-full border rounded-lg min-w-[600px]">
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left py-2 sm:py-3 px-2 sm:px-4 font-semibold text-sm sm:text-base">Traditional Markets</th>
                  <th className="text-left py-2 sm:py-3 px-2 sm:px-4 font-semibold text-blue-600 text-sm sm:text-base">ShareHODL</th>
                  <th className="text-left py-2 sm:py-3 px-2 sm:px-4 font-semibold text-green-600 text-sm sm:text-base">Improvement</th>
                </tr>
              </thead>
              <tbody>
                {advantages.map((item, index) => (
                  <tr key={index} className="border-t">
                    <td className="py-2 sm:py-3 px-2 sm:px-4 text-red-600 text-xs sm:text-sm">{item.traditional}</td>
                    <td className="py-2 sm:py-3 px-2 sm:px-4 text-blue-600 font-semibold text-xs sm:text-sm">{item.sharehodl}</td>
                    <td className="py-2 sm:py-3 px-2 sm:px-4 text-green-600 font-bold text-xs sm:text-sm">{item.improvement}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* About Section */}
        <div className="grid gap-6 sm:gap-8 md:grid-cols-2 mb-8 md:mb-12">
          <div className="border rounded-lg p-4 sm:p-6">
            <h3 className="font-bold text-lg sm:text-xl mb-3 sm:mb-4">
              For Companies
            </h3>
            <ul className="space-y-2 text-xs sm:text-sm">
              <li className="flex items-start sm:items-center gap-2">
                <span className="w-2 h-2 bg-green-500 rounded-full mt-1.5 sm:mt-0 flex-shrink-0"></span>
                <span>Streamlined IPO process: 2-6 weeks vs 12-18 months</span>
              </li>
              <li className="flex items-start sm:items-center gap-2">
                <span className="w-2 h-2 bg-green-500 rounded-full mt-1.5 sm:mt-0 flex-shrink-0"></span>
                <span>Lower costs: $1K-25K vs $10M+ traditional</span>
              </li>
              <li className="flex items-start sm:items-center gap-2">
                <span className="w-2 h-2 bg-green-500 rounded-full mt-1.5 sm:mt-0 flex-shrink-0"></span>
                <span>Global investor access 24/7</span>
              </li>
              <li className="flex items-start sm:items-center gap-2">
                <span className="w-2 h-2 bg-green-500 rounded-full mt-1.5 sm:mt-0 flex-shrink-0"></span>
                <span>Automated compliance and governance</span>
              </li>
              <li className="flex items-start sm:items-center gap-2">
                <span className="w-2 h-2 bg-green-500 rounded-full mt-1.5 sm:mt-0 flex-shrink-0"></span>
                <span>Fractional shares for broader accessibility</span>
              </li>
            </ul>
            <div className="mt-4">
              <a 
                href="https://business.sharehodl.com"
                className="inline-block bg-blue-500 text-white px-3 sm:px-4 py-2 text-sm rounded hover:bg-blue-600 transition-colors w-full sm:w-auto text-center"
              >
                Learn More
              </a>
            </div>
          </div>

          <div className="border rounded-lg p-4 sm:p-6">
            <h3 className="font-bold text-lg sm:text-xl mb-3 sm:mb-4">
              For Investors
            </h3>
            <ul className="space-y-2 text-xs sm:text-sm">
              <li className="flex items-start sm:items-center gap-2">
                <span className="w-2 h-2 bg-blue-500 rounded-full mt-1.5 sm:mt-0 flex-shrink-0"></span>
                <span>Professional trading tools (FOK, IOC orders)</span>
              </li>
              <li className="flex items-start sm:items-center gap-2">
                <span className="w-2 h-2 bg-blue-500 rounded-full mt-1.5 sm:mt-0 flex-shrink-0"></span>
                <span>Circuit breakers and trading safeguards</span>
              </li>
              <li className="flex items-start sm:items-center gap-2">
                <span className="w-2 h-2 bg-blue-500 rounded-full mt-1.5 sm:mt-0 flex-shrink-0"></span>
                <span>Direct ownership and custody</span>
              </li>
              <li className="flex items-start sm:items-center gap-2">
                <span className="w-2 h-2 bg-blue-500 rounded-full mt-1.5 sm:mt-0 flex-shrink-0"></span>
                <span>On-chain governance and voting</span>
              </li>
              <li className="flex items-start sm:items-center gap-2">
                <span className="w-2 h-2 bg-blue-500 rounded-full mt-1.5 sm:mt-0 flex-shrink-0"></span>
                <span>Instant settlement and 24/7 trading</span>
              </li>
            </ul>
            <div className="mt-4">
              <a 
                href="https://trade.sharehodl.com"
                className="inline-block bg-green-500 text-white px-3 sm:px-4 py-2 text-sm rounded hover:bg-green-600 transition-colors w-full sm:w-auto text-center"
              >
                Start Trading
              </a>
            </div>
          </div>
        </div>

        {/* Technology */}
        <div className="border rounded-lg p-4 sm:p-6 mb-8 md:mb-12">
          <h3 className="font-bold text-lg sm:text-xl mb-4 sm:mb-6 text-center px-4">
            Institutional-Grade Technology
          </h3>
          <div className="grid gap-4 sm:gap-6 sm:grid-cols-2 lg:grid-cols-4 text-center">
            <div className="p-3 sm:p-0">
              <h4 className="font-semibold mb-2 text-sm sm:text-base">Blockchain Security</h4>
              <p className="text-xs sm:text-sm text-muted-foreground">
                Built on Cosmos SDK with institutional-grade validators and Byzantine fault tolerance
              </p>
            </div>
            <div className="p-3 sm:p-0">
              <h4 className="font-semibold mb-2 text-sm sm:text-base">Professional Trading</h4>
              <p className="text-xs sm:text-sm text-muted-foreground">
                FOK/IOC orders, circuit breakers, and professional order types for institutional traders
              </p>
            </div>
            <div className="p-3 sm:p-0">
              <h4 className="font-semibold mb-2 text-sm sm:text-base">Governance & Compliance</h4>
              <p className="text-xs sm:text-sm text-muted-foreground">
                On-chain governance, automated compliance, and regulatory-ready infrastructure
              </p>
            </div>
            <div className="p-3 sm:p-0">
              <h4 className="font-semibold mb-2 text-sm sm:text-base">Enterprise Ready</h4>
              <p className="text-xs sm:text-sm text-muted-foreground">
                99.97% uptime, 6-second settlement, and enterprise-grade API integrations
              </p>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="text-center text-muted-foreground">
          <p className="mb-2">
            ShareHODL Professional Blockchain Platform
          </p>
          <p className="text-sm mb-4">
            Enterprise-grade infrastructure with institutional security and professional trading features.
          </p>
          <div className="flex justify-center items-center gap-6 pt-4 border-t">
            <a 
              href="https://x.com/share_hodl" 
              target="_blank" 
              rel="noopener noreferrer"
              className="text-muted-foreground hover:text-foreground transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
              </svg>
              <span className="text-sm">@share_hodl</span>
            </a>
            <a 
              href="https://github.com/x-word-wide/sharehodl" 
              target="_blank" 
              rel="noopener noreferrer"
              className="text-muted-foreground hover:text-foreground transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd"/>
              </svg>
              <span className="text-sm">GitHub</span>
            </a>
          </div>
        </div>
      </main>
    </div>
  );
}