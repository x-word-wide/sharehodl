"use client";

import { useState } from "react";
import {
  AntiDilutionType,
  AntiDilutionTypeLabels,
  AntiDilutionTypeDescriptions,
  AntiDilutionProvision,
  IssuanceRecord,
  AntiDilutionAdjustment,
  formatShares,
  formatCurrency,
  formatPercent,
  formatDateTime,
  calculateFullRatchetAdjustment,
  calculateWeightedAverageAdjustment,
} from "../types/equity";

interface AntiDilutionManagerProps {
  companyId: string;
  companyName: string;
  shareClasses: {
    classId: string;
    name: string;
    outstandingShares: string;
    hasAntiDilution: boolean;
    antiDilutionType: AntiDilutionType;
  }[];
  isFounder?: boolean;
}

// Mock data for demonstration - in production this would come from blockchain queries
const mockProvisions: AntiDilutionProvision[] = [
  {
    companyId: "1",
    classId: "SERIES_A",
    provisionType: AntiDilutionType.FullRatchet,
    initialPrice: "10.00",
    currentAdjustedPrice: "8.50",
    triggerThreshold: "0.00",
    adjustmentCap: "0.50",
    minimumPrice: "1.00",
    isActive: true,
    lastAdjustmentTime: "2024-06-15T10:30:00Z",
    totalAdjustments: 2,
    createdAt: "2024-01-01T00:00:00Z",
  },
];

const mockIssuances: IssuanceRecord[] = [
  {
    id: "1",
    companyId: "1",
    classId: "COMMON",
    sharesIssued: "1000000",
    issuePrice: "10.00",
    totalRaised: "10000000",
    timestamp: "2024-01-15T14:00:00Z",
    isDownRound: false,
    priceDropPercent: "0",
    purpose: "Series A Funding",
  },
  {
    id: "2",
    companyId: "1",
    classId: "COMMON",
    sharesIssued: "500000",
    issuePrice: "8.50",
    totalRaised: "4250000",
    timestamp: "2024-06-15T10:30:00Z",
    isDownRound: true,
    priceDropPercent: "15.00",
    purpose: "Bridge Round",
  },
];

const mockAdjustments: AntiDilutionAdjustment[] = [
  {
    id: "1",
    companyId: "1",
    classId: "SERIES_A",
    shareholder: "sharehodl1abc...xyz",
    originalShares: "100000",
    adjustedShares: "117647",
    sharesAdded: "17647",
    triggerPrice: "8.50",
    adjustmentRatio: "1.176",
    adjustmentType: AntiDilutionType.FullRatchet,
    issuanceId: "2",
    timestamp: "2024-06-15T10:30:00Z",
  },
];

export default function AntiDilutionManager({
  companyId,
  companyName,
  shareClasses,
  isFounder = false,
}: AntiDilutionManagerProps) {
  const [activeTab, setActiveTab] = useState<"overview" | "provisions" | "issuances" | "adjustments" | "register">("overview");
  const [selectedClass, setSelectedClass] = useState<string>("");
  const [simulationMode, setSimulationMode] = useState(false);

  // Registration form state
  const [registerForm, setRegisterForm] = useState({
    classId: "",
    provisionType: AntiDilutionType.WeightedAverage,
    triggerThreshold: "0",
    adjustmentCap: "50",
    minimumPrice: "1.00",
  });

  // Simulation state
  const [simulation, setSimulation] = useState({
    currentShares: "100000",
    currentPrice: "10.00",
    newPrice: "7.50",
    newShares: "500000",
    totalOutstanding: "5000000",
  });

  const calculateSimulatedAdjustment = () => {
    const { currentShares, currentPrice, newPrice, newShares, totalOutstanding } = simulation;

    const fullRatchet = calculateFullRatchetAdjustment(currentShares, currentPrice, newPrice);
    const weightedAverage = calculateWeightedAverageAdjustment(
      currentShares, currentPrice, newPrice, newShares, totalOutstanding
    );

    const originalNum = parseFloat(currentShares);

    return {
      fullRatchet: {
        newShares: fullRatchet,
        sharesAdded: (parseFloat(fullRatchet) - originalNum).toFixed(0),
        percentIncrease: (((parseFloat(fullRatchet) - originalNum) / originalNum) * 100).toFixed(2),
      },
      weightedAverage: {
        newShares: weightedAverage,
        sharesAdded: (parseFloat(weightedAverage) - originalNum).toFixed(0),
        percentIncrease: (((parseFloat(weightedAverage) - originalNum) / originalNum) * 100).toFixed(2),
      },
    };
  };

  const tabs = [
    { id: "overview", label: "Overview", icon: "" },
    { id: "provisions", label: "Provisions", icon: "" },
    { id: "issuances", label: "Issuance History", icon: "" },
    { id: "adjustments", label: "Adjustments", icon: "" },
    ...(isFounder ? [{ id: "register", label: "Register Protection", icon: "" }] : []),
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold">Anti-Dilution Protection</h2>
          <p className="text-gray-600">{companyName} - Investor Protection Management</p>
        </div>
        <button
          onClick={() => setSimulationMode(!simulationMode)}
          className={`px-4 py-2 rounded font-medium ${
            simulationMode
              ? "bg-purple-600 text-white"
              : "border border-purple-600 text-purple-600"
          }`}
        >
          {simulationMode ? "Exit Simulation" : "Simulation Mode"}
        </button>
      </div>

      {/* Simulation Panel */}
      {simulationMode && (
        <div className="bg-purple-50 border border-purple-200 rounded-lg p-6">
          <h3 className="font-bold text-purple-800 mb-4">Down Round Simulation</h3>
          <p className="text-sm text-purple-600 mb-4">
            Calculate how anti-dilution protection would adjust shareholdings in a down round scenario.
          </p>

          <div className="grid md:grid-cols-5 gap-4 mb-6">
            <div>
              <label className="block text-sm font-medium text-purple-700 mb-1">Your Shares</label>
              <input
                type="number"
                value={simulation.currentShares}
                onChange={(e) => setSimulation({...simulation, currentShares: e.target.value})}
                className="w-full border border-purple-300 rounded p-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-purple-700 mb-1">Original Price</label>
              <input
                type="number"
                step="0.01"
                value={simulation.currentPrice}
                onChange={(e) => setSimulation({...simulation, currentPrice: e.target.value})}
                className="w-full border border-purple-300 rounded p-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-purple-700 mb-1">New Round Price</label>
              <input
                type="number"
                step="0.01"
                value={simulation.newPrice}
                onChange={(e) => setSimulation({...simulation, newPrice: e.target.value})}
                className="w-full border border-purple-300 rounded p-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-purple-700 mb-1">New Shares Issued</label>
              <input
                type="number"
                value={simulation.newShares}
                onChange={(e) => setSimulation({...simulation, newShares: e.target.value})}
                className="w-full border border-purple-300 rounded p-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-purple-700 mb-1">Total Outstanding</label>
              <input
                type="number"
                value={simulation.totalOutstanding}
                onChange={(e) => setSimulation({...simulation, totalOutstanding: e.target.value})}
                className="w-full border border-purple-300 rounded p-2"
              />
            </div>
          </div>

          {parseFloat(simulation.newPrice) < parseFloat(simulation.currentPrice) && (
            <div className="grid md:grid-cols-2 gap-4">
              {(() => {
                const results = calculateSimulatedAdjustment();
                return (
                  <>
                    <div className="bg-white border border-purple-300 rounded-lg p-4">
                      <h4 className="font-bold text-purple-800 mb-2">Full Ratchet Protection</h4>
                      <div className="space-y-2 text-sm">
                        <div className="flex justify-between">
                          <span>Original Shares:</span>
                          <span className="font-mono">{formatShares(simulation.currentShares)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Adjusted Shares:</span>
                          <span className="font-mono font-bold text-green-600">{formatShares(results.fullRatchet.newShares)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Shares Added:</span>
                          <span className="font-mono text-green-600">+{formatShares(results.fullRatchet.sharesAdded)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Increase:</span>
                          <span className="font-mono text-green-600">+{results.fullRatchet.percentIncrease}%</span>
                        </div>
                      </div>
                    </div>
                    <div className="bg-white border border-purple-300 rounded-lg p-4">
                      <h4 className="font-bold text-purple-800 mb-2">Weighted Average Protection</h4>
                      <div className="space-y-2 text-sm">
                        <div className="flex justify-between">
                          <span>Original Shares:</span>
                          <span className="font-mono">{formatShares(simulation.currentShares)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Adjusted Shares:</span>
                          <span className="font-mono font-bold text-blue-600">{formatShares(results.weightedAverage.newShares)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Shares Added:</span>
                          <span className="font-mono text-blue-600">+{formatShares(results.weightedAverage.sharesAdded)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Increase:</span>
                          <span className="font-mono text-blue-600">+{results.weightedAverage.percentIncrease}%</span>
                        </div>
                      </div>
                    </div>
                  </>
                );
              })()}
            </div>
          )}

          {parseFloat(simulation.newPrice) >= parseFloat(simulation.currentPrice) && (
            <div className="bg-green-100 border border-green-300 rounded p-4 text-center">
              <span className="text-green-800">
                Not a down round - no anti-dilution adjustment needed
              </span>
            </div>
          )}
        </div>
      )}

      {/* Tab Navigation */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-4 overflow-x-auto">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as typeof activeTab)}
              className={`py-2 px-3 border-b-2 font-medium text-sm whitespace-nowrap ${
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

      {/* Overview Tab */}
      {activeTab === "overview" && (
        <div className="space-y-6">
          {/* Summary Stats */}
          <div className="grid md:grid-cols-4 gap-4">
            <div className="border rounded-lg p-4 text-center">
              <div className="text-2xl font-bold text-blue-600">{mockProvisions.filter(p => p.isActive).length}</div>
              <p className="text-sm text-gray-600">Active Provisions</p>
            </div>
            <div className="border rounded-lg p-4 text-center">
              <div className="text-2xl font-bold text-green-600">{mockAdjustments.length}</div>
              <p className="text-sm text-gray-600">Total Adjustments</p>
            </div>
            <div className="border rounded-lg p-4 text-center">
              <div className="text-2xl font-bold text-purple-600">{mockIssuances.filter(i => i.isDownRound).length}</div>
              <p className="text-sm text-gray-600">Down Rounds</p>
            </div>
            <div className="border rounded-lg p-4 text-center">
              <div className="text-2xl font-bold text-orange-600">
                {formatShares(mockAdjustments.reduce((sum, a) => sum + parseFloat(a.sharesAdded), 0).toString())}
              </div>
              <p className="text-sm text-gray-600">Shares Adjusted</p>
            </div>
          </div>

          {/* What is Anti-Dilution */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
            <h3 className="font-bold text-blue-800 mb-3">What is Anti-Dilution Protection?</h3>
            <p className="text-blue-700 mb-4">
              Anti-dilution protection shields investors from ownership dilution when a company issues new shares
              at a lower price than previous rounds (down rounds). This protection automatically adjusts
              shareholdings to compensate for the price decrease.
            </p>
            <div className="grid md:grid-cols-3 gap-4 mt-4">
              {[AntiDilutionType.FullRatchet, AntiDilutionType.WeightedAverage, AntiDilutionType.BroadBasedWeightedAverage].map((type) => (
                <div key={type} className="bg-white rounded p-4">
                  <h4 className="font-semibold text-blue-800">{AntiDilutionTypeLabels[type]}</h4>
                  <p className="text-sm text-blue-600 mt-1">{AntiDilutionTypeDescriptions[type]}</p>
                </div>
              ))}
            </div>
          </div>

          {/* Protected Share Classes */}
          <div className="border rounded-lg p-6">
            <h3 className="font-bold mb-4">Share Classes Protection Status</h3>
            <div className="space-y-3">
              {shareClasses.map((sc) => (
                <div key={sc.classId} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                  <div>
                    <span className="font-semibold">{sc.name}</span>
                    <span className="text-gray-500 ml-2">({sc.classId})</span>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-sm text-gray-600">
                      {formatShares(sc.outstandingShares)} shares
                    </span>
                    {sc.hasAntiDilution ? (
                      <span className="px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm font-medium">
                        {AntiDilutionTypeLabels[sc.antiDilutionType]}
                      </span>
                    ) : (
                      <span className="px-3 py-1 bg-gray-100 text-gray-600 rounded-full text-sm">
                        No Protection
                      </span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Provisions Tab */}
      {activeTab === "provisions" && (
        <div className="space-y-4">
          <h3 className="font-bold text-lg">Registered Anti-Dilution Provisions</h3>
          {mockProvisions.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No anti-dilution provisions registered yet.
            </div>
          ) : (
            <div className="space-y-4">
              {mockProvisions.map((provision) => (
                <div key={`${provision.companyId}-${provision.classId}`} className="border rounded-lg p-6">
                  <div className="flex justify-between items-start mb-4">
                    <div>
                      <h4 className="font-bold text-lg">{provision.classId}</h4>
                      <span className={`px-2 py-1 rounded text-xs ${
                        provision.isActive
                          ? "bg-green-100 text-green-800"
                          : "bg-gray-100 text-gray-600"
                      }`}>
                        {provision.isActive ? "Active" : "Inactive"}
                      </span>
                    </div>
                    <span className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm font-medium">
                      {AntiDilutionTypeLabels[provision.provisionType]}
                    </span>
                  </div>

                  <div className="grid md:grid-cols-4 gap-4 text-sm">
                    <div>
                      <span className="text-gray-500">Initial Price</span>
                      <div className="font-mono font-semibold">{formatCurrency(provision.initialPrice)}</div>
                    </div>
                    <div>
                      <span className="text-gray-500">Current Adjusted Price</span>
                      <div className="font-mono font-semibold text-blue-600">{formatCurrency(provision.currentAdjustedPrice)}</div>
                    </div>
                    <div>
                      <span className="text-gray-500">Adjustment Cap</span>
                      <div className="font-semibold">{formatPercent(provision.adjustmentCap)}</div>
                    </div>
                    <div>
                      <span className="text-gray-500">Minimum Price Floor</span>
                      <div className="font-mono font-semibold">{formatCurrency(provision.minimumPrice)}</div>
                    </div>
                  </div>

                  <div className="mt-4 pt-4 border-t flex justify-between text-sm text-gray-500">
                    <span>Total Adjustments: {provision.totalAdjustments}</span>
                    <span>Last Adjustment: {formatDateTime(provision.lastAdjustmentTime)}</span>
                    <span>Created: {formatDateTime(provision.createdAt)}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Issuances Tab */}
      {activeTab === "issuances" && (
        <div className="space-y-4">
          <h3 className="font-bold text-lg">Share Issuance History</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-3 px-4">Date</th>
                  <th className="text-left py-3 px-4">Class</th>
                  <th className="text-right py-3 px-4">Shares Issued</th>
                  <th className="text-right py-3 px-4">Price</th>
                  <th className="text-right py-3 px-4">Total Raised</th>
                  <th className="text-left py-3 px-4">Purpose</th>
                  <th className="text-center py-3 px-4">Type</th>
                </tr>
              </thead>
              <tbody>
                {mockIssuances.map((issuance) => (
                  <tr key={issuance.id} className="border-b hover:bg-gray-50">
                    <td className="py-3 px-4">{formatDateTime(issuance.timestamp)}</td>
                    <td className="py-3 px-4 font-medium">{issuance.classId}</td>
                    <td className="py-3 px-4 text-right font-mono">{formatShares(issuance.sharesIssued)}</td>
                    <td className="py-3 px-4 text-right font-mono">{formatCurrency(issuance.issuePrice)}</td>
                    <td className="py-3 px-4 text-right font-mono">{formatCurrency(issuance.totalRaised)}</td>
                    <td className="py-3 px-4">{issuance.purpose}</td>
                    <td className="py-3 px-4 text-center">
                      {issuance.isDownRound ? (
                        <span className="px-2 py-1 bg-red-100 text-red-800 rounded text-xs font-medium">
                          Down Round (-{formatPercent(issuance.priceDropPercent)})
                        </span>
                      ) : (
                        <span className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs font-medium">
                          Normal
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Adjustments Tab */}
      {activeTab === "adjustments" && (
        <div className="space-y-4">
          <h3 className="font-bold text-lg">Anti-Dilution Adjustments</h3>
          {mockAdjustments.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No adjustments have been made yet.
            </div>
          ) : (
            <div className="space-y-4">
              {mockAdjustments.map((adjustment) => (
                <div key={adjustment.id} className="border rounded-lg p-6">
                  <div className="flex justify-between items-start mb-4">
                    <div>
                      <h4 className="font-mono text-sm text-gray-600">{adjustment.shareholder}</h4>
                      <span className="text-sm text-gray-500">{adjustment.classId}</span>
                    </div>
                    <span className="px-3 py-1 bg-purple-100 text-purple-800 rounded-full text-sm font-medium">
                      {AntiDilutionTypeLabels[adjustment.adjustmentType]}
                    </span>
                  </div>

                  <div className="grid md:grid-cols-5 gap-4 text-sm">
                    <div>
                      <span className="text-gray-500">Original Shares</span>
                      <div className="font-mono font-semibold">{formatShares(adjustment.originalShares)}</div>
                    </div>
                    <div>
                      <span className="text-gray-500">Adjusted Shares</span>
                      <div className="font-mono font-semibold text-green-600">{formatShares(adjustment.adjustedShares)}</div>
                    </div>
                    <div>
                      <span className="text-gray-500">Shares Added</span>
                      <div className="font-mono font-semibold text-green-600">+{formatShares(adjustment.sharesAdded)}</div>
                    </div>
                    <div>
                      <span className="text-gray-500">Trigger Price</span>
                      <div className="font-mono font-semibold text-red-600">{formatCurrency(adjustment.triggerPrice)}</div>
                    </div>
                    <div>
                      <span className="text-gray-500">Adjustment Ratio</span>
                      <div className="font-semibold">{adjustment.adjustmentRatio}x</div>
                    </div>
                  </div>

                  <div className="mt-4 pt-4 border-t text-sm text-gray-500">
                    Adjusted on {formatDateTime(adjustment.timestamp)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Register Tab (Founders Only) */}
      {activeTab === "register" && isFounder && (
        <div className="max-w-2xl">
          <h3 className="font-bold text-lg mb-4">Register Anti-Dilution Protection</h3>
          <p className="text-gray-600 mb-6">
            Set up anti-dilution protection for a share class. This will automatically adjust
            shareholdings when shares are issued at a lower price.
          </p>

          <form className="space-y-6">
            <div>
              <label className="block font-semibold mb-2">Share Class</label>
              <select
                className="w-full border rounded p-3"
                value={registerForm.classId}
                onChange={(e) => setRegisterForm({...registerForm, classId: e.target.value})}
              >
                <option value="">Select a share class...</option>
                {shareClasses.filter(sc => !sc.hasAntiDilution).map((sc) => (
                  <option key={sc.classId} value={sc.classId}>
                    {sc.name} ({sc.classId}) - {formatShares(sc.outstandingShares)} shares
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block font-semibold mb-2">Protection Type</label>
              <div className="space-y-3">
                {[AntiDilutionType.FullRatchet, AntiDilutionType.WeightedAverage, AntiDilutionType.BroadBasedWeightedAverage].map((type) => (
                  <label key={type} className="flex items-start gap-3 p-4 border rounded-lg cursor-pointer hover:bg-gray-50">
                    <input
                      type="radio"
                      name="provisionType"
                      value={type}
                      checked={registerForm.provisionType === type}
                      onChange={() => setRegisterForm({...registerForm, provisionType: type})}
                      className="mt-1"
                    />
                    <div>
                      <div className="font-medium">{AntiDilutionTypeLabels[type]}</div>
                      <div className="text-sm text-gray-600">{AntiDilutionTypeDescriptions[type]}</div>
                    </div>
                  </label>
                ))}
              </div>
            </div>

            <div className="grid md:grid-cols-3 gap-4">
              <div>
                <label className="block font-semibold mb-2">Trigger Threshold (%)</label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  max="100"
                  className="w-full border rounded p-3"
                  value={registerForm.triggerThreshold}
                  onChange={(e) => setRegisterForm({...registerForm, triggerThreshold: e.target.value})}
                  placeholder="0 = any decrease"
                />
                <p className="text-xs text-gray-500 mt-1">Minimum price drop to trigger (0 = any)</p>
              </div>
              <div>
                <label className="block font-semibold mb-2">Adjustment Cap (%)</label>
                <input
                  type="number"
                  step="1"
                  min="1"
                  max="100"
                  className="w-full border rounded p-3"
                  value={registerForm.adjustmentCap}
                  onChange={(e) => setRegisterForm({...registerForm, adjustmentCap: e.target.value})}
                  placeholder="50"
                />
                <p className="text-xs text-gray-500 mt-1">Max adjustment per round</p>
              </div>
              <div>
                <label className="block font-semibold mb-2">Minimum Price Floor</label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  className="w-full border rounded p-3"
                  value={registerForm.minimumPrice}
                  onChange={(e) => setRegisterForm({...registerForm, minimumPrice: e.target.value})}
                  placeholder="1.00"
                />
                <p className="text-xs text-gray-500 mt-1">No adjustments below this price</p>
              </div>
            </div>

            <div className="bg-yellow-50 border border-yellow-200 rounded p-4">
              <h4 className="font-semibold text-yellow-800 mb-2">Important Notice</h4>
              <p className="text-sm text-yellow-700">
                Anti-dilution protection is recorded on-chain and cannot be removed once registered.
                Make sure you understand the implications before proceeding. This protection applies
                to all current and future shareholders of the selected share class.
              </p>
            </div>

            <button
              type="submit"
              disabled={!registerForm.classId}
              className="w-full bg-blue-500 text-white py-3 px-6 rounded font-semibold disabled:bg-gray-300 disabled:cursor-not-allowed hover:bg-blue-600"
            >
              Register Anti-Dilution Protection
            </button>
          </form>
        </div>
      )}
    </div>
  );
}
