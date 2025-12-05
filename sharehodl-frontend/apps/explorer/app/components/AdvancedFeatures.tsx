"use client";

export default function AdvancedFeatures() {
  const governanceStats = {
    activeProposals: 15,
    emergencyProposals: 1,
    totalVoters: 3247,
    delegatedPower: "12.3M HODL",
    passRate: 89
  };

  const tradingStats = {
    dailyVolume: "45.2M",
    activeTradingPairs: 12,
    circuitBreakers: "NORMAL",
    averageSettlement: "6.2s",
    institutionalOrders: 2847
  };

  const validatorStats = {
    totalValidators: 127,
    goldTier: 15,
    silverTier: 34,
    bronzeTier: 78,
    slashingEvents: 0
  };

  const companyStats = {
    listedCompanies: 8,
    pendingIPOs: 3,
    totalMarketCap: "2.1B",
    avgTradingVolume: "156M",
    shareClassTypes: 24
  };

  return (
    <div className="space-y-6">
      {/* Hero Banner */}
      <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white p-6 rounded-lg">
        <h2 className="text-2xl font-bold mb-2">ShareHODL: Institutional-Grade Blockchain Infrastructure</h2>
        <p className="text-blue-100">Professional equity trading with advanced governance, emergency protocols, and institutional safeguards</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* Governance Features */}
        <div className="bg-white border rounded-lg p-6">
          <h3 className="font-bold text-lg mb-4 flex items-center gap-2">
            Advanced Governance
          </h3>
          <div className="space-y-3 text-sm">
            <div className="flex justify-between">
              <span>Active Proposals:</span>
              <span className="font-semibold">{governanceStats.activeProposals}</span>
            </div>
            <div className="flex justify-between">
              <span>Emergency Proposals:</span>
              <span className="font-semibold text-red-600">{governanceStats.emergencyProposals}</span>
            </div>
            <div className="flex justify-between">
              <span>Total Voters:</span>
              <span className="font-semibold">{governanceStats.totalVoters.toLocaleString()}</span>
            </div>
            <div className="flex justify-between">
              <span>Delegated Power:</span>
              <span className="font-semibold">{governanceStats.delegatedPower}</span>
            </div>
            <div className="flex justify-between">
              <span>Pass Rate:</span>
              <span className="font-semibold text-green-600">{governanceStats.passRate}%</span>
            </div>
          </div>
          
          <div className="mt-4 p-3 bg-blue-50 border border-blue-200 rounded">
            <div className="text-xs text-blue-800">
              <strong>Features:</strong> Multi-tier proposals, emergency response, voting delegation, company governance, anti-takeover protections
            </div>
          </div>
        </div>

        {/* Trading Features */}
        <div className="bg-white border rounded-lg p-6">
          <h3 className="font-bold text-lg mb-4 flex items-center gap-2">
            Professional Trading
          </h3>
          <div className="space-y-3 text-sm">
            <div className="flex justify-between">
              <span>Daily Volume:</span>
              <span className="font-semibold">${tradingStats.dailyVolume}</span>
            </div>
            <div className="flex justify-between">
              <span>Active Pairs:</span>
              <span className="font-semibold">{tradingStats.activeTradingPairs}</span>
            </div>
            <div className="flex justify-between">
              <span>Circuit Breakers:</span>
              <span className="font-semibold text-green-600">{tradingStats.circuitBreakers}</span>
            </div>
            <div className="flex justify-between">
              <span>Avg Settlement:</span>
              <span className="font-semibold text-blue-600">{tradingStats.averageSettlement}</span>
            </div>
            <div className="flex justify-between">
              <span>Pro Orders:</span>
              <span className="font-semibold">{tradingStats.institutionalOrders.toLocaleString()}</span>
            </div>
          </div>
          
          <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded">
            <div className="text-xs text-green-800">
              <strong>Features:</strong> FOK/IOC orders, circuit breakers, T+0 settlement, atomic swaps, ownership limits
            </div>
          </div>
        </div>

        {/* Validator System */}
        <div className="bg-white border rounded-lg p-6">
          <h3 className="font-bold text-lg mb-4 flex items-center gap-2">
            Validator Network
          </h3>
          <div className="space-y-3 text-sm">
            <div className="flex justify-between">
              <span>Total Validators:</span>
              <span className="font-semibold">{validatorStats.totalValidators}</span>
            </div>
            <div className="flex justify-between">
              <span>Gold Tier:</span>
              <span className="font-semibold text-yellow-600">{validatorStats.goldTier}</span>
            </div>
            <div className="flex justify-between">
              <span>Silver Tier:</span>
              <span className="font-semibold text-gray-500">{validatorStats.silverTier}</span>
            </div>
            <div className="flex justify-between">
              <span>Bronze Tier:</span>
              <span className="font-semibold text-orange-600">{validatorStats.bronzeTier}</span>
            </div>
            <div className="flex justify-between">
              <span>Slashing Events:</span>
              <span className="font-semibold text-green-600">{validatorStats.slashingEvents}</span>
            </div>
          </div>
          
          <div className="mt-4 p-3 bg-purple-50 border border-purple-200 rounded">
            <div className="text-xs text-purple-800">
              <strong>Features:</strong> Dual-role validation, tiered system, business verification, governance participation
            </div>
          </div>
        </div>

        {/* Company Ecosystem */}
        <div className="bg-white border rounded-lg p-6">
          <h3 className="font-bold text-lg mb-4 flex items-center gap-2">
            Company Ecosystem
          </h3>
          <div className="space-y-3 text-sm">
            <div className="flex justify-between">
              <span>Listed Companies:</span>
              <span className="font-semibold">{companyStats.listedCompanies}</span>
            </div>
            <div className="flex justify-between">
              <span>Pending IPOs:</span>
              <span className="font-semibold text-orange-600">{companyStats.pendingIPOs}</span>
            </div>
            <div className="flex justify-between">
              <span>Market Cap:</span>
              <span className="font-semibold">${companyStats.totalMarketCap}</span>
            </div>
            <div className="flex justify-between">
              <span>Daily Volume:</span>
              <span className="font-semibold">${companyStats.avgTradingVolume}</span>
            </div>
            <div className="flex justify-between">
              <span>Share Classes:</span>
              <span className="font-semibold">{companyStats.shareClassTypes}</span>
            </div>
          </div>
          
          <div className="mt-4 p-3 bg-orange-50 border border-orange-200 rounded">
            <div className="text-xs text-orange-800">
              <strong>Features:</strong> ERC-equity tokens, automated dividends, corporate governance, IPO streamlining
            </div>
          </div>
        </div>
      </div>

      {/* Feature Comparison */}
      <div className="bg-white border rounded-lg p-6">
        <h3 className="font-bold text-lg mb-4">ShareHODL vs Traditional Stock Markets</h3>
        
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b">
                <th className="text-left py-3 px-4">Feature</th>
                <th className="text-center py-3 px-4 text-green-600">ShareHODL</th>
                <th className="text-center py-3 px-4 text-red-600">Traditional</th>
                <th className="text-center py-3 px-4">Advantage</th>
              </tr>
            </thead>
            <tbody className="text-sm">
              <tr className="border-b">
                <td className="py-3 px-4 font-medium">Settlement Time</td>
                <td className="text-center py-3 px-4 text-green-600 font-semibold">6 seconds</td>
                <td className="text-center py-3 px-4 text-red-600">T+2 (48-72 hours)</td>
                <td className="text-center py-3 px-4 text-green-600 font-bold">28,800x faster</td>
              </tr>
              <tr className="border-b">
                <td className="py-3 px-4 font-medium">Trading Hours</td>
                <td className="text-center py-3 px-4 text-green-600 font-semibold">24/7/365</td>
                <td className="text-center py-3 px-4 text-red-600">8hr/day, 5days/week</td>
                <td className="text-center py-3 px-4 text-green-600 font-bold">7.5x more access</td>
              </tr>
              <tr className="border-b">
                <td className="py-3 px-4 font-medium">Trading Fees</td>
                <td className="text-center py-3 px-4 text-green-600 font-semibold">$0.005</td>
                <td className="text-center py-3 px-4 text-red-600">$5-15+</td>
                <td className="text-center py-3 px-4 text-green-600 font-bold">1000-3000x cheaper</td>
              </tr>
              <tr className="border-b">
                <td className="py-3 px-4 font-medium">Minimum Investment</td>
                <td className="text-center py-3 px-4 text-green-600 font-semibold">$1 (fractional)</td>
                <td className="text-center py-3 px-4 text-red-600">Full share price</td>
                <td className="text-center py-3 px-4 text-green-600 font-bold">100x more accessible</td>
              </tr>
              <tr className="border-b">
                <td className="py-3 px-4 font-medium">Ownership</td>
                <td className="text-center py-3 px-4 text-green-600 font-semibold">Direct custody</td>
                <td className="text-center py-3 px-4 text-red-600">Beneficial (broker held)</td>
                <td className="text-center py-3 px-4 text-green-600 font-bold">True ownership</td>
              </tr>
              <tr>
                <td className="py-3 px-4 font-medium">Governance</td>
                <td className="text-center py-3 px-4 text-green-600 font-semibold">On-chain voting</td>
                <td className="text-center py-3 px-4 text-red-600">Proxy voting</td>
                <td className="text-center py-3 px-4 text-green-600 font-bold">Direct participation</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Emergency Status */}
      <div className="bg-red-50 border-l-4 border-red-500 p-4 rounded">
        <div className="flex items-center">
          <span className="text-red-500 mr-2"></span>
          <div>
            <h4 className="font-semibold text-red-800">Emergency Governance Active</h4>
            <p className="text-red-600">Critical security proposal requires validator attention - 47 minutes remaining for Severity 5 emergency</p>
          </div>
        </div>
      </div>
    </div>
  );
}