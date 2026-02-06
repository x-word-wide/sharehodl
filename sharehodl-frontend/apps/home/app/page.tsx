"use client";

import { useState, useEffect } from "react";
import { WalletButton, useBlockchain } from "@repo/ui";

export default function Home() {
  const { networkStatus } = useBlockchain();
  const [activeSection, setActiveSection] = useState("platform");

  // Use real network data when available
  const platformStats = {
    listedCompanies: 4,
    totalMarketCap: "8.3T",
    dailyVolume: "45.2M",
    activeValidators: networkStatus?.validatorCount || 1,
    settlementTime: "2 seconds",
    blockHeight: networkStatus?.latestBlockHeight
      ? parseInt(networkStatus.latestBlockHeight).toLocaleString()
      : "---"
  };

  const features = [
    {
      title: "Professional Trading",
      description: "FOK/IOC orders, circuit breakers, and institutional-grade trading features",
      link: "http://localhost:3002"
    },
    {
      title: "Blockchain Explorer",
      description: "Real-time network monitoring, block explorer, and transaction tracking",
      link: "http://localhost:3003"
    },
    {
      title: "Corporate Governance",
      description: "On-chain voting, proposal management, and shareholder democracy",
      link: "http://localhost:3001"
    },
    {
      title: "Digital Wallet",
      description: "Secure asset management, portfolio tracking, and transaction history",
      link: "http://localhost:3004"
    },
    {
      title: "Business Portal",
      description: "IPO applications, validator registration, and enterprise services",
      link: "http://localhost:3005"
    },
    {
      title: "DeFi Lending",
      description: "Collateralized loans, lending pools, and P2P marketplace",
      link: "http://localhost:3006"
    },
    {
      title: "Asset Inheritance",
      description: "Dead man switch, beneficiary management, and secure legacy transfer",
      link: "http://localhost:3007"
    }
  ];

  const advantages = [
    {
      traditional: "T+2 Settlement (48-72 hours)",
      sharehodl: "2 Second Settlement",
      improvement: "43,200x faster"
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
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Header */}
      <header className="border-b border-gray-800 sticky top-0 z-50 bg-gray-950/95 backdrop-blur">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-2">
            <span className="text-2xl font-bold bg-gradient-to-r from-blue-500 to-purple-500 bg-clip-text text-transparent">
              ShareHODL
            </span>
            {networkStatus?.connected && (
              <span className="text-xs px-2 py-1 bg-green-900/30 text-green-400 rounded-full">
                Live
              </span>
            )}
          </div>
          <div className="flex items-center gap-4">
            <nav className="hidden md:flex items-center gap-6 text-sm">
              <a href="http://localhost:3002" className="text-gray-400 hover:text-white transition">Trade</a>
              <a href="http://localhost:3003" className="text-gray-400 hover:text-white transition">Explorer</a>
              <a href="http://localhost:3001" className="text-gray-400 hover:text-white transition">Governance</a>
              <a href="http://localhost:3006" className="text-gray-400 hover:text-white transition">Lending</a>
              <a href="http://localhost:3007" className="text-gray-400 hover:text-white transition">Inheritance</a>
            </nav>
            <WalletButton />
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Hero Section */}
        <div className="text-center mb-12 py-12">
          <h1 className="text-4xl sm:text-5xl md:text-6xl font-bold mb-6 bg-gradient-to-r from-blue-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
            The Future of Equity Markets
          </h1>
          <p className="text-lg sm:text-xl text-gray-400 mb-8 max-w-3xl mx-auto">
            Professional blockchain infrastructure for institutional-grade equity trading,
            corporate governance, and business services.
          </p>
          <div className="flex justify-center gap-4 flex-wrap">
            <a
              href="http://localhost:3002"
              className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
            >
              Start Trading
            </a>
            <a
              href="http://localhost:3003"
              className="px-6 py-3 bg-gray-800 hover:bg-gray-700 text-white rounded-lg font-medium transition-colors"
            >
              Explore Network
            </a>
          </div>
        </div>

        {/* Network Status */}
        <div className={`mb-12 p-6 rounded-xl border ${
          networkStatus?.connected
            ? 'bg-green-900/10 border-green-800'
            : 'bg-gray-900 border-gray-800'
        }`}>
          <div className="flex items-center justify-between flex-wrap gap-4">
            <div className="flex items-center gap-3">
              <div className={`w-3 h-3 rounded-full ${
                networkStatus?.connected ? 'bg-green-400 animate-pulse' : 'bg-gray-600'
              }`} />
              <div>
                <p className="font-medium">
                  {networkStatus?.connected ? 'Network Online' : 'Connecting...'}
                </p>
                <p className="text-sm text-gray-400">
                  {networkStatus?.chainId || 'sharehodl-1'} | Block #{platformStats.blockHeight}
                </p>
              </div>
            </div>
            <div className="flex gap-6 text-sm">
              <div className="text-center">
                <p className="text-2xl font-bold text-blue-400">{platformStats.settlementTime}</p>
                <p className="text-gray-500">Settlement</p>
              </div>
              <div className="text-center">
                <p className="text-2xl font-bold text-green-400">{platformStats.activeValidators}</p>
                <p className="text-gray-500">Validators</p>
              </div>
            </div>
          </div>
        </div>

        {/* Navigation Cards */}
        <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 mb-12">
          {features.map((feature, index) => (
            <a
              key={index}
              href={feature.link}
              className="group border border-gray-800 rounded-xl p-6 text-center hover:border-blue-500/50 hover:bg-gray-900/50 transition-all duration-300"
            >
              <h3 className="font-semibold mb-2 text-lg group-hover:text-blue-400 transition-colors">
                {feature.title}
              </h3>
              <p className="text-sm text-gray-500">
                {feature.description}
              </p>
            </a>
          ))}
        </div>

        {/* Platform Stats */}
        <div className="grid gap-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 mb-12">
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-blue-400">{platformStats.listedCompanies}</div>
            <p className="text-sm text-gray-500">Listed Companies</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-green-400">${platformStats.totalMarketCap}</div>
            <p className="text-sm text-gray-500">Total Market Cap</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-purple-400">${platformStats.dailyVolume}</div>
            <p className="text-sm text-gray-500">Daily Volume</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-orange-400">{platformStats.activeValidators}</div>
            <p className="text-sm text-gray-500">Active Validators</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-red-400">{platformStats.settlementTime}</div>
            <p className="text-sm text-gray-500">Settlement Time</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-indigo-400">{platformStats.blockHeight}</div>
            <p className="text-sm text-gray-500">Block Height</p>
          </div>
        </div>

        {/* Advantages Comparison */}
        <div className="mb-12">
          <h2 className="text-2xl sm:text-3xl font-bold text-center mb-8">
            ShareHODL vs Traditional Markets
          </h2>
          <div className="overflow-x-auto">
            <table className="w-full border border-gray-800 rounded-xl overflow-hidden min-w-[600px]">
              <thead className="bg-gray-900">
                <tr>
                  <th className="text-left py-4 px-4 font-semibold text-gray-400">Traditional Markets</th>
                  <th className="text-left py-4 px-4 font-semibold text-blue-400">ShareHODL</th>
                  <th className="text-left py-4 px-4 font-semibold text-green-400">Improvement</th>
                </tr>
              </thead>
              <tbody>
                {advantages.map((item, index) => (
                  <tr key={index} className="border-t border-gray-800">
                    <td className="py-3 px-4 text-red-400 text-sm">{item.traditional}</td>
                    <td className="py-3 px-4 text-blue-400 font-medium text-sm">{item.sharehodl}</td>
                    <td className="py-3 px-4 text-green-400 font-bold text-sm">{item.improvement}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* For Companies / For Investors */}
        <div className="grid gap-6 md:grid-cols-2 mb-12">
          <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
            <h3 className="font-bold text-xl mb-4 text-blue-400">For Companies</h3>
            <ul className="space-y-3 text-sm">
              <li className="flex items-center gap-3">
                <span className="w-2 h-2 bg-green-500 rounded-full flex-shrink-0"></span>
                <span className="text-gray-300">Streamlined IPO process: 2-6 weeks vs 12-18 months</span>
              </li>
              <li className="flex items-center gap-3">
                <span className="w-2 h-2 bg-green-500 rounded-full flex-shrink-0"></span>
                <span className="text-gray-300">Lower costs: $1K-25K vs $10M+ traditional</span>
              </li>
              <li className="flex items-center gap-3">
                <span className="w-2 h-2 bg-green-500 rounded-full flex-shrink-0"></span>
                <span className="text-gray-300">Global investor access 24/7</span>
              </li>
              <li className="flex items-center gap-3">
                <span className="w-2 h-2 bg-green-500 rounded-full flex-shrink-0"></span>
                <span className="text-gray-300">Automated compliance and governance</span>
              </li>
            </ul>
            <div className="mt-6">
              <a
                href="http://localhost:3005"
                className="inline-block bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 text-sm rounded-lg transition-colors"
              >
                Get Started
              </a>
            </div>
          </div>

          <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
            <h3 className="font-bold text-xl mb-4 text-green-400">For Investors</h3>
            <ul className="space-y-3 text-sm">
              <li className="flex items-center gap-3">
                <span className="w-2 h-2 bg-blue-500 rounded-full flex-shrink-0"></span>
                <span className="text-gray-300">Professional trading tools (FOK, IOC orders)</span>
              </li>
              <li className="flex items-center gap-3">
                <span className="w-2 h-2 bg-blue-500 rounded-full flex-shrink-0"></span>
                <span className="text-gray-300">Circuit breakers and trading safeguards</span>
              </li>
              <li className="flex items-center gap-3">
                <span className="w-2 h-2 bg-blue-500 rounded-full flex-shrink-0"></span>
                <span className="text-gray-300">Direct ownership and custody</span>
              </li>
              <li className="flex items-center gap-3">
                <span className="w-2 h-2 bg-blue-500 rounded-full flex-shrink-0"></span>
                <span className="text-gray-300">On-chain governance and voting</span>
              </li>
            </ul>
            <div className="mt-6">
              <a
                href="http://localhost:3002"
                className="inline-block bg-green-600 hover:bg-green-700 text-white px-4 py-2 text-sm rounded-lg transition-colors"
              >
                Start Trading
              </a>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="text-center text-gray-500 pt-8 border-t border-gray-800">
          <p className="mb-2 text-gray-400">
            ShareHODL Professional Blockchain Platform
          </p>
          <p className="text-sm mb-4">
            Enterprise-grade infrastructure with institutional security and professional trading features.
          </p>
          <div className="flex justify-center items-center gap-6">
            <a
              href="https://x.com/share_hodl"
              target="_blank"
              rel="noopener noreferrer"
              className="text-gray-500 hover:text-white transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
              </svg>
              <span className="text-sm">@share_hodl</span>
            </a>
            <a
              href="https://github.com/x-word-wide/sharehodl"
              target="_blank"
              rel="noopener noreferrer"
              className="text-gray-500 hover:text-white transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
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
