"use client";

import { useState } from "react";

export default function Home() {
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
    <div className="min-h-screen bg-background">
      <main className="container mx-auto px-4 py-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold mb-4 flex items-center justify-center gap-3">
            <span className="text-3xl">🏛️</span>
            ShareHODL Governance Portal
          </h1>
          <p className="text-muted-foreground text-lg max-w-3xl mx-auto">
            Institutional-grade governance with multi-tier proposals, emergency response, and professional voting delegation.
          </p>
        </div>

        {/* Emergency Alert Banner */}
        <div className="mb-6 bg-red-50 border-l-4 border-red-500 p-4 rounded">
          <div className="flex items-center">
            <span className="text-red-500 mr-2">🚨</span>
            <div>
              <h4 className="font-semibold text-red-800">Emergency Proposal Active</h4>
              <p className="text-red-600">Critical security proposal requires immediate attention - 47 minutes remaining</p>
            </div>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid gap-4 md:grid-cols-5 mb-8">
          <div className="border rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-red-600">1</div>
            <p className="text-sm text-muted-foreground">Emergency</p>
          </div>
          <div className="border rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-blue-600">8</div>
            <p className="text-sm text-muted-foreground">Company</p>
          </div>
          <div className="border rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-green-600">4</div>
            <p className="text-sm text-muted-foreground">Protocol</p>
          </div>
          <div className="border rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-purple-600">2</div>
            <p className="text-sm text-muted-foreground">Validator</p>
          </div>
          <div className="border rounded-lg p-4 text-center">
            <div className="text-2xl font-bold">3,247</div>
            <p className="text-sm text-muted-foreground">Total Voters</p>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="border-b border-gray-200 mb-6">
          <nav className="-mb-px flex space-x-8">
            {[
              { id: "proposals", label: "Active Proposals", icon: "📋" },
              { id: "delegation", label: "Vote Delegation", icon: "🤝" },
              { id: "create", label: "Create Proposal", icon: "➕" },
              { id: "history", label: "Voting History", icon: "📊" }
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === tab.id
                    ? "border-blue-500 text-blue-600"
                    : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
                }`}
              >
                <span className="mr-2">{tab.icon}</span>
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Content */}
        {activeTab === "proposals" && (
          <div>
            {/* Proposal Type Filter */}
            <div className="flex gap-2 mb-6">
              {[
                { id: "emergency", label: "Emergency", color: "red", count: 1 },
                { id: "company", label: "Company", color: "blue", count: 8 },
                { id: "protocol", label: "Protocol", color: "green", count: 4 },
                { id: "validator", label: "Validator", color: "purple", count: 2 }
              ].map((type) => (
                <button
                  key={type.id}
                  onClick={() => setSelectedProposalType(type.id)}
                  className={`px-4 py-2 rounded-lg border ${
                    selectedProposalType === type.id
                      ? `bg-${type.color}-50 border-${type.color}-500 text-${type.color}-700`
                      : "border-gray-200 hover:border-gray-300"
                  }`}
                >
                  {type.label} ({type.count})
                </button>
              ))}
            </div>

            {/* Emergency Proposals */}
            {selectedProposalType === "emergency" && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-red-600 flex items-center gap-2">
                  🚨 Emergency Proposals
                </h3>
                {emergencyProposals.map((proposal) => (
                  <div key={proposal.id} className="border border-red-200 rounded-lg p-6 bg-red-50">
                    <div className="flex justify-between items-start">
                      <div>
                        <div className="flex items-center gap-3 mb-2">
                          <span className="bg-red-500 text-white px-2 py-1 rounded text-xs font-bold">
                            SEVERITY {proposal.severity}
                          </span>
                          <span className="text-sm text-gray-600">{proposal.id}</span>
                        </div>
                        <h4 className="font-semibold text-lg">{proposal.title}</h4>
                        <p className="text-sm text-gray-600 mt-1">{proposal.type}</p>
                      </div>
                      <div className="text-right">
                        <div className="text-red-600 font-bold">{proposal.timeRemaining}</div>
                        <div className="text-xs text-gray-500">remaining</div>
                      </div>
                    </div>
                    <div className="mt-4 flex gap-4">
                      <button className="bg-green-500 text-white px-4 py-2 rounded font-semibold">Vote YES</button>
                      <button className="bg-red-500 text-white px-4 py-2 rounded font-semibold">Vote NO</button>
                      <button className="bg-gray-500 text-white px-4 py-2 rounded font-semibold">Abstain</button>
                    </div>
                    <div className="mt-3 text-sm">
                      Current: <span className="text-green-600">{proposal.votes.yes}% Yes</span> | 
                      <span className="text-red-600 ml-1">{proposal.votes.no}% No</span> | 
                      <span className="text-gray-600 ml-1">{proposal.votes.abstain}% Abstain</span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Company Proposals */}
            {selectedProposalType === "company" && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-blue-600 flex items-center gap-2">
                  🏢 Company Governance Proposals
                </h3>
                {companyProposals.map((proposal) => (
                  <div key={proposal.id} className="border rounded-lg p-6">
                    <div className="flex justify-between items-start">
                      <div>
                        <div className="flex items-center gap-3 mb-2">
                          <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded text-xs font-semibold">
                            {proposal.company}
                          </span>
                          <span className="bg-gray-100 px-2 py-1 rounded text-xs">{proposal.type}</span>
                          <span className="text-sm text-gray-600">{proposal.id}</span>
                        </div>
                        <h4 className="font-semibold text-lg">{proposal.title}</h4>
                      </div>
                      <div className="text-right">
                        <div className="font-semibold">{proposal.timeRemaining}</div>
                        <div className="text-xs text-gray-500">remaining</div>
                      </div>
                    </div>
                    <div className="mt-3 grid grid-cols-2 gap-4 text-sm">
                      <div>Quorum Required: <span className="font-semibold">{proposal.quorum}</span></div>
                      <div>Approval Threshold: <span className="font-semibold">{proposal.threshold}</span></div>
                    </div>
                    <div className="mt-4 flex gap-3">
                      <button className="bg-green-500 text-white px-4 py-2 rounded">Vote YES</button>
                      <button className="bg-red-500 text-white px-4 py-2 rounded">Vote NO</button>
                      <button className="bg-gray-500 text-white px-4 py-2 rounded">Abstain</button>
                      <button className="border border-blue-500 text-blue-500 px-4 py-2 rounded">Delegate Vote</button>
                    </div>
                    <div className="mt-3 text-sm">
                      <span className="text-green-600">{proposal.votes.yes}% Yes</span> | 
                      <span className="text-red-600 ml-1">{proposal.votes.no}% No</span> | 
                      <span className="text-gray-600 ml-1">{proposal.votes.abstain}% Abstain</span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Protocol Proposals */}
            {selectedProposalType === "protocol" && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-green-600 flex items-center gap-2">
                  ⚙️ Protocol Governance Proposals
                </h3>
                {protocolProposals.map((proposal) => (
                  <div key={proposal.id} className="border rounded-lg p-6">
                    <div className="flex justify-between items-start">
                      <div>
                        <div className="flex items-center gap-3 mb-2">
                          <span className="bg-green-100 text-green-800 px-2 py-1 rounded text-xs font-semibold">
                            {proposal.type}
                          </span>
                          <span className="text-sm text-gray-600">{proposal.id}</span>
                        </div>
                        <h4 className="font-semibold text-lg">{proposal.title}</h4>
                        <p className="text-sm text-gray-600 mt-1">Submitted by: {proposal.submitter}</p>
                      </div>
                      <div className="text-right">
                        <div className="font-semibold">{proposal.timeRemaining}</div>
                        <div className="text-xs text-gray-500">remaining</div>
                      </div>
                    </div>
                    <div className="mt-4 flex gap-3">
                      <button className="bg-green-500 text-white px-4 py-2 rounded">Vote YES</button>
                      <button className="bg-red-500 text-white px-4 py-2 rounded">Vote NO</button>
                      <button className="bg-gray-500 text-white px-4 py-2 rounded">Abstain</button>
                    </div>
                    <div className="mt-3 text-sm">
                      <span className="text-green-600">{proposal.votes.yes}% Yes</span> | 
                      <span className="text-red-600 ml-1">{proposal.votes.no}% No</span> | 
                      <span className="text-gray-600 ml-1">{proposal.votes.abstain}% Abstain</span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Validator Proposals */}
            {selectedProposalType === "validator" && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-purple-600 flex items-center gap-2">
                  🛡️ Validator Governance Proposals
                </h3>
                {validatorProposals.map((proposal) => (
                  <div key={proposal.id} className="border rounded-lg p-6">
                    <div className="flex justify-between items-start">
                      <div>
                        <div className="flex items-center gap-3 mb-2">
                          <span className="bg-purple-100 text-purple-800 px-2 py-1 rounded text-xs font-semibold">
                            {proposal.type}
                          </span>
                          <span className="bg-yellow-100 text-yellow-800 px-2 py-1 rounded text-xs">
                            Requires: {proposal.requiredTier}
                          </span>
                        </div>
                        <h4 className="font-semibold text-lg">{proposal.title}</h4>
                      </div>
                      <div className="text-right">
                        <div className="font-semibold">{proposal.timeRemaining}</div>
                      </div>
                    </div>
                    <div className="mt-4 flex gap-3">
                      <button className="bg-green-500 text-white px-4 py-2 rounded">Vote YES</button>
                      <button className="bg-red-500 text-white px-4 py-2 rounded">Vote NO</button>
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
              🤝 Voting Power Delegation
            </h3>
            <div className="grid gap-6 md:grid-cols-2">
              <div className="border rounded-lg p-6">
                <h4 className="font-semibold mb-4">Delegate Your Votes</h4>
                <p className="text-sm text-gray-600 mb-4">
                  Delegate your voting power to experienced governance participants.
                </p>
                <div className="space-y-3">
                  {delegations.map((delegate, i) => (
                    <div key={i} className="border rounded p-3">
                      <div className="flex justify-between items-start">
                        <div>
                          <div className="font-semibold">{delegate.delegate}</div>
                          <div className="text-sm text-gray-600">
                            {delegate.votingPower} • {delegate.proposals} proposals • {delegate.reputation}% approval
                          </div>
                        </div>
                        <button className="bg-blue-500 text-white px-3 py-1 rounded text-sm">
                          Delegate
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
              
              <div className="border rounded-lg p-6">
                <h4 className="font-semibold mb-4">Your Delegations</h4>
                <div className="text-center py-8 text-gray-500">
                  <p>No active delegations</p>
                  <p className="text-sm mt-2">Delegate your voting power to participate efficiently</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Create Proposal Tab */}
        {activeTab === "create" && (
          <div>
            <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
              ➕ Create New Proposal
            </h3>
            <div className="max-w-2xl">
              <div className="space-y-6">
                <div>
                  <label className="block font-semibold mb-2">Proposal Type</label>
                  <select className="w-full border rounded p-3">
                    <option>Protocol Parameter Change</option>
                    <option>Software Upgrade</option>
                    <option>Company Governance (Board Election)</option>
                    <option>Company Governance (Merger & Acquisition)</option>
                    <option>Validator Tier Promotion</option>
                    <option>Emergency Action</option>
                  </select>
                </div>
                
                <div>
                  <label className="block font-semibold mb-2">Title</label>
                  <input 
                    type="text" 
                    className="w-full border rounded p-3" 
                    placeholder="Enter proposal title..."
                  />
                </div>
                
                <div>
                  <label className="block font-semibold mb-2">Description</label>
                  <textarea 
                    className="w-full border rounded p-3 h-32" 
                    placeholder="Detailed proposal description..."
                  />
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block font-semibold mb-2">Voting Period</label>
                    <select className="w-full border rounded p-3">
                      <option>7 days</option>
                      <option>14 days</option>
                      <option>21 days</option>
                      <option>30 days</option>
                    </select>
                  </div>
                  
                  <div>
                    <label className="block font-semibold mb-2">Quorum Required</label>
                    <select className="w-full border rounded p-3">
                      <option>33.4%</option>
                      <option>50%</option>
                      <option>67%</option>
                    </select>
                  </div>
                </div>
                
                <button className="bg-blue-500 text-white px-6 py-3 rounded-lg font-semibold">
                  Submit Proposal
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Voting History Tab */}
        {activeTab === "history" && (
          <div>
            <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
              📊 Your Voting History
            </h3>
            <div className="border rounded-lg p-6 text-center">
              <p className="text-gray-500">Voting history will appear here</p>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}