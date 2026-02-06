"use client";

import { useState } from "react";
import { WalletButton, useWallet, useBlockchain } from "@repo/ui";
import { Vote, Clock, Users, AlertTriangle, Shield, Building, Check, X, Minus } from "lucide-react";

export default function Home() {
  const { connected, address } = useWallet();
  const { networkStatus } = useBlockchain();
  const [activeTab, setActiveTab] = useState("proposals");
  const [selectedProposalType, setSelectedProposalType] = useState("protocol");

  const emergencyProposals = [
    {
      id: "EMERGENCY-001",
      type: "Security Breach",
      severity: 5,
      title: "Immediate Circuit Breaker Activation",
      timeRemaining: "47 minutes",
      status: "CRITICAL",
      votes: { yes: 89, no: 12, abstain: 4 }
    }
  ];

  const companyProposals = [
    {
      id: "AAPL-024",
      company: "Apple Inc.",
      type: "Board Election",
      title: "Elect Sarah Johnson as Independent Director",
      timeRemaining: "5 days",
      quorum: "67%",
      threshold: "50%",
      votes: { yes: 45, no: 23, abstain: 8 }
    },
    {
      id: "TSLA-012",
      company: "Tesla Inc.",
      type: "Merger & Acquisition",
      title: "Approve acquisition of SolarTech Corp",
      timeRemaining: "12 days",
      quorum: "75%",
      threshold: "67%",
      votes: { yes: 67, no: 28, abstain: 5 }
    }
  ];

  const protocolProposals = [
    {
      id: "PROTOCOL-015",
      type: "Parameter Change",
      title: "Reduce minimum validator stake to 10,000 HODL",
      timeRemaining: "3 days",
      submitter: "Gold Validator",
      votes: { yes: 78, no: 15, abstain: 7 }
    },
    {
      id: "PROTOCOL-016",
      type: "Software Upgrade",
      title: "Implement cross-chain bridge to Ethereum",
      timeRemaining: "8 days",
      submitter: "Platinum Validator",
      votes: { yes: 82, no: 12, abstain: 6 }
    }
  ];

  const validatorProposals = [
    {
      id: "VALIDATOR-008",
      type: "Tier Promotion",
      title: "Promote Validator-Alpha to Gold Tier",
      timeRemaining: "2 days",
      requiredTier: "Silver+",
      votes: { yes: 23, no: 4, abstain: 2 }
    }
  ];

  const delegations = [
    {
      delegate: "Institutional Capital Partners",
      votingPower: "2.3M HODL",
      proposals: 12,
      reputation: 98
    },
    {
      delegate: "Blockchain Governance Expert",
      votingPower: "890K HODL",
      proposals: 8,
      reputation: 95
    }
  ];

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Header */}
      <header className="border-b border-gray-800 sticky top-0 z-50 bg-gray-950/95 backdrop-blur">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-2">
            <Vote className="h-6 w-6 text-purple-400" />
            <span className="text-2xl font-bold bg-gradient-to-r from-purple-500 to-pink-500 bg-clip-text text-transparent">
              Governance
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
              <a href="http://localhost:3003" className="text-gray-400 hover:text-white transition">Explorer</a>
            </nav>
            <WalletButton />
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold mb-4 bg-gradient-to-r from-purple-400 via-pink-400 to-red-400 bg-clip-text text-transparent">
            ShareHODL Governance
          </h1>
          <p className="text-gray-400 text-lg max-w-3xl mx-auto">
            Vote on company decisions, protocol changes, and shape the future of decentralized equity markets.
          </p>
        </div>

        {/* Emergency Alert Banner */}
        {emergencyProposals.length > 0 && (
          <div className="mb-6 bg-red-900/20 border border-red-800 rounded-xl p-4">
            <div className="flex items-center">
              <AlertTriangle className="h-5 w-5 text-red-400 mr-3" />
              <div>
                <h4 className="font-semibold text-red-400">Emergency Proposal Active</h4>
                <p className="text-red-300/80 text-sm">Critical security proposal requires immediate attention - 47 minutes remaining</p>
              </div>
            </div>
          </div>
        )}

        {/* Stats Cards */}
        <div className="grid gap-4 grid-cols-2 md:grid-cols-5 mb-8">
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-red-400">1</div>
            <p className="text-sm text-gray-500">Emergency</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-blue-400">8</div>
            <p className="text-sm text-gray-500">Company</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-green-400">4</div>
            <p className="text-sm text-gray-500">Protocol</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-purple-400">2</div>
            <p className="text-sm text-gray-500">Validator</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-4 text-center bg-gray-900/50">
            <div className="text-2xl font-bold text-white">3,247</div>
            <p className="text-sm text-gray-500">Total Voters</p>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="border-b border-gray-800 mb-6">
          <nav className="-mb-px flex flex-wrap gap-2">
            {[
              { id: "proposals", label: "Active Proposals", icon: Vote },
              { id: "delegation", label: "Vote Delegation", icon: Users },
              { id: "create", label: "Create Proposal", icon: Shield },
              { id: "history", label: "Voting History", icon: Clock }
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-2 px-4 flex items-center gap-2 rounded-t-lg transition-colors ${
                  activeTab === tab.id
                    ? "bg-gray-800 text-white border-b-2 border-purple-500"
                    : "text-gray-500 hover:text-gray-300"
                }`}
              >
                <tab.icon className="h-4 w-4" />
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Content */}
        {activeTab === "proposals" && (
          <div>
            {/* Proposal Type Filter */}
            <div className="flex flex-wrap gap-2 mb-6">
              {[
                { id: "emergency", label: "Emergency", color: "red", count: 1 },
                { id: "company", label: "Company", color: "blue", count: 8 },
                { id: "protocol", label: "Protocol", color: "green", count: 4 },
                { id: "validator", label: "Validator", color: "purple", count: 2 }
              ].map((type) => (
                <button
                  key={type.id}
                  onClick={() => setSelectedProposalType(type.id)}
                  className={`px-4 py-2 rounded-lg border transition-colors ${
                    selectedProposalType === type.id
                      ? type.id === "emergency" ? "bg-red-900/30 border-red-600 text-red-400" :
                        type.id === "company" ? "bg-blue-900/30 border-blue-600 text-blue-400" :
                        type.id === "protocol" ? "bg-green-900/30 border-green-600 text-green-400" :
                        "bg-purple-900/30 border-purple-600 text-purple-400"
                      : "border-gray-700 text-gray-400 hover:border-gray-600"
                  }`}
                >
                  {type.label} ({type.count})
                </button>
              ))}
            </div>

            {/* Emergency Proposals */}
            {selectedProposalType === "emergency" && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-red-400 flex items-center gap-2">
                  <AlertTriangle className="h-5 w-5" />
                  Emergency Proposals
                </h3>
                {emergencyProposals.map((proposal) => (
                  <div key={proposal.id} className="border border-red-800 rounded-xl p-6 bg-red-900/10">
                    <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
                      <div>
                        <div className="flex items-center gap-3 mb-2">
                          <span className="bg-red-600 text-white px-2 py-1 rounded text-xs font-bold">
                            SEVERITY {proposal.severity}
                          </span>
                          <span className="text-sm text-gray-500">{proposal.id}</span>
                        </div>
                        <h4 className="font-semibold text-lg">{proposal.title}</h4>
                        <p className="text-sm text-gray-400 mt-1">{proposal.type}</p>
                      </div>
                      <div className="text-right">
                        <div className="text-red-400 font-bold">{proposal.timeRemaining}</div>
                        <div className="text-xs text-gray-500">remaining</div>
                      </div>
                    </div>
                    <div className="mt-4 flex flex-wrap gap-3">
                      <button disabled={!connected} className="bg-green-600 hover:bg-green-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded font-semibold flex items-center gap-2">
                        <Check className="h-4 w-4" /> Vote YES
                      </button>
                      <button disabled={!connected} className="bg-red-600 hover:bg-red-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded font-semibold flex items-center gap-2">
                        <X className="h-4 w-4" /> Vote NO
                      </button>
                      <button disabled={!connected} className="bg-gray-600 hover:bg-gray-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded font-semibold flex items-center gap-2">
                        <Minus className="h-4 w-4" /> Abstain
                      </button>
                    </div>
                    <div className="mt-3 text-sm">
                      <span className="text-green-400">{proposal.votes.yes}% Yes</span>
                      <span className="text-gray-600 mx-2">|</span>
                      <span className="text-red-400">{proposal.votes.no}% No</span>
                      <span className="text-gray-600 mx-2">|</span>
                      <span className="text-gray-400">{proposal.votes.abstain}% Abstain</span>
                    </div>
                    {!connected && (
                      <p className="text-sm text-yellow-500 mt-2">Connect wallet to vote</p>
                    )}
                  </div>
                ))}
              </div>
            )}

            {/* Company Proposals */}
            {selectedProposalType === "company" && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-blue-400 flex items-center gap-2">
                  <Building className="h-5 w-5" />
                  Company Governance Proposals
                </h3>
                {companyProposals.map((proposal) => (
                  <div key={proposal.id} className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
                    <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
                      <div>
                        <div className="flex flex-wrap items-center gap-3 mb-2">
                          <span className="bg-blue-900/50 text-blue-400 px-2 py-1 rounded text-xs font-semibold">
                            {proposal.company}
                          </span>
                          <span className="bg-gray-800 px-2 py-1 rounded text-xs text-gray-400">{proposal.type}</span>
                          <span className="text-sm text-gray-600">{proposal.id}</span>
                        </div>
                        <h4 className="font-semibold text-lg">{proposal.title}</h4>
                      </div>
                      <div className="text-right">
                        <div className="font-semibold text-white">{proposal.timeRemaining}</div>
                        <div className="text-xs text-gray-500">remaining</div>
                      </div>
                    </div>
                    <div className="mt-3 grid grid-cols-2 gap-4 text-sm">
                      <div className="text-gray-400">Quorum Required: <span className="font-semibold text-white">{proposal.quorum}</span></div>
                      <div className="text-gray-400">Approval Threshold: <span className="font-semibold text-white">{proposal.threshold}</span></div>
                    </div>
                    <div className="mt-4 flex flex-wrap gap-3">
                      <button disabled={!connected} className="bg-green-600 hover:bg-green-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2">
                        <Check className="h-4 w-4" /> Vote YES
                      </button>
                      <button disabled={!connected} className="bg-red-600 hover:bg-red-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2">
                        <X className="h-4 w-4" /> Vote NO
                      </button>
                      <button disabled={!connected} className="bg-gray-600 hover:bg-gray-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2">
                        <Minus className="h-4 w-4" /> Abstain
                      </button>
                      <button disabled={!connected} className="border border-blue-500 text-blue-400 hover:bg-blue-900/30 disabled:border-gray-700 disabled:text-gray-500 disabled:cursor-not-allowed px-4 py-2 rounded">
                        Delegate Vote
                      </button>
                    </div>
                    <div className="mt-3 text-sm">
                      <span className="text-green-400">{proposal.votes.yes}% Yes</span>
                      <span className="text-gray-600 mx-2">|</span>
                      <span className="text-red-400">{proposal.votes.no}% No</span>
                      <span className="text-gray-600 mx-2">|</span>
                      <span className="text-gray-400">{proposal.votes.abstain}% Abstain</span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Protocol Proposals */}
            {selectedProposalType === "protocol" && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-green-400 flex items-center gap-2">
                  <Shield className="h-5 w-5" />
                  Protocol Governance Proposals
                </h3>
                {protocolProposals.map((proposal) => (
                  <div key={proposal.id} className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
                    <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
                      <div>
                        <div className="flex items-center gap-3 mb-2">
                          <span className="bg-green-900/50 text-green-400 px-2 py-1 rounded text-xs font-semibold">
                            {proposal.type}
                          </span>
                          <span className="text-sm text-gray-600">{proposal.id}</span>
                        </div>
                        <h4 className="font-semibold text-lg">{proposal.title}</h4>
                        <p className="text-sm text-gray-500 mt-1">Submitted by: {proposal.submitter}</p>
                      </div>
                      <div className="text-right">
                        <div className="font-semibold text-white">{proposal.timeRemaining}</div>
                        <div className="text-xs text-gray-500">remaining</div>
                      </div>
                    </div>
                    <div className="mt-4 flex flex-wrap gap-3">
                      <button disabled={!connected} className="bg-green-600 hover:bg-green-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2">
                        <Check className="h-4 w-4" /> Vote YES
                      </button>
                      <button disabled={!connected} className="bg-red-600 hover:bg-red-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2">
                        <X className="h-4 w-4" /> Vote NO
                      </button>
                      <button disabled={!connected} className="bg-gray-600 hover:bg-gray-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2">
                        <Minus className="h-4 w-4" /> Abstain
                      </button>
                    </div>
                    <div className="mt-3 text-sm">
                      <span className="text-green-400">{proposal.votes.yes}% Yes</span>
                      <span className="text-gray-600 mx-2">|</span>
                      <span className="text-red-400">{proposal.votes.no}% No</span>
                      <span className="text-gray-600 mx-2">|</span>
                      <span className="text-gray-400">{proposal.votes.abstain}% Abstain</span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Validator Proposals */}
            {selectedProposalType === "validator" && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-purple-400 flex items-center gap-2">
                  <Users className="h-5 w-5" />
                  Validator Governance Proposals
                </h3>
                {validatorProposals.map((proposal) => (
                  <div key={proposal.id} className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
                    <div className="flex flex-col md:flex-row md:justify-between md:items-start gap-4">
                      <div>
                        <div className="flex items-center gap-3 mb-2">
                          <span className="bg-purple-900/50 text-purple-400 px-2 py-1 rounded text-xs font-semibold">
                            {proposal.type}
                          </span>
                          <span className="bg-yellow-900/50 text-yellow-400 px-2 py-1 rounded text-xs">
                            Requires: {proposal.requiredTier}
                          </span>
                        </div>
                        <h4 className="font-semibold text-lg">{proposal.title}</h4>
                      </div>
                      <div className="text-right">
                        <div className="font-semibold text-white">{proposal.timeRemaining}</div>
                      </div>
                    </div>
                    <div className="mt-4 flex flex-wrap gap-3">
                      <button disabled={!connected} className="bg-green-600 hover:bg-green-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2">
                        <Check className="h-4 w-4" /> Vote YES
                      </button>
                      <button disabled={!connected} className="bg-red-600 hover:bg-red-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-4 py-2 rounded flex items-center gap-2">
                        <X className="h-4 w-4" /> Vote NO
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Vote Delegation Tab */}
        {activeTab === "delegation" && (
          <div>
            <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
              <Users className="h-5 w-5 text-purple-400" />
              Voting Power Delegation
            </h3>
            <div className="grid gap-6 md:grid-cols-2">
              <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
                <h4 className="font-semibold mb-4">Delegate Your Votes</h4>
                <p className="text-sm text-gray-400 mb-4">
                  Delegate your voting power to experienced governance participants.
                </p>
                <div className="space-y-3">
                  {delegations.map((delegate, i) => (
                    <div key={i} className="border border-gray-700 rounded-lg p-3">
                      <div className="flex justify-between items-start">
                        <div>
                          <div className="font-semibold">{delegate.delegate}</div>
                          <div className="text-sm text-gray-500">
                            {delegate.votingPower} | {delegate.proposals} proposals | {delegate.reputation}% approval
                          </div>
                        </div>
                        <button disabled={!connected} className="bg-purple-600 hover:bg-purple-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white px-3 py-1 rounded text-sm">
                          Delegate
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
                <h4 className="font-semibold mb-4">Your Delegations</h4>
                {connected ? (
                  <div className="text-center py-8 text-gray-500">
                    <p>No active delegations</p>
                    <p className="text-sm mt-2">Delegate your voting power to participate efficiently</p>
                  </div>
                ) : (
                  <div className="text-center py-8 text-gray-500">
                    <p>Connect wallet to view delegations</p>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Create Proposal Tab */}
        {activeTab === "create" && (
          <div>
            <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
              <Shield className="h-5 w-5 text-purple-400" />
              Create New Proposal
            </h3>
            {!connected ? (
              <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                <Shield className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-400 mb-4">Connect your wallet to create proposals</p>
                <WalletButton />
              </div>
            ) : (
              <div className="max-w-2xl">
                <div className="space-y-6">
                  <div>
                    <label className="block font-semibold mb-2 text-gray-300">Proposal Type</label>
                    <select className="w-full bg-gray-900 border border-gray-700 rounded-lg p-3 text-white focus:outline-none focus:ring-2 focus:ring-purple-500">
                      <option>Protocol Parameter Change</option>
                      <option>Software Upgrade</option>
                      <option>Company Governance (Board Election)</option>
                      <option>Company Governance (Merger & Acquisition)</option>
                      <option>Validator Tier Promotion</option>
                      <option>Emergency Action</option>
                    </select>
                  </div>

                  <div>
                    <label className="block font-semibold mb-2 text-gray-300">Title</label>
                    <input
                      type="text"
                      className="w-full bg-gray-900 border border-gray-700 rounded-lg p-3 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500"
                      placeholder="Enter proposal title..."
                    />
                  </div>

                  <div>
                    <label className="block font-semibold mb-2 text-gray-300">Description</label>
                    <textarea
                      className="w-full bg-gray-900 border border-gray-700 rounded-lg p-3 h-32 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500"
                      placeholder="Detailed proposal description..."
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block font-semibold mb-2 text-gray-300">Voting Period</label>
                      <select className="w-full bg-gray-900 border border-gray-700 rounded-lg p-3 text-white focus:outline-none focus:ring-2 focus:ring-purple-500">
                        <option>7 days</option>
                        <option>14 days</option>
                        <option>21 days</option>
                        <option>30 days</option>
                      </select>
                    </div>

                    <div>
                      <label className="block font-semibold mb-2 text-gray-300">Quorum Required</label>
                      <select className="w-full bg-gray-900 border border-gray-700 rounded-lg p-3 text-white focus:outline-none focus:ring-2 focus:ring-purple-500">
                        <option>33.4%</option>
                        <option>50%</option>
                        <option>67%</option>
                      </select>
                    </div>
                  </div>

                  <div className="bg-gray-900 border border-gray-700 rounded-lg p-4">
                    <div className="flex justify-between text-sm mb-2">
                      <span className="text-gray-400">Proposal Deposit</span>
                      <span>100 HODL</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">Refunded if passes quorum</span>
                      <span className="text-green-400">Yes</span>
                    </div>
                  </div>

                  <button className="bg-purple-600 hover:bg-purple-700 text-white px-6 py-3 rounded-lg font-semibold transition-colors">
                    Submit Proposal
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Voting History Tab */}
        {activeTab === "history" && (
          <div>
            <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
              <Clock className="h-5 w-5 text-purple-400" />
              Your Voting History
            </h3>
            {connected ? (
              <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                <Clock className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-500">No voting history yet</p>
                <p className="text-sm text-gray-600 mt-2">Your votes will appear here once you participate in governance</p>
              </div>
            ) : (
              <div className="border border-gray-800 rounded-xl p-8 text-center bg-gray-900/50">
                <p className="text-gray-500">Connect wallet to view voting history</p>
              </div>
            )}
          </div>
        )}

        {/* Footer */}
        <div className="text-center text-gray-500 mt-12 pt-8 border-t border-gray-800">
          <p className="mb-2 text-gray-400">ShareHODL Governance Portal</p>
          <p className="text-sm mb-4">
            Shape the future of decentralized equity markets through on-chain governance.
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
