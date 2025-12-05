"use client";

import { useState } from "react";

export default function Home() {
  const [activeTab, setActiveTab] = useState("overview");
  const [selectedService, setSelectedService] = useState("");

  const listedCompanies = [
    { symbol: "AAPL", name: "Apple Inc.", marketCap: "2.8T", shares: "15.7B", price: "$185.25", status: "Active" },
    { symbol: "TSLA", name: "Tesla Inc.", marketCap: "780B", shares: "3.2B", price: "$245.80", status: "Active" },
    { symbol: "GOOGL", name: "Alphabet Inc.", marketCap: "1.7T", shares: "620M", price: "$2,750.00", status: "Active" },
    { symbol: "MSFT", name: "Microsoft Corp.", marketCap: "2.9T", shares: "7.4B", price: "$385.60", status: "Active" }
  ];

  const pendingIPOs = [
    { company: "TechStart AI", sector: "Artificial Intelligence", valuation: "$2.1B", shares: "100M", status: "Due Diligence", validator: "Gold Tier" },
    { company: "GreenEnergy Solutions", sector: "Renewable Energy", valuation: "$890M", shares: "75M", status: "Documentation Review", validator: "Silver Tier" },
    { company: "BioTech Innovations", sector: "Biotechnology", valuation: "$1.5B", shares: "50M", status: "Final Approval", validator: "Platinum Tier" }
  ];

  const validators = [
    { name: "Institutional Capital Validators", tier: "Platinum", companies: 25, success: 98, specialization: "Fortune 500", stake: "500K HODL" },
    { name: "TechStart Verification Hub", tier: "Gold", companies: 18, success: 95, specialization: "Technology Startups", stake: "250K HODL" },
    { name: "Green Finance Validators", tier: "Gold", companies: 12, success: 97, specialization: "ESG & Sustainability", stake: "250K HODL" },
    { name: "Healthcare Compliance Group", tier: "Silver", companies: 8, success: 94, specialization: "Healthcare & Biotech", stake: "100K HODL" }
  ];

  const businessServices = [
    {
      id: "ipo",
      title: "Initial Public Offering (IPO)",
      description: "List your company on ShareHODL blockchain for global trading",
      features: ["Streamlined process: 2-6 weeks vs 12-18 months", "Lower costs: $1K-25K vs $10M+ traditional", "24/7 global trading access", "Instant settlement (T+0)", "Fractional share trading"],
      timeline: "2-6 weeks",
      cost: "$1,000 - $25,000",
      icon: "🏢"
    },
    {
      id: "validator",
      title: "Become a Business Validator",
      description: "Verify other businesses and earn rewards through our dual-role system",
      features: ["Earn validation rewards", "Governance voting rights", "Tier-based privileges", "Network reputation building", "Professional networking"],
      timeline: "1-2 weeks",
      cost: "50,000 - 500,000 HODL stake (Mainnet) / 1,000 - 30,000 HODL (Testnet)",
      icon: "🛡️"
    },
    {
      id: "integration",
      title: "Enterprise Integration",
      description: "Integrate ShareHODL blockchain into your existing business operations",
      features: ["API integration", "Custom smart contracts", "Automated dividend distribution", "Real-time analytics", "Compliance reporting"],
      timeline: "2-4 weeks",
      cost: "Custom pricing",
      icon: "⚙️"
    },
    {
      id: "governance",
      title: "Corporate Governance Tools",
      description: "Advanced on-chain governance for shareholder democracy",
      features: ["Board elections", "Proposal voting", "Shareholder communications", "Proxy voting", "Compliance automation"],
      timeline: "1-2 weeks",
      cost: "$500 - $5,000/month",
      icon: "🗳️"
    }
  ];

  return (
    <div className="min-h-screen bg-background">
      <main className="container mx-auto px-4 py-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold mb-4 flex items-center justify-center gap-3">
            <span className="text-3xl">🏢</span>
            ShareHODL Business Portal
          </h1>
          <p className="text-muted-foreground text-lg max-w-3xl mx-auto">
            Revolutionary blockchain infrastructure for business listings, validator services, and enterprise integration. 
            Join the future of equity markets.
          </p>
        </div>

        {/* Quick Stats */}
        <div className="grid gap-4 md:grid-cols-5 mb-8">
          <div className="border rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-blue-600">{listedCompanies.length}</div>
            <p className="text-sm text-muted-foreground">Listed Companies</p>
          </div>
          <div className="border rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-orange-600">{pendingIPOs.length}</div>
            <p className="text-sm text-muted-foreground">Pending IPOs</p>
          </div>
          <div className="border rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-purple-600">{validators.length}</div>
            <p className="text-sm text-muted-foreground">Active Validators</p>
          </div>
          <div className="border rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-green-600">$8.3T</div>
            <p className="text-sm text-muted-foreground">Total Market Cap</p>
          </div>
          <div className="border rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-red-600">99.97%</div>
            <p className="text-sm text-muted-foreground">Platform Uptime</p>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="border-b border-gray-200 mb-6">
          <nav className="-mb-px flex space-x-8">
            {[
              { id: "overview", label: "Platform Overview", icon: "📊" },
              { id: "services", label: "Business Services", icon: "🛠️" },
              { id: "ipo", label: "IPO Application", icon: "🏢" },
              { id: "validator", label: "Validator Registration", icon: "🛡️" },
              { id: "companies", label: "Listed Companies", icon: "📈" }
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
        {activeTab === "overview" && (
          <div className="space-y-8">
            {/* Hero Section */}
            <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white p-8 rounded-lg">
              <h2 className="text-3xl font-bold mb-4">Welcome to the Future of Business Finance</h2>
              <div className="grid md:grid-cols-3 gap-6">
                <div>
                  <h3 className="font-semibold mb-2 flex items-center gap-2">
                    ⚡ Instant Settlement
                  </h3>
                  <p className="text-blue-100 text-sm">
                    T+0 settlement in 6 seconds vs traditional T+2 (48-72 hours)
                  </p>
                </div>
                <div>
                  <h3 className="font-semibold mb-2 flex items-center gap-2">
                    🌍 24/7 Global Trading
                  </h3>
                  <p className="text-blue-100 text-sm">
                    Never-closing markets with 99.9% uptime vs 13% traditional market hours
                  </p>
                </div>
                <div>
                  <h3 className="font-semibold mb-2 flex items-center gap-2">
                    💰 Ultra-Low Costs
                  </h3>
                  <p className="text-blue-100 text-sm">
                    $0.005 trading fees vs $5-15+ traditional broker fees
                  </p>
                </div>
              </div>
            </div>

            {/* Platform Advantages */}
            <div className="grid md:grid-cols-2 gap-6">
              <div className="border rounded-lg p-6">
                <h3 className="font-bold text-lg mb-4 flex items-center gap-2">
                  🚀 For Companies Going Public
                </h3>
                <ul className="space-y-2 text-sm">
                  <li className="flex items-center gap-2">✅ <span>Reduced IPO timeline: 2-6 weeks vs 12-18 months</span></li>
                  <li className="flex items-center gap-2">✅ <span>Lower costs: $1K-25K vs $10M+ traditional</span></li>
                  <li className="flex items-center gap-2">✅ <span>Global investor access 24/7</span></li>
                  <li className="flex items-center gap-2">✅ <span>Fractional shares for broader accessibility</span></li>
                  <li className="flex items-center gap-2">✅ <span>Automated compliance and governance</span></li>
                </ul>
              </div>

              <div className="border rounded-lg p-6">
                <h3 className="font-bold text-lg mb-4 flex items-center gap-2">
                  🛡️ For Business Validators
                </h3>
                <ul className="space-y-2 text-sm">
                  <li className="flex items-center gap-2">✅ <span>Earn rewards for business verification</span></li>
                  <li className="flex items-center gap-2">✅ <span>Governance voting rights by tier</span></li>
                  <li className="flex items-center gap-2">✅ <span>Build reputation and network</span></li>
                  <li className="flex items-center gap-2">✅ <span>Dual-role validation system</span></li>
                  <li className="flex items-center gap-2">✅ <span>Professional business networking</span></li>
                </ul>
              </div>
            </div>

            {/* Process Comparison */}
            <div className="border rounded-lg p-6">
              <h3 className="font-bold text-lg mb-4">IPO Process Comparison</h3>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left py-2">Stage</th>
                      <th className="text-center py-2 text-green-600">ShareHODL</th>
                      <th className="text-center py-2 text-red-600">Traditional</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-b">
                      <td className="py-2 font-medium">Initial Application</td>
                      <td className="text-center py-2 text-green-600">1-2 days</td>
                      <td className="text-center py-2 text-red-600">2-4 weeks</td>
                    </tr>
                    <tr className="border-b">
                      <td className="py-2 font-medium">Due Diligence</td>
                      <td className="text-center py-2 text-green-600">1-2 weeks</td>
                      <td className="text-center py-2 text-red-600">3-6 months</td>
                    </tr>
                    <tr className="border-b">
                      <td className="py-2 font-medium">Documentation</td>
                      <td className="text-center py-2 text-green-600">3-5 days</td>
                      <td className="text-center py-2 text-red-600">2-4 months</td>
                    </tr>
                    <tr className="border-b">
                      <td className="py-2 font-medium">Regulatory Review</td>
                      <td className="text-center py-2 text-green-600">1-2 weeks</td>
                      <td className="text-center py-2 text-red-600">4-8 months</td>
                    </tr>
                    <tr>
                      <td className="py-2 font-medium font-bold">Total Timeline</td>
                      <td className="text-center py-2 text-green-600 font-bold">2-6 weeks</td>
                      <td className="text-center py-2 text-red-600 font-bold">12-18 months</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}

        {activeTab === "services" && (
          <div className="grid gap-6 md:grid-cols-2">
            {businessServices.map((service) => (
              <div key={service.id} className="border rounded-lg p-6 hover:shadow-lg transition-shadow">
                <div className="flex items-start gap-4">
                  <div className="text-4xl">{service.icon}</div>
                  <div className="flex-1">
                    <h3 className="font-bold text-lg mb-2">{service.title}</h3>
                    <p className="text-gray-600 mb-4">{service.description}</p>
                    
                    <div className="space-y-2 mb-4">
                      {service.features.map((feature, i) => (
                        <div key={i} className="flex items-center gap-2 text-sm">
                          <span className="text-green-500">✓</span>
                          <span>{feature}</span>
                        </div>
                      ))}
                    </div>
                    
                    <div className="grid grid-cols-2 gap-4 text-sm mb-4">
                      <div>
                        <span className="font-semibold">Timeline:</span> {service.timeline}
                      </div>
                      <div>
                        <span className="font-semibold">Cost:</span> {service.cost}
                      </div>
                    </div>
                    
                    <button 
                      onClick={() => setSelectedService(service.id)}
                      className="w-full bg-blue-500 text-white py-2 px-4 rounded font-semibold hover:bg-blue-600"
                    >
                      Learn More
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {activeTab === "ipo" && (
          <div className="max-w-4xl mx-auto">
            <div className="border rounded-lg p-6">
              <h2 className="text-2xl font-bold mb-6 flex items-center gap-2">
                🏢 IPO Application Portal
              </h2>
              
              <form className="space-y-6">
                <div className="grid md:grid-cols-2 gap-6">
                  <div>
                    <label className="block font-semibold mb-2">Company Name</label>
                    <input type="text" className="w-full border rounded p-3" placeholder="Your Company Inc." />
                  </div>
                  <div>
                    <label className="block font-semibold mb-2">Stock Symbol</label>
                    <input type="text" className="w-full border rounded p-3" placeholder="SYMB" maxLength={5} />
                  </div>
                </div>
                
                <div className="grid md:grid-cols-2 gap-6">
                  <div>
                    <label className="block font-semibold mb-2">Industry Sector</label>
                    <select className="w-full border rounded p-3">
                      <option>Technology</option>
                      <option>Healthcare</option>
                      <option>Finance</option>
                      <option>Energy</option>
                      <option>Manufacturing</option>
                      <option>Retail</option>
                      <option>Other</option>
                    </select>
                  </div>
                  <div>
                    <label className="block font-semibold mb-2">Company Valuation</label>
                    <input type="number" className="w-full border rounded p-3" placeholder="1000000000" />
                  </div>
                </div>
                
                <div>
                  <label className="block font-semibold mb-2">Business Description</label>
                  <textarea className="w-full border rounded p-3 h-32" placeholder="Describe your business, products/services, and value proposition..."></textarea>
                </div>
                
                <div className="grid md:grid-cols-3 gap-6">
                  <div>
                    <label className="block font-semibold mb-2">Shares to Issue</label>
                    <input type="number" className="w-full border rounded p-3" placeholder="100000000" />
                  </div>
                  <div>
                    <label className="block font-semibold mb-2">Initial Share Price</label>
                    <input type="number" className="w-full border rounded p-3" placeholder="10.00" step="0.01" />
                  </div>
                  <div>
                    <label className="block font-semibold mb-2">Preferred Validator Tier</label>
                    <select className="w-full border rounded p-3">
                      <option>Platinum ($50K)</option>
                      <option>Gold ($25K)</option>
                      <option>Silver ($10K)</option>
                      <option>Bronze ($1K)</option>
                    </select>
                  </div>
                </div>
                
                <div className="bg-blue-50 border border-blue-200 rounded p-4">
                  <h4 className="font-semibold text-blue-800 mb-2">Required Documents</h4>
                  <div className="grid md:grid-cols-2 gap-2 text-sm text-blue-700">
                    <div>✓ Audited financial statements (3 years)</div>
                    <div>✓ Business registration documents</div>
                    <div>✓ Management team bios</div>
                    <div>✓ Corporate governance structure</div>
                    <div>✓ Risk assessment documentation</div>
                    <div>✓ Technology/IP documentation</div>
                  </div>
                </div>
                
                <div className="flex gap-4">
                  <button type="button" className="flex-1 border border-blue-500 text-blue-500 py-3 px-6 rounded font-semibold">
                    Save Draft
                  </button>
                  <button type="submit" className="flex-1 bg-blue-500 text-white py-3 px-6 rounded font-semibold">
                    Submit IPO Application
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {activeTab === "validator" && (
          <div className="space-y-6">
            <div className="border rounded-lg p-6">
              <h2 className="text-2xl font-bold mb-6 flex items-center gap-2">
                🛡️ Business Validator Registration
              </h2>
              
              <div className="grid md:grid-cols-4 gap-4 mb-6">
                {[
                  { tier: "Bronze", stake: "50K HODL", companies: "1-5", rewards: "0.5%", color: "orange", testnet: "1K HODL" },
                  { tier: "Silver", stake: "100K HODL", companies: "6-15", rewards: "1.0%", color: "gray", testnet: "5K HODL" },
                  { tier: "Gold", stake: "250K HODL", companies: "16-30", rewards: "1.5%", color: "yellow", testnet: "10K HODL" },
                  { tier: "Platinum", stake: "500K HODL", companies: "30+", rewards: "2.0%", color: "purple", testnet: "30K HODL" }
                ].map((tier) => (
                  <div key={tier.tier} className="border rounded-lg p-4 text-center">
                    <div className="text-2xl mb-2">
                      {tier.tier === "Bronze" ? "🥉" : tier.tier === "Silver" ? "🥈" : tier.tier === "Gold" ? "🥇" : "💎"}
                    </div>
                    <h3 className="font-bold mb-2">{tier.tier} Tier</h3>
                    <div className="text-sm space-y-1">
                      <div><strong>Mainnet:</strong> {tier.stake}</div>
                      <div className="text-sm text-muted-foreground">Testnet: {tier.testnet}</div>
                      <div><strong>Capacity:</strong> {tier.companies}</div>
                      <div><strong>Rewards:</strong> {tier.rewards}</div>
                    </div>
                  </div>
                ))}
              </div>
              
              <form className="space-y-6">
                <div className="grid md:grid-cols-2 gap-6">
                  <div>
                    <label className="block font-semibold mb-2">Organization Name</label>
                    <input type="text" className="w-full border rounded p-3" placeholder="Your Validation Firm" />
                  </div>
                  <div>
                    <label className="block font-semibold mb-2">Desired Validator Tier</label>
                    <select className="w-full border rounded p-3">
                      <option>Bronze (10K HODL)</option>
                      <option>Silver (25K HODL)</option>
                      <option>Gold (50K HODL)</option>
                      <option>Platinum (100K HODL)</option>
                    </select>
                  </div>
                </div>
                
                <div>
                  <label className="block font-semibold mb-2">Specialization Areas</label>
                  <div className="grid md:grid-cols-3 gap-4">
                    {["Technology", "Healthcare", "Finance", "Energy", "Manufacturing", "Retail", "ESG/Sustainability", "Emerging Markets", "IPO Services"].map((spec) => (
                      <label key={spec} className="flex items-center gap-2">
                        <input type="checkbox" className="rounded" />
                        <span className="text-sm">{spec}</span>
                      </label>
                    ))}
                  </div>
                </div>
                
                <div>
                  <label className="block font-semibold mb-2">Team Credentials</label>
                  <textarea className="w-full border rounded p-3 h-32" placeholder="Describe your team's experience, certifications, and relevant background..."></textarea>
                </div>
                
                <div className="grid md:grid-cols-2 gap-6">
                  <div>
                    <label className="block font-semibold mb-2">Years of Experience</label>
                    <input type="number" className="w-full border rounded p-3" placeholder="10" />
                  </div>
                  <div>
                    <label className="block font-semibold mb-2">Previous Validations</label>
                    <input type="number" className="w-full border rounded p-3" placeholder="25" />
                  </div>
                </div>
                
                <div className="bg-purple-50 border border-purple-200 rounded p-4">
                  <h4 className="font-semibold text-purple-800 mb-2">Validator Benefits</h4>
                  <div className="grid md:grid-cols-2 gap-2 text-sm text-purple-700">
                    <div>✓ Earn validation rewards</div>
                    <div>✓ Governance voting rights</div>
                    <div>✓ Network reputation building</div>
                    <div>✓ Business networking opportunities</div>
                    <div>✓ Professional development</div>
                    <div>✓ Revenue generation</div>
                  </div>
                </div>
                
                <button type="submit" className="w-full bg-purple-500 text-white py-3 px-6 rounded font-semibold">
                  Apply to Become Validator
                </button>
              </form>
            </div>
          </div>
        )}

        {activeTab === "companies" && (
          <div className="space-y-6">
            <div className="flex justify-between items-center">
              <h2 className="text-2xl font-bold">Listed Companies</h2>
              <button className="bg-blue-500 text-white px-4 py-2 rounded font-semibold">
                Apply for Listing
              </button>
            </div>
            
            <div className="grid gap-4">
              {listedCompanies.map((company) => (
                <div key={company.symbol} className="border rounded-lg p-6 hover:shadow-lg transition-shadow">
                  <div className="flex justify-between items-start">
                    <div>
                      <h3 className="font-bold text-lg">{company.name}</h3>
                      <p className="text-gray-600">{company.symbol}</p>
                    </div>
                    <div className="text-right">
                      <div className="text-2xl font-bold text-green-600">{company.price}</div>
                      <div className="text-sm text-gray-500">{company.status}</div>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mt-4 text-sm">
                    <div>
                      <span className="text-gray-500">Market Cap:</span>
                      <div className="font-semibold">${company.marketCap}</div>
                    </div>
                    <div>
                      <span className="text-gray-500">Shares Outstanding:</span>
                      <div className="font-semibold">{company.shares}</div>
                    </div>
                    <div>
                      <span className="text-gray-500">24h Volume:</span>
                      <div className="font-semibold">$45.2M</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
            
            <div className="border rounded-lg p-6">
              <h3 className="font-bold text-lg mb-4">Pending IPO Applications</h3>
              <div className="space-y-4">
                {pendingIPOs.map((ipo, i) => (
                  <div key={i} className="border rounded p-4">
                    <div className="flex justify-between items-start">
                      <div>
                        <h4 className="font-semibold">{ipo.company}</h4>
                        <p className="text-sm text-gray-600">{ipo.sector}</p>
                      </div>
                      <div className="text-right">
                        <div className="font-semibold">{ipo.valuation}</div>
                        <div className="text-sm text-gray-500">{ipo.shares} shares</div>
                      </div>
                    </div>
                    <div className="flex justify-between items-center mt-3">
                      <span className={`px-3 py-1 rounded text-xs font-semibold ${
                        ipo.status === "Final Approval" 
                          ? "bg-green-100 text-green-800"
                          : ipo.status === "Due Diligence"
                          ? "bg-yellow-100 text-yellow-800"
                          : "bg-blue-100 text-blue-800"
                      }`}>
                        {ipo.status}
                      </span>
                      <span className="text-sm text-gray-500">Validator: {ipo.validator}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}