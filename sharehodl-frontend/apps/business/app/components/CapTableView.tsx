"use client";

import { useState } from "react";
import {
  CompanyCapTable,
  ShareClassSummary,
  ShareholderSummary,
  AntiDilutionType,
  AntiDilutionTypeLabels,
  formatShares,
  formatCurrency,
  formatPercent,
} from "../types/equity";
import AntiDilutionManager from "./AntiDilutionManager";

interface CapTableViewProps {
  companyId: string;
  companyName: string;
  companySymbol: string;
  isFounder?: boolean;
}

// Mock data - in production this would come from blockchain queries
const mockCapTable: CompanyCapTable = {
  company: {
    id: "1",
    name: "TechStart AI",
    symbol: "TSAI",
    description: "Leading AI solutions provider",
    founder: "sharehodl1founder...xyz",
    website: "https://techstart.ai",
    industry: "Technology",
    country: "USA",
    status: "Active",
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-06-15T10:30:00Z",
  },
  shareClasses: [
    {
      classId: "COMMON",
      name: "Common Stock",
      authorizedShares: "100000000",
      issuedShares: "45000000",
      outstandingShares: "45000000",
      holderCount: 1250,
      hasAntiDilution: false,
      antiDilutionType: AntiDilutionType.None,
      votingRights: true,
      dividendRights: true,
    },
    {
      classId: "SERIES_A",
      name: "Series A Preferred",
      authorizedShares: "20000000",
      issuedShares: "15000000",
      outstandingShares: "15000000",
      holderCount: 85,
      hasAntiDilution: true,
      antiDilutionType: AntiDilutionType.FullRatchet,
      votingRights: true,
      dividendRights: true,
    },
    {
      classId: "SERIES_B",
      name: "Series B Preferred",
      authorizedShares: "25000000",
      issuedShares: "10000000",
      outstandingShares: "10000000",
      holderCount: 42,
      hasAntiDilution: true,
      antiDilutionType: AntiDilutionType.WeightedAverage,
      votingRights: true,
      dividendRights: true,
    },
  ],
  topHolders: [
    { owner: "sharehodl1founder...xyz", totalShares: "25000000", ownershipPercent: "35.71", classId: "COMMON", votingPower: "35.71" },
    { owner: "sharehodl1vcfund1...abc", totalShares: "10000000", ownershipPercent: "14.29", classId: "SERIES_A", votingPower: "14.29" },
    { owner: "sharehodl1vcfund2...def", totalShares: "8000000", ownershipPercent: "11.43", classId: "SERIES_B", votingPower: "11.43" },
    { owner: "sharehodl1angel1...ghi", totalShares: "5000000", ownershipPercent: "7.14", classId: "SERIES_A", votingPower: "7.14" },
    { owner: "sharehodl1esop...jkl", totalShares: "4000000", ownershipPercent: "5.71", classId: "COMMON", votingPower: "5.71" },
  ],
  issuanceStats: {
    totalIssuances: 5,
    totalSharesIssued: "70000000",
    totalRaised: "35000000",
    averageIssuePrice: "0.50",
    lastIssuancePrice: "0.35",
    downRoundCount: 1,
  },
  antiDilution: {
    protectedClasses: 2,
    totalAdjustments: 15,
    totalSharesAdjusted: "2500000",
    activeProvisions: 2,
  },
};

export default function CapTableView({
  companyId,
  companyName,
  companySymbol,
  isFounder = false,
}: CapTableViewProps) {
  const [activeView, setActiveView] = useState<"summary" | "shareholders" | "classes" | "anti-dilution">("summary");
  const capTable = mockCapTable; // In production: useCapTableQuery(companyId)

  const totalOutstanding = capTable.shareClasses.reduce(
    (sum, sc) => sum + parseFloat(sc.outstandingShares),
    0
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold flex items-center gap-2">
            {companyName}
            <span className="text-gray-500 font-normal">({companySymbol})</span>
          </h2>
          <p className="text-gray-600">Cap Table & Equity Management</p>
        </div>
        <div className="text-right">
          <div className="text-sm text-gray-500">Total Outstanding</div>
          <div className="text-2xl font-bold">{formatShares(totalOutstanding.toString())}</div>
        </div>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        <div className="border rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-blue-600">{capTable.shareClasses.length}</div>
          <p className="text-sm text-gray-600">Share Classes</p>
        </div>
        <div className="border rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-green-600">
            {capTable.shareClasses.reduce((sum, sc) => sum + sc.holderCount, 0).toLocaleString()}
          </div>
          <p className="text-sm text-gray-600">Total Shareholders</p>
        </div>
        <div className="border rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-purple-600">{formatCurrency(capTable.issuanceStats.totalRaised)}</div>
          <p className="text-sm text-gray-600">Total Raised</p>
        </div>
        <div className="border rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-orange-600">{capTable.antiDilution.protectedClasses}</div>
          <p className="text-sm text-gray-600">Protected Classes</p>
        </div>
        <div className="border rounded-lg p-4 text-center">
          <div className={`text-2xl font-bold ${capTable.issuanceStats.downRoundCount > 0 ? 'text-red-600' : 'text-green-600'}`}>
            {capTable.issuanceStats.downRoundCount}
          </div>
          <p className="text-sm text-gray-600">Down Rounds</p>
        </div>
      </div>

      {/* View Navigation */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-4 overflow-x-auto">
          {[
            { id: "summary", label: "Summary", icon: "" },
            { id: "shareholders", label: "Top Shareholders", icon: "" },
            { id: "classes", label: "Share Classes", icon: "" },
            { id: "anti-dilution", label: "Anti-Dilution", icon: "" },
          ].map((view) => (
            <button
              key={view.id}
              onClick={() => setActiveView(view.id as typeof activeView)}
              className={`py-2 px-3 border-b-2 font-medium text-sm whitespace-nowrap ${
                activeView === view.id
                  ? "border-blue-500 text-blue-600"
                  : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
              }`}
            >
              <span className="mr-2">{view.icon}</span>
              {view.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Summary View */}
      {activeView === "summary" && (
        <div className="space-y-6">
          {/* Ownership Chart Placeholder */}
          <div className="border rounded-lg p-6">
            <h3 className="font-bold text-lg mb-4">Ownership Distribution</h3>
            <div className="flex items-center gap-8">
              {/* Simple bar chart */}
              <div className="flex-1 space-y-3">
                {capTable.shareClasses.map((sc) => {
                  const percent = (parseFloat(sc.outstandingShares) / totalOutstanding) * 100;
                  return (
                    <div key={sc.classId}>
                      <div className="flex justify-between text-sm mb-1">
                        <span>{sc.name}</span>
                        <span className="font-mono">{percent.toFixed(1)}%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-4">
                        <div
                          className={`h-4 rounded-full ${
                            sc.classId === "COMMON" ? "bg-blue-500" :
                            sc.classId === "SERIES_A" ? "bg-green-500" :
                            sc.classId === "SERIES_B" ? "bg-purple-500" :
                            "bg-orange-500"
                          }`}
                          style={{ width: `${percent}%` }}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
              {/* Legend */}
              <div className="space-y-2">
                {capTable.shareClasses.map((sc) => (
                  <div key={sc.classId} className="flex items-center gap-2 text-sm">
                    <div className={`w-3 h-3 rounded ${
                      sc.classId === "COMMON" ? "bg-blue-500" :
                      sc.classId === "SERIES_A" ? "bg-green-500" :
                      sc.classId === "SERIES_B" ? "bg-purple-500" :
                      "bg-orange-500"
                    }`} />
                    <span>{sc.name}</span>
                    {sc.hasAntiDilution && (
                      <span className="text-xs bg-yellow-100 text-yellow-800 px-1 rounded">Protected</span>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Issuance Statistics */}
          <div className="grid md:grid-cols-2 gap-6">
            <div className="border rounded-lg p-6">
              <h3 className="font-bold text-lg mb-4">Issuance Statistics</h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-gray-600">Total Issuances</span>
                  <span className="font-semibold">{capTable.issuanceStats.totalIssuances}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Total Shares Issued</span>
                  <span className="font-mono font-semibold">{formatShares(capTable.issuanceStats.totalSharesIssued)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Total Capital Raised</span>
                  <span className="font-mono font-semibold">{formatCurrency(capTable.issuanceStats.totalRaised)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Average Issue Price</span>
                  <span className="font-mono font-semibold">{formatCurrency(capTable.issuanceStats.averageIssuePrice)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Last Issue Price</span>
                  <span className={`font-mono font-semibold ${
                    parseFloat(capTable.issuanceStats.lastIssuancePrice) < parseFloat(capTable.issuanceStats.averageIssuePrice)
                      ? 'text-red-600'
                      : 'text-green-600'
                  }`}>
                    {formatCurrency(capTable.issuanceStats.lastIssuancePrice)}
                  </span>
                </div>
              </div>
            </div>

            <div className="border rounded-lg p-6">
              <h3 className="font-bold text-lg mb-4">Anti-Dilution Summary</h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-gray-600">Protected Share Classes</span>
                  <span className="font-semibold">{capTable.antiDilution.protectedClasses}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Active Provisions</span>
                  <span className="font-semibold">{capTable.antiDilution.activeProvisions}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Total Adjustments Made</span>
                  <span className="font-semibold">{capTable.antiDilution.totalAdjustments}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Shares Adjusted</span>
                  <span className="font-mono font-semibold text-green-600">+{formatShares(capTable.antiDilution.totalSharesAdjusted)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Down Rounds Recorded</span>
                  <span className={`font-semibold ${capTable.issuanceStats.downRoundCount > 0 ? 'text-red-600' : ''}`}>
                    {capTable.issuanceStats.downRoundCount}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Shareholders View */}
      {activeView === "shareholders" && (
        <div className="space-y-4">
          <h3 className="font-bold text-lg">Top Shareholders</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-3 px-4">Rank</th>
                  <th className="text-left py-3 px-4">Address</th>
                  <th className="text-left py-3 px-4">Share Class</th>
                  <th className="text-right py-3 px-4">Total Shares</th>
                  <th className="text-right py-3 px-4">Ownership %</th>
                  <th className="text-right py-3 px-4">Voting Power</th>
                </tr>
              </thead>
              <tbody>
                {capTable.topHolders.map((holder, index) => (
                  <tr key={holder.owner} className="border-b hover:bg-gray-50">
                    <td className="py-3 px-4 font-semibold">#{index + 1}</td>
                    <td className="py-3 px-4">
                      <span className="font-mono text-sm">{holder.owner}</span>
                      {holder.owner.includes("founder") && (
                        <span className="ml-2 text-xs bg-blue-100 text-blue-800 px-2 py-0.5 rounded">Founder</span>
                      )}
                      {holder.owner.includes("vcfund") && (
                        <span className="ml-2 text-xs bg-purple-100 text-purple-800 px-2 py-0.5 rounded">VC Fund</span>
                      )}
                      {holder.owner.includes("esop") && (
                        <span className="ml-2 text-xs bg-green-100 text-green-800 px-2 py-0.5 rounded">ESOP</span>
                      )}
                    </td>
                    <td className="py-3 px-4">{holder.classId}</td>
                    <td className="py-3 px-4 text-right font-mono">{formatShares(holder.totalShares)}</td>
                    <td className="py-3 px-4 text-right font-mono font-semibold">{formatPercent(holder.ownershipPercent)}</td>
                    <td className="py-3 px-4 text-right font-mono">{formatPercent(holder.votingPower)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="text-center text-gray-500 text-sm">
            Showing top 5 shareholders. Connect wallet to see full cap table.
          </div>
        </div>
      )}

      {/* Share Classes View */}
      {activeView === "classes" && (
        <div className="space-y-4">
          <h3 className="font-bold text-lg">Share Classes</h3>
          <div className="grid gap-4">
            {capTable.shareClasses.map((sc) => (
              <div key={sc.classId} className="border rounded-lg p-6">
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h4 className="font-bold text-lg">{sc.name}</h4>
                    <span className="text-gray-500 font-mono">{sc.classId}</span>
                  </div>
                  <div className="flex gap-2">
                    {sc.votingRights && (
                      <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs">Voting</span>
                    )}
                    {sc.dividendRights && (
                      <span className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs">Dividends</span>
                    )}
                    {sc.hasAntiDilution && (
                      <span className="px-2 py-1 bg-yellow-100 text-yellow-800 rounded text-xs">
                        {AntiDilutionTypeLabels[sc.antiDilutionType]}
                      </span>
                    )}
                  </div>
                </div>

                <div className="grid md:grid-cols-4 gap-4 text-sm">
                  <div>
                    <span className="text-gray-500">Authorized</span>
                    <div className="font-mono font-semibold">{formatShares(sc.authorizedShares)}</div>
                  </div>
                  <div>
                    <span className="text-gray-500">Issued</span>
                    <div className="font-mono font-semibold">{formatShares(sc.issuedShares)}</div>
                  </div>
                  <div>
                    <span className="text-gray-500">Outstanding</span>
                    <div className="font-mono font-semibold">{formatShares(sc.outstandingShares)}</div>
                  </div>
                  <div>
                    <span className="text-gray-500">Holders</span>
                    <div className="font-semibold">{sc.holderCount.toLocaleString()}</div>
                  </div>
                </div>

                {/* Utilization bar */}
                <div className="mt-4">
                  <div className="flex justify-between text-xs text-gray-500 mb-1">
                    <span>Utilization</span>
                    <span>{((parseFloat(sc.issuedShares) / parseFloat(sc.authorizedShares)) * 100).toFixed(1)}%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-blue-500 h-2 rounded-full"
                      style={{ width: `${(parseFloat(sc.issuedShares) / parseFloat(sc.authorizedShares)) * 100}%` }}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Anti-Dilution View */}
      {activeView === "anti-dilution" && (
        <AntiDilutionManager
          companyId={companyId}
          companyName={companyName}
          shareClasses={capTable.shareClasses.map((sc) => ({
            classId: sc.classId,
            name: sc.name,
            outstandingShares: sc.outstandingShares,
            hasAntiDilution: sc.hasAntiDilution,
            antiDilutionType: sc.antiDilutionType,
          }))}
          isFounder={isFounder}
        />
      )}
    </div>
  );
}
