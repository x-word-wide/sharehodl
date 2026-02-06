"use client";

import { useState } from "react";
import { WalletButton, useWallet, useBlockchain } from "@repo/ui";
import {
  Landmark,
  TrendingUp,
  TrendingDown,
  Percent,
  AlertTriangle,
  Clock,
  Shield,
  Plus,
  ArrowRight,
  RefreshCw,
  Wallet,
  Users,
  PiggyBank,
  Zap,
} from "lucide-react";

export default function Home() {
  const { connected, address, balances } = useWallet();
  const { networkStatus } = useBlockchain();
  const [activeTab, setActiveTab] = useState<"loans" | "pools" | "p2p" | "liquidations">("loans");

  // Mock data for demonstration
  const myLoans = [
    {
      id: "LOAN-001",
      type: "borrowed",
      principal: "10,000",
      collateral: "15,000 HODL",
      collateralRatio: 150,
      interestRate: 8.5,
      duration: "90 days",
      remaining: "45 days",
      status: "active",
      accrued: "350.00",
    },
    {
      id: "LOAN-002",
      type: "lent",
      principal: "5,000",
      borrower: "hodl1abc...xyz",
      collateralRatio: 165,
      interestRate: 9.2,
      duration: "60 days",
      remaining: "12 days",
      status: "active",
      accrued: "125.50",
    },
  ];

  const lendingPools = [
    {
      name: "HODL Stable Pool",
      totalLiquidity: "2,500,000",
      utilization: 72,
      supplyAPY: 5.8,
      borrowAPY: 8.2,
      myDeposit: connected ? "1,250" : null,
    },
    {
      name: "High Yield Pool",
      totalLiquidity: "850,000",
      utilization: 85,
      supplyAPY: 8.5,
      borrowAPY: 12.4,
      myDeposit: null,
    },
  ];

  const p2pOffers = [
    {
      id: "OFFER-001",
      type: "lend",
      amount: "25,000",
      interestRate: 7.5,
      duration: "30 days",
      collateralRequired: 150,
      lender: "hodl1def...uvw",
      trustCeiling: "50,000",
    },
    {
      id: "REQ-001",
      type: "borrow",
      amount: "8,000",
      interestRate: 9.0,
      duration: "60 days",
      collateralOffered: "13,600 HODL",
      borrower: "hodl1ghi...rst",
      trustCeiling: "20,000",
    },
  ];

  const liquidatableLoans = [
    {
      id: "LOAN-099",
      borrower: "hodl1xyz...abc",
      principal: "15,000",
      collateral: "17,250 HODL",
      collateralRatio: 115,
      liquidationBonus: "10%",
    },
  ];

  const formatBalance = () => {
    const hodl = balances.find((b) => b.denom === "uhodl");
    if (hodl) {
      return (parseInt(hodl.amount) / 1000000).toLocaleString(undefined, {
        minimumFractionDigits: 2,
      });
    }
    return "0.00";
  };

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Header */}
      <header className="border-b border-gray-800 sticky top-0 z-50 bg-gray-950/95 backdrop-blur">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-2">
            <Landmark className="h-6 w-6 text-amber-400" />
            <span className="text-2xl font-bold bg-gradient-to-r from-amber-500 to-orange-500 bg-clip-text text-transparent">
              Lending
            </span>
            {networkStatus?.connected && (
              <span className="text-xs px-2 py-1 bg-green-900/30 text-green-400 rounded-full">
                Live
              </span>
            )}
          </div>
          <div className="flex items-center gap-4">
            <nav className="hidden md:flex items-center gap-6 text-sm">
              <a href="http://localhost:3000" className="text-gray-400 hover:text-white transition">Home</a>
              <a href="http://localhost:3002" className="text-gray-400 hover:text-white transition">Trade</a>
              <a href="http://localhost:3004" className="text-gray-400 hover:text-white transition">Wallet</a>
            </nav>
            <WalletButton />
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold mb-4 bg-gradient-to-r from-amber-400 via-orange-400 to-red-400 bg-clip-text text-transparent">
            ShareHODL Lending
          </h1>
          <p className="text-gray-400 text-lg max-w-3xl mx-auto">
            Collateralized loans with equity and HODL tokens. Earn yield by lending or borrow against your assets.
          </p>
        </div>

        {/* Stats Overview */}
        <div className="grid gap-4 grid-cols-2 md:grid-cols-4 mb-8">
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-amber-400">$3.5M</div>
            <p className="text-sm text-gray-500">Total Value Locked</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-green-400">5.8%</div>
            <p className="text-sm text-gray-500">Avg Supply APY</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-blue-400">8.2%</div>
            <p className="text-sm text-gray-500">Avg Borrow APY</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-purple-400">142</div>
            <p className="text-sm text-gray-500">Active Loans</p>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="border-b border-gray-800 mb-6">
          <nav className="-mb-px flex flex-wrap gap-2">
            {[
              { id: "loans" as const, label: "My Loans", icon: Wallet },
              { id: "pools" as const, label: "Lending Pools", icon: PiggyBank },
              { id: "p2p" as const, label: "P2P Marketplace", icon: Users },
              { id: "liquidations" as const, label: "Liquidations", icon: AlertTriangle },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-2 px-4 flex items-center gap-2 rounded-t-lg transition-colors ${
                  activeTab === tab.id
                    ? "bg-gray-800 text-white border-b-2 border-amber-500"
                    : "text-gray-500 hover:text-gray-300"
                }`}
              >
                <tab.icon className="h-4 w-4" />
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* My Loans Tab */}
        {activeTab === "loans" && (
          <div>
            <div className="flex justify-between items-center mb-6">
              <h3 className="text-lg font-semibold flex items-center gap-2">
                <Wallet className="h-5 w-5 text-amber-400" />
                My Loans
              </h3>
              {connected && (
                <button className="bg-amber-600 hover:bg-amber-700 text-white px-4 py-2 rounded-lg flex items-center gap-2 transition-colors">
                  <Plus className="h-4 w-4" /> Create Loan
                </button>
              )}
            </div>

            {!connected ? (
              <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                <Wallet className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-400 mb-4">Connect your wallet to view and manage loans</p>
                <WalletButton />
              </div>
            ) : myLoans.length === 0 ? (
              <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                <p className="text-gray-500">No active loans</p>
                <p className="text-sm text-gray-600 mt-2">
                  Create a loan request or browse lending pools to get started
                </p>
              </div>
            ) : (
              <div className="space-y-4">
                {myLoans.map((loan) => (
                  <div
                    key={loan.id}
                    className={`border rounded-xl p-6 bg-gray-900/50 ${
                      loan.type === "borrowed" ? "border-orange-800/50" : "border-green-800/50"
                    }`}
                  >
                    <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
                      <div>
                        <div className="flex items-center gap-3 mb-2">
                          <span
                            className={`px-2 py-1 rounded text-xs font-semibold ${
                              loan.type === "borrowed"
                                ? "bg-orange-900/50 text-orange-400"
                                : "bg-green-900/50 text-green-400"
                            }`}
                          >
                            {loan.type === "borrowed" ? "Borrowed" : "Lent"}
                          </span>
                          <span className="text-sm text-gray-500">{loan.id}</span>
                          <span
                            className={`px-2 py-1 rounded text-xs ${
                              loan.collateralRatio >= 150
                                ? "bg-green-900/30 text-green-400"
                                : loan.collateralRatio >= 130
                                ? "bg-yellow-900/30 text-yellow-400"
                                : "bg-red-900/30 text-red-400"
                            }`}
                          >
                            {loan.collateralRatio}% Collateral
                          </span>
                        </div>
                        <p className="text-2xl font-bold">{loan.principal} HODL</p>
                        {loan.type === "borrowed" && (
                          <p className="text-sm text-gray-500 mt-1">
                            Collateral: {loan.collateral}
                          </p>
                        )}
                      </div>
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                        <div>
                          <p className="text-gray-500">Interest Rate</p>
                          <p className="font-semibold">{loan.interestRate}% APR</p>
                        </div>
                        <div>
                          <p className="text-gray-500">Duration</p>
                          <p className="font-semibold">{loan.duration}</p>
                        </div>
                        <div>
                          <p className="text-gray-500">Remaining</p>
                          <p className="font-semibold text-amber-400">{loan.remaining}</p>
                        </div>
                        <div>
                          <p className="text-gray-500">Interest Accrued</p>
                          <p className="font-semibold text-green-400">{loan.accrued} HODL</p>
                        </div>
                      </div>
                    </div>
                    <div className="mt-4 flex flex-wrap gap-3">
                      {loan.type === "borrowed" ? (
                        <>
                          <button className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded flex items-center gap-2">
                            Repay Loan
                          </button>
                          <button className="border border-amber-500 text-amber-400 hover:bg-amber-900/30 px-4 py-2 rounded">
                            Add Collateral
                          </button>
                        </>
                      ) : (
                        <button className="border border-gray-600 text-gray-400 px-4 py-2 rounded">
                          View Details
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Lending Pools Tab */}
        {activeTab === "pools" && (
          <div>
            <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
              <PiggyBank className="h-5 w-5 text-amber-400" />
              Lending Pools
            </h3>
            <div className="space-y-4">
              {lendingPools.map((pool, index) => (
                <div key={index} className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
                  <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
                    <div>
                      <h4 className="text-xl font-semibold">{pool.name}</h4>
                      <p className="text-sm text-gray-500 mt-1">
                        Total Liquidity: ${pool.totalLiquidity} | Utilization: {pool.utilization}%
                      </p>
                    </div>
                    <div className="grid grid-cols-2 gap-6">
                      <div className="text-center">
                        <div className="flex items-center gap-1 justify-center">
                          <TrendingUp className="h-4 w-4 text-green-400" />
                          <span className="text-2xl font-bold text-green-400">{pool.supplyAPY}%</span>
                        </div>
                        <p className="text-xs text-gray-500">Supply APY</p>
                      </div>
                      <div className="text-center">
                        <div className="flex items-center gap-1 justify-center">
                          <TrendingDown className="h-4 w-4 text-orange-400" />
                          <span className="text-2xl font-bold text-orange-400">{pool.borrowAPY}%</span>
                        </div>
                        <p className="text-xs text-gray-500">Borrow APY</p>
                      </div>
                    </div>
                  </div>

                  {/* Utilization bar */}
                  <div className="mt-4">
                    <div className="flex justify-between text-xs text-gray-500 mb-1">
                      <span>Utilization</span>
                      <span>{pool.utilization}%</span>
                    </div>
                    <div className="w-full bg-gray-800 rounded-full h-2">
                      <div
                        className={`h-2 rounded-full ${
                          pool.utilization > 80 ? "bg-orange-500" : "bg-green-500"
                        }`}
                        style={{ width: `${pool.utilization}%` }}
                      />
                    </div>
                  </div>

                  {pool.myDeposit && (
                    <div className="mt-4 p-3 bg-green-900/20 border border-green-800/50 rounded-lg">
                      <p className="text-sm text-green-400">
                        Your deposit: <span className="font-bold">{pool.myDeposit} HODL</span>
                      </p>
                    </div>
                  )}

                  <div className="mt-4 flex flex-wrap gap-3">
                    <button
                      disabled={!connected}
                      className="bg-green-600 hover:bg-green-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2"
                    >
                      <Plus className="h-4 w-4" /> Deposit
                    </button>
                    <button
                      disabled={!connected}
                      className="bg-orange-600 hover:bg-orange-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2"
                    >
                      Borrow
                    </button>
                    {pool.myDeposit && (
                      <button className="border border-gray-600 text-gray-400 hover:bg-gray-800 px-4 py-2 rounded">
                        Withdraw
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* P2P Marketplace Tab */}
        {activeTab === "p2p" && (
          <div>
            <div className="flex justify-between items-center mb-6">
              <h3 className="text-lg font-semibold flex items-center gap-2">
                <Users className="h-5 w-5 text-amber-400" />
                P2P Lending Marketplace
              </h3>
              {connected && (
                <div className="flex gap-2">
                  <button className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg flex items-center gap-2">
                    <Plus className="h-4 w-4" /> Post Offer
                  </button>
                  <button className="bg-orange-600 hover:bg-orange-700 text-white px-4 py-2 rounded-lg flex items-center gap-2">
                    <Plus className="h-4 w-4" /> Request Loan
                  </button>
                </div>
              )}
            </div>

            {/* Trust Ceiling Info */}
            {connected && (
              <div className="mb-6 p-4 bg-blue-900/20 border border-blue-800/50 rounded-xl">
                <div className="flex items-center gap-3">
                  <Shield className="h-5 w-5 text-blue-400" />
                  <div>
                    <p className="font-semibold text-blue-400">Your Trust Ceiling</p>
                    <p className="text-sm text-gray-400">
                      Stake HODL to increase your P2P lending/borrowing limit. Current: <span className="text-white font-semibold">25,000 HODL</span>
                    </p>
                  </div>
                  <button className="ml-auto bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded text-sm">
                    Manage Stake
                  </button>
                </div>
              </div>
            )}

            <div className="space-y-4">
              {p2pOffers.map((offer) => (
                <div
                  key={offer.id}
                  className={`border rounded-xl p-6 bg-gray-900/50 ${
                    offer.type === "lend" ? "border-green-800/50" : "border-orange-800/50"
                  }`}
                >
                  <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
                    <div>
                      <div className="flex items-center gap-3 mb-2">
                        <span
                          className={`px-2 py-1 rounded text-xs font-semibold ${
                            offer.type === "lend"
                              ? "bg-green-900/50 text-green-400"
                              : "bg-orange-900/50 text-orange-400"
                          }`}
                        >
                          {offer.type === "lend" ? "Lending Offer" : "Borrow Request"}
                        </span>
                        <span className="text-sm text-gray-500">{offer.id}</span>
                      </div>
                      <p className="text-2xl font-bold">{offer.amount} HODL</p>
                      <p className="text-sm text-gray-500 mt-1">
                        {offer.type === "lend"
                          ? `From: ${offer.lender}`
                          : `By: ${offer.borrower}`}
                      </p>
                    </div>
                    <div className="grid grid-cols-3 gap-4 text-sm">
                      <div>
                        <p className="text-gray-500">Interest Rate</p>
                        <p className="font-semibold">{offer.interestRate}% APR</p>
                      </div>
                      <div>
                        <p className="text-gray-500">Duration</p>
                        <p className="font-semibold">{offer.duration}</p>
                      </div>
                      <div>
                        <p className="text-gray-500">Collateral</p>
                        <p className="font-semibold">
                          {offer.type === "lend"
                            ? `${offer.collateralRequired}% Required`
                            : offer.collateralOffered}
                        </p>
                      </div>
                    </div>
                  </div>
                  <div className="mt-4 flex flex-wrap gap-3">
                    <button
                      disabled={!connected}
                      className={`px-4 py-2 rounded flex items-center gap-2 disabled:bg-gray-700 disabled:cursor-not-allowed ${
                        offer.type === "lend"
                          ? "bg-green-600 hover:bg-green-700 text-white"
                          : "bg-orange-600 hover:bg-orange-700 text-white"
                      }`}
                    >
                      {offer.type === "lend" ? "Accept Offer" : "Fund Request"}
                      <ArrowRight className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Liquidations Tab */}
        {activeTab === "liquidations" && (
          <div>
            <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-red-400" />
              Liquidation Opportunities
            </h3>
            <div className="mb-6 p-4 bg-yellow-900/20 border border-yellow-800/50 rounded-xl">
              <div className="flex items-center gap-3">
                <Zap className="h-5 w-5 text-yellow-400" />
                <div>
                  <p className="font-semibold text-yellow-400">Liquidation Rewards</p>
                  <p className="text-sm text-gray-400">
                    Liquidate under-collateralized loans (below 125%) and earn a 10% bonus on the collateral.
                  </p>
                </div>
              </div>
            </div>

            {liquidatableLoans.length === 0 ? (
              <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                <AlertTriangle className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-500">No liquidatable loans available</p>
                <p className="text-sm text-gray-600 mt-2">
                  Check back when loans fall below the 125% collateral threshold
                </p>
              </div>
            ) : (
              <div className="space-y-4">
                {liquidatableLoans.map((loan) => (
                  <div
                    key={loan.id}
                    className="border border-red-800/50 rounded-xl p-6 bg-red-900/10"
                  >
                    <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
                      <div>
                        <div className="flex items-center gap-3 mb-2">
                          <span className="bg-red-600 text-white px-2 py-1 rounded text-xs font-bold">
                            LIQUIDATABLE
                          </span>
                          <span className="text-sm text-gray-500">{loan.id}</span>
                        </div>
                        <p className="text-2xl font-bold">{loan.principal} HODL</p>
                        <p className="text-sm text-gray-500 mt-1">
                          Borrower: {loan.borrower}
                        </p>
                      </div>
                      <div className="grid grid-cols-3 gap-4 text-sm">
                        <div>
                          <p className="text-gray-500">Collateral</p>
                          <p className="font-semibold">{loan.collateral}</p>
                        </div>
                        <div>
                          <p className="text-gray-500">Ratio</p>
                          <p className="font-semibold text-red-400">{loan.collateralRatio}%</p>
                        </div>
                        <div>
                          <p className="text-gray-500">Bonus</p>
                          <p className="font-semibold text-green-400">{loan.liquidationBonus}</p>
                        </div>
                      </div>
                    </div>
                    <div className="mt-4">
                      <button
                        disabled={!connected}
                        className="bg-red-600 hover:bg-red-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2"
                      >
                        <Zap className="h-4 w-4" /> Liquidate Loan
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Footer */}
        <div className="text-center text-gray-500 mt-12 pt-8 border-t border-gray-800">
          <p className="mb-2 text-gray-400">ShareHODL Lending Protocol</p>
          <p className="text-sm mb-4">
            Collateralized lending with transparent on-chain terms and automatic liquidation protection.
          </p>
          <div className="flex justify-center items-center gap-6">
            <a
              href="https://x.com/share_hodl"
              target="_blank"
              rel="noopener noreferrer"
              className="text-gray-500 hover:text-white transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
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
                <path
                  fillRule="evenodd"
                  d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
                  clipRule="evenodd"
                />
              </svg>
              <span className="text-sm">GitHub</span>
            </a>
          </div>
        </div>
      </main>
    </div>
  );
}
