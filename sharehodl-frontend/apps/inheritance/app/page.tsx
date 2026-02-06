"use client";

import { useState } from "react";
import { WalletButton, useWallet, useBlockchain } from "@repo/ui";
import {
  Shield,
  Clock,
  Users,
  AlertTriangle,
  Plus,
  Check,
  X,
  Heart,
  Timer,
  Gift,
  FileText,
  Bell,
  Activity,
  Lock,
  Unlock,
} from "lucide-react";

export default function Home() {
  const { connected, address } = useWallet();
  const { networkStatus } = useBlockchain();
  const [activeTab, setActiveTab] = useState<"plans" | "claims" | "create" | "activity">("plans");

  // Mock data for demonstration
  const myPlans = [
    {
      id: "PLAN-001",
      status: "active",
      totalAssets: "125,000 HODL",
      beneficiaries: 3,
      inactivityPeriod: "365 days",
      gracePeriod: "30 days",
      lastActivity: "2 days ago",
      daysUntilTrigger: 363,
    },
  ];

  const triggeredPlans = [
    {
      id: "PLAN-099",
      owner: "hodl1abc...xyz",
      status: "triggered",
      totalAssets: "50,000 HODL",
      gracePeriodRemaining: "15 days",
      triggeredAt: "2024-01-15",
    },
  ];

  const beneficiaryOf = [
    {
      planId: "PLAN-045",
      owner: "hodl1def...uvw",
      myShare: "25%",
      specificAssets: "500 APPLE shares",
      status: "active",
      priority: 1,
    },
  ];

  const claimablePlans = [
    {
      planId: "PLAN-088",
      owner: "hodl1ghi...rst",
      myShare: "33.3%",
      estimatedValue: "15,000 HODL",
      claimWindowEnds: "45 days",
      status: "ready_to_claim",
    },
  ];

  const recentActivity = [
    { time: "2 days ago", type: "transaction", description: "Sent 100 HODL to hodl1xyz..." },
    { time: "5 days ago", type: "trade", description: "Swapped 50 APPLE for 9,250 HODL" },
    { time: "1 week ago", type: "vote", description: "Voted on proposal PROTOCOL-015" },
  ];

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Header */}
      <header className="border-b border-gray-800 sticky top-0 z-50 bg-gray-950/95 backdrop-blur">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-2">
            <Shield className="h-6 w-6 text-indigo-400" />
            <span className="text-2xl font-bold bg-gradient-to-r from-indigo-500 to-purple-500 bg-clip-text text-transparent">
              Inheritance
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
              <a href="http://localhost:3004" className="text-gray-400 hover:text-white transition">Wallet</a>
              <a href="http://localhost:3006" className="text-gray-400 hover:text-white transition">Lending</a>
            </nav>
            <WalletButton />
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold mb-4 bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
            Asset Inheritance
          </h1>
          <p className="text-gray-400 text-lg max-w-3xl mx-auto">
            Secure your digital legacy with automated inheritance plans and dead man switch protection.
          </p>
        </div>

        {/* Stats Overview */}
        <div className="grid gap-4 grid-cols-2 md:grid-cols-4 mb-8">
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-indigo-400">$2.1M</div>
            <p className="text-sm text-gray-500">Assets Protected</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-green-400">847</div>
            <p className="text-sm text-gray-500">Active Plans</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-purple-400">12</div>
            <p className="text-sm text-gray-500">Triggered (Grace)</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-amber-400">156</div>
            <p className="text-sm text-gray-500">Completed Claims</p>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="border-b border-gray-800 mb-6">
          <nav className="-mb-px flex flex-wrap gap-2">
            {[
              { id: "plans" as const, label: "My Plans", icon: FileText },
              { id: "claims" as const, label: "Claims", icon: Gift },
              { id: "create" as const, label: "Create Plan", icon: Plus },
              { id: "activity" as const, label: "Activity", icon: Activity },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-2 px-4 flex items-center gap-2 rounded-t-lg transition-colors ${
                  activeTab === tab.id
                    ? "bg-gray-800 text-white border-b-2 border-indigo-500"
                    : "text-gray-500 hover:text-gray-300"
                }`}
              >
                <tab.icon className="h-4 w-4" />
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* My Plans Tab */}
        {activeTab === "plans" && (
          <div className="space-y-8">
            {/* Plans I Own */}
            <div>
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-semibold flex items-center gap-2">
                  <FileText className="h-5 w-5 text-indigo-400" />
                  Plans I Own
                </h3>
                {connected && (
                  <button
                    onClick={() => setActiveTab("create")}
                    className="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg flex items-center gap-2 transition-colors"
                  >
                    <Plus className="h-4 w-4" /> Create Plan
                  </button>
                )}
              </div>

              {!connected ? (
                <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                  <Shield className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                  <p className="text-gray-400 mb-4">Connect your wallet to manage inheritance plans</p>
                  <WalletButton />
                </div>
              ) : myPlans.length === 0 ? (
                <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                  <p className="text-gray-500">No active inheritance plans</p>
                  <p className="text-sm text-gray-600 mt-2">
                    Create a plan to protect your assets and assign beneficiaries
                  </p>
                </div>
              ) : (
                <div className="space-y-4">
                  {myPlans.map((plan) => (
                    <div
                      key={plan.id}
                      className="border border-indigo-800/50 rounded-xl p-6 bg-gray-900/50"
                    >
                      <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
                        <div>
                          <div className="flex items-center gap-3 mb-2">
                            <span className="bg-green-900/50 text-green-400 px-2 py-1 rounded text-xs font-semibold">
                              Active
                            </span>
                            <span className="text-sm text-gray-500">{plan.id}</span>
                          </div>
                          <p className="text-2xl font-bold">{plan.totalAssets}</p>
                          <p className="text-sm text-gray-500 mt-1">
                            {plan.beneficiaries} beneficiaries configured
                          </p>
                        </div>
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                          <div>
                            <p className="text-gray-500">Inactivity Period</p>
                            <p className="font-semibold">{plan.inactivityPeriod}</p>
                          </div>
                          <div>
                            <p className="text-gray-500">Grace Period</p>
                            <p className="font-semibold">{plan.gracePeriod}</p>
                          </div>
                          <div>
                            <p className="text-gray-500">Last Activity</p>
                            <p className="font-semibold text-green-400">{plan.lastActivity}</p>
                          </div>
                          <div>
                            <p className="text-gray-500">Days Until Trigger</p>
                            <p className="font-semibold text-indigo-400">{plan.daysUntilTrigger}</p>
                          </div>
                        </div>
                      </div>

                      {/* Dead Man Switch Status */}
                      <div className="mt-4 p-4 bg-green-900/20 border border-green-800/50 rounded-lg">
                        <div className="flex items-center gap-3">
                          <Activity className="h-5 w-5 text-green-400" />
                          <div className="flex-1">
                            <p className="text-sm text-green-400 font-semibold">Dead Man Switch: Safe</p>
                            <p className="text-xs text-gray-500">
                              Your recent activity keeps the switch inactive. {plan.daysUntilTrigger} days remaining.
                            </p>
                          </div>
                          <div className="w-32 bg-gray-800 rounded-full h-2">
                            <div
                              className="bg-green-500 h-2 rounded-full"
                              style={{ width: `${((365 - plan.daysUntilTrigger) / 365) * 100}%` }}
                            />
                          </div>
                        </div>
                      </div>

                      <div className="mt-4 flex flex-wrap gap-3">
                        <button className="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded flex items-center gap-2">
                          <Users className="h-4 w-4" /> Manage Beneficiaries
                        </button>
                        <button className="border border-gray-600 text-gray-400 hover:bg-gray-800 px-4 py-2 rounded">
                          Edit Plan
                        </button>
                        <button className="border border-red-600 text-red-400 hover:bg-red-900/30 px-4 py-2 rounded">
                          Cancel Plan
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Plans I'm a Beneficiary Of */}
            {connected && (
              <div>
                <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                  <Heart className="h-5 w-5 text-pink-400" />
                  Plans I'm a Beneficiary Of
                </h3>
                {beneficiaryOf.length === 0 ? (
                  <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                    <p className="text-gray-500">You're not listed as a beneficiary on any plans</p>
                  </div>
                ) : (
                  <div className="space-y-4">
                    {beneficiaryOf.map((plan) => (
                      <div
                        key={plan.planId}
                        className="border border-pink-800/50 rounded-xl p-6 bg-gray-900/50"
                      >
                        <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
                          <div>
                            <div className="flex items-center gap-3 mb-2">
                              <span className="bg-pink-900/50 text-pink-400 px-2 py-1 rounded text-xs font-semibold">
                                Priority {plan.priority}
                              </span>
                              <span className="text-sm text-gray-500">{plan.planId}</span>
                            </div>
                            <p className="text-sm text-gray-400">Owner: {plan.owner}</p>
                          </div>
                          <div className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                              <p className="text-gray-500">Your Share</p>
                              <p className="font-semibold text-pink-400">{plan.myShare}</p>
                            </div>
                            <div>
                              <p className="text-gray-500">Specific Assets</p>
                              <p className="font-semibold">{plan.specificAssets || "None"}</p>
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* Triggered Plans (Grace Period) */}
            {triggeredPlans.length > 0 && (
              <div>
                <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                  <AlertTriangle className="h-5 w-5 text-yellow-400" />
                  Triggered Plans (Grace Period)
                </h3>
                <div className="space-y-4">
                  {triggeredPlans.map((plan) => (
                    <div
                      key={plan.id}
                      className="border border-yellow-800/50 rounded-xl p-6 bg-yellow-900/10"
                    >
                      <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
                        <div>
                          <div className="flex items-center gap-3 mb-2">
                            <span className="bg-yellow-600 text-white px-2 py-1 rounded text-xs font-bold animate-pulse">
                              TRIGGERED
                            </span>
                            <span className="text-sm text-gray-500">{plan.id}</span>
                          </div>
                          <p className="text-2xl font-bold">{plan.totalAssets}</p>
                          <p className="text-sm text-gray-500 mt-1">
                            Owner: {plan.owner}
                          </p>
                        </div>
                        <div className="text-right">
                          <p className="text-yellow-400 font-bold text-xl">{plan.gracePeriodRemaining}</p>
                          <p className="text-xs text-gray-500">Grace period remaining</p>
                        </div>
                      </div>
                      <div className="mt-4 p-3 bg-yellow-900/30 border border-yellow-700 rounded-lg">
                        <p className="text-sm text-yellow-300">
                          If you're the owner, confirm your activity to cancel this trigger.
                          Otherwise, assets will transfer to beneficiaries after the grace period.
                        </p>
                      </div>
                      <div className="mt-4">
                        <button className="bg-yellow-600 hover:bg-yellow-700 text-white px-4 py-2 rounded flex items-center gap-2">
                          <Check className="h-4 w-4" /> Confirm Activity (Cancel Trigger)
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Claims Tab */}
        {activeTab === "claims" && (
          <div>
            <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
              <Gift className="h-5 w-5 text-indigo-400" />
              Available Claims
            </h3>

            {!connected ? (
              <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                <Gift className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-400 mb-4">Connect your wallet to view claimable inheritances</p>
                <WalletButton />
              </div>
            ) : claimablePlans.length === 0 ? (
              <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                <Gift className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-500">No inheritances available to claim</p>
                <p className="text-sm text-gray-600 mt-2">
                  Claims become available after an owner's inactivity period + grace period expires
                </p>
              </div>
            ) : (
              <div className="space-y-4">
                {claimablePlans.map((plan) => (
                  <div
                    key={plan.planId}
                    className="border border-green-800/50 rounded-xl p-6 bg-green-900/10"
                  >
                    <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
                      <div>
                        <div className="flex items-center gap-3 mb-2">
                          <span className="bg-green-600 text-white px-2 py-1 rounded text-xs font-bold">
                            READY TO CLAIM
                          </span>
                          <span className="text-sm text-gray-500">{plan.planId}</span>
                        </div>
                        <p className="text-2xl font-bold text-green-400">{plan.estimatedValue}</p>
                        <p className="text-sm text-gray-500 mt-1">
                          From: {plan.owner} | Your share: {plan.myShare}
                        </p>
                      </div>
                      <div className="text-right">
                        <p className="text-sm text-gray-500">Claim window ends in</p>
                        <p className="font-bold text-amber-400">{plan.claimWindowEnds}</p>
                      </div>
                    </div>
                    <div className="mt-4">
                      <button className="bg-green-600 hover:bg-green-700 text-white px-6 py-3 rounded-lg flex items-center gap-2 font-semibold">
                        <Gift className="h-5 w-5" /> Claim Inheritance
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Create Plan Tab */}
        {activeTab === "create" && (
          <div>
            <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
              <Plus className="h-5 w-5 text-indigo-400" />
              Create Inheritance Plan
            </h3>

            {!connected ? (
              <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                <Shield className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-400 mb-4">Connect your wallet to create an inheritance plan</p>
                <WalletButton />
              </div>
            ) : (
              <div className="max-w-2xl">
                <div className="space-y-6">
                  {/* Dead Man Switch Settings */}
                  <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
                    <h4 className="font-semibold mb-4 flex items-center gap-2">
                      <Timer className="h-5 w-5 text-amber-400" />
                      Dead Man Switch Settings
                    </h4>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm text-gray-400 mb-2">Inactivity Period</label>
                        <select className="w-full bg-gray-900 border border-gray-700 rounded-lg p-3 text-white">
                          <option value="180">180 days (6 months)</option>
                          <option value="365" selected>365 days (1 year)</option>
                          <option value="730">730 days (2 years)</option>
                        </select>
                        <p className="text-xs text-gray-500 mt-1">
                          Time of inactivity before the switch triggers
                        </p>
                      </div>
                      <div>
                        <label className="block text-sm text-gray-400 mb-2">Grace Period</label>
                        <select className="w-full bg-gray-900 border border-gray-700 rounded-lg p-3 text-white">
                          <option value="30" selected>30 days</option>
                          <option value="60">60 days</option>
                          <option value="90">90 days</option>
                        </select>
                        <p className="text-xs text-gray-500 mt-1">
                          Time to cancel trigger after activation
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Beneficiaries */}
                  <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
                    <h4 className="font-semibold mb-4 flex items-center gap-2">
                      <Users className="h-5 w-5 text-pink-400" />
                      Beneficiaries
                    </h4>
                    <div className="space-y-4">
                      <div className="border border-gray-700 rounded-lg p-4">
                        <div className="grid grid-cols-3 gap-4">
                          <div>
                            <label className="block text-sm text-gray-400 mb-2">Address</label>
                            <input
                              type="text"
                              placeholder="hodl1..."
                              className="w-full bg-gray-900 border border-gray-700 rounded-lg p-2 text-white text-sm"
                            />
                          </div>
                          <div>
                            <label className="block text-sm text-gray-400 mb-2">Share %</label>
                            <input
                              type="number"
                              placeholder="50"
                              className="w-full bg-gray-900 border border-gray-700 rounded-lg p-2 text-white text-sm"
                            />
                          </div>
                          <div>
                            <label className="block text-sm text-gray-400 mb-2">Priority</label>
                            <input
                              type="number"
                              placeholder="1"
                              className="w-full bg-gray-900 border border-gray-700 rounded-lg p-2 text-white text-sm"
                            />
                          </div>
                        </div>
                      </div>
                      <button className="text-indigo-400 hover:text-indigo-300 text-sm flex items-center gap-1">
                        <Plus className="h-4 w-4" /> Add Another Beneficiary
                      </button>
                    </div>
                  </div>

                  {/* Asset Selection */}
                  <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
                    <h4 className="font-semibold mb-4 flex items-center gap-2">
                      <Lock className="h-5 w-5 text-indigo-400" />
                      Assets to Include
                    </h4>
                    <div className="space-y-3">
                      <label className="flex items-center gap-3 p-3 border border-gray-700 rounded-lg cursor-pointer hover:border-gray-600">
                        <input type="checkbox" defaultChecked className="w-4 h-4 accent-indigo-500" />
                        <span>All HODL tokens</span>
                      </label>
                      <label className="flex items-center gap-3 p-3 border border-gray-700 rounded-lg cursor-pointer hover:border-gray-600">
                        <input type="checkbox" defaultChecked className="w-4 h-4 accent-indigo-500" />
                        <span>All equity holdings</span>
                      </label>
                      <label className="flex items-center gap-3 p-3 border border-gray-700 rounded-lg cursor-pointer hover:border-gray-600">
                        <input type="checkbox" className="w-4 h-4 accent-indigo-500" />
                        <span>Staked HODL positions</span>
                      </label>
                      <label className="flex items-center gap-3 p-3 border border-gray-700 rounded-lg cursor-pointer hover:border-gray-600">
                        <input type="checkbox" className="w-4 h-4 accent-indigo-500" />
                        <span>Lending positions</span>
                      </label>
                    </div>
                  </div>

                  {/* Charity Fallback */}
                  <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
                    <h4 className="font-semibold mb-4 flex items-center gap-2">
                      <Heart className="h-5 w-5 text-red-400" />
                      Unclaimed Asset Destination
                    </h4>
                    <div>
                      <label className="block text-sm text-gray-400 mb-2">Charity Address (Optional)</label>
                      <input
                        type="text"
                        placeholder="hodl1charity... (receives unclaimed assets after claim window)"
                        className="w-full bg-gray-900 border border-gray-700 rounded-lg p-3 text-white"
                      />
                      <p className="text-xs text-gray-500 mt-2">
                        If beneficiaries don't claim within 180 days, assets go to this address
                      </p>
                    </div>
                  </div>

                  <button className="w-full bg-indigo-600 hover:bg-indigo-700 text-white py-4 rounded-lg font-semibold flex items-center justify-center gap-2">
                    <Shield className="h-5 w-5" /> Create Inheritance Plan
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Activity Tab */}
        {activeTab === "activity" && (
          <div>
            <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
              <Activity className="h-5 w-5 text-indigo-400" />
              Recent Activity
            </h3>

            <div className="mb-6 p-4 bg-blue-900/20 border border-blue-800/50 rounded-xl">
              <div className="flex items-center gap-3">
                <Bell className="h-5 w-5 text-blue-400" />
                <div>
                  <p className="font-semibold text-blue-400">Activity Auto-Cancels Triggers</p>
                  <p className="text-sm text-gray-400">
                    Any on-chain activity (transactions, votes, trades) automatically resets your dead man switch timer.
                  </p>
                </div>
              </div>
            </div>

            {!connected ? (
              <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                <Activity className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-400 mb-4">Connect your wallet to view activity</p>
                <WalletButton />
              </div>
            ) : (
              <div className="space-y-4">
                {recentActivity.map((activity, index) => (
                  <div
                    key={index}
                    className="border border-gray-800 rounded-xl p-4 bg-gray-900/50 flex items-center gap-4"
                  >
                    <div className="w-10 h-10 bg-gray-800 rounded-full flex items-center justify-center">
                      <Activity className="h-5 w-5 text-gray-400" />
                    </div>
                    <div className="flex-1">
                      <p className="font-medium">{activity.description}</p>
                      <p className="text-sm text-gray-500">{activity.time}</p>
                    </div>
                    <span className="text-xs px-2 py-1 bg-green-900/30 text-green-400 rounded">
                      Timer Reset
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Footer */}
        <div className="text-center text-gray-500 mt-12 pt-8 border-t border-gray-800">
          <p className="mb-2 text-gray-400">ShareHODL Inheritance Protocol</p>
          <p className="text-sm mb-4">
            Secure your digital legacy with automated dead man switch protection and multi-beneficiary distribution.
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
