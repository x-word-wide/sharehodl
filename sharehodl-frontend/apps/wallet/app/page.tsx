"use client";

import { useState } from "react";
import { WalletButton, useWallet, useBlockchain, useStaking } from "@repo/ui";
import { Copy, Check, Send, Download, RefreshCw, Wallet, TrendingUp, Clock, ExternalLink, Coins, Award, ArrowUpRight, ArrowDownRight, Loader2, AlertCircle, CheckCircle } from "lucide-react";

export default function Home() {
  const { address, balances, connected, refreshBalances } = useWallet();
  const { networkStatus } = useBlockchain();
  const staking = useStaking(address);

  const [copied, setCopied] = useState(false);
  const [activeTab, setActiveTab] = useState<'assets' | 'staking' | 'send' | 'receive' | 'history'>('assets');
  const [sendAmount, setSendAmount] = useState('');
  const [sendAddress, setSendAddress] = useState('');
  const [sendDenom, setSendDenom] = useState('uhodl');
  const [delegateAmount, setDelegateAmount] = useState('');
  const [selectedValidator, setSelectedValidator] = useState('');
  const [undelegateAmount, setUndelegateAmount] = useState('');
  const [showUndelegateModal, setShowUndelegateModal] = useState<string | null>(null);

  const handleCopy = async () => {
    if (address) {
      await navigator.clipboard.writeText(address);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const formatAddress = (addr: string) => {
    if (!addr) return '';
    return `${addr.slice(0, 12)}...${addr.slice(-8)}`;
  };

  const formatBalance = (amount: string, denom: string) => {
    const num = parseInt(amount) / 1000000;
    const symbol = denom.replace('u', '').toUpperCase();
    return { amount: num.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 6 }), symbol };
  };

  const formatStakeAmount = (amount: string) => {
    const num = parseFloat(amount) / 1000000;
    return num.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  };

  const getTotalValue = () => {
    let total = 0;
    balances.forEach(b => {
      const amount = parseInt(b.amount) / 1000000;
      if (b.denom === 'uhodl') {
        total += amount;
      } else {
        total += amount * 0.01;
      }
    });
    return total.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  };

  const handleDelegate = async () => {
    if (!selectedValidator || !delegateAmount) return;
    const success = await staking.delegate(selectedValidator, delegateAmount);
    if (success) {
      setDelegateAmount('');
      setSelectedValidator('');
    }
  };

  const handleUndelegate = async (validatorAddress: string) => {
    if (!undelegateAmount) return;
    const success = await staking.undelegate(validatorAddress, undelegateAmount);
    if (success) {
      setUndelegateAmount('');
      setShowUndelegateModal(null);
    }
  };

  const handleClaimRewards = async (validatorAddress?: string) => {
    await staking.claimRewards(validatorAddress);
  };

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Header */}
      <header className="border-b border-gray-800 sticky top-0 z-50 bg-gray-950/95 backdrop-blur">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-2">
            <Wallet className="h-6 w-6 text-blue-400" />
            <span className="text-2xl font-bold bg-gradient-to-r from-blue-500 to-purple-500 bg-clip-text text-transparent">
              ShareWallet
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

      {/* Transaction Status Toast */}
      {(staking.txStatus.loading || staking.txStatus.error || staking.txStatus.success) && (
        <div className="fixed top-20 right-4 z-50 max-w-sm">
          {staking.txStatus.loading && (
            <div className="bg-blue-900/90 border border-blue-700 rounded-lg p-4 flex items-center gap-3 shadow-lg">
              <Loader2 className="h-5 w-5 text-blue-400 animate-spin" />
              <span>Processing transaction...</span>
            </div>
          )}
          {staking.txStatus.error && (
            <div className="bg-red-900/90 border border-red-700 rounded-lg p-4 shadow-lg">
              <div className="flex items-center gap-3 mb-2">
                <AlertCircle className="h-5 w-5 text-red-400" />
                <span className="font-semibold">Transaction Failed</span>
              </div>
              <p className="text-sm text-red-300">{staking.txStatus.error}</p>
              <button
                onClick={staking.clearTxStatus}
                className="mt-2 text-sm text-red-400 hover:text-red-300"
              >
                Dismiss
              </button>
            </div>
          )}
          {staking.txStatus.success && (
            <div className="bg-green-900/90 border border-green-700 rounded-lg p-4 shadow-lg">
              <div className="flex items-center gap-3 mb-2">
                <CheckCircle className="h-5 w-5 text-green-400" />
                <span className="font-semibold">Success!</span>
              </div>
              <p className="text-sm text-green-300">{staking.txStatus.success}</p>
              <button
                onClick={staking.clearTxStatus}
                className="mt-2 text-sm text-green-400 hover:text-green-300"
              >
                Dismiss
              </button>
            </div>
          )}
        </div>
      )}

      <main className="container mx-auto px-4 py-8">
        {!connected ? (
          /* Not Connected State */
          <div className="max-w-lg mx-auto text-center py-16">
            <div className="w-24 h-24 bg-gray-800 rounded-full flex items-center justify-center mx-auto mb-6">
              <Wallet className="h-12 w-12 text-gray-500" />
            </div>
            <h1 className="text-3xl font-bold mb-4">Welcome to ShareWallet</h1>
            <p className="text-gray-400 mb-8">
              Connect your Keplr wallet to manage your ShareHODL assets, view balances,
              and send/receive tokens securely.
            </p>
            <WalletButton />
            <div className="mt-12 grid gap-6 md:grid-cols-3 text-left">
              <div className="border border-gray-800 rounded-xl p-6">
                <div className="w-10 h-10 bg-blue-900/30 rounded-lg flex items-center justify-center mb-4">
                  <TrendingUp className="h-5 w-5 text-blue-400" />
                </div>
                <h3 className="font-semibold mb-2">Portfolio Tracking</h3>
                <p className="text-sm text-gray-500">View all your assets and their values in real-time</p>
              </div>
              <div className="border border-gray-800 rounded-xl p-6">
                <div className="w-10 h-10 bg-green-900/30 rounded-lg flex items-center justify-center mb-4">
                  <Send className="h-5 w-5 text-green-400" />
                </div>
                <h3 className="font-semibold mb-2">Instant Transfers</h3>
                <p className="text-sm text-gray-500">Send tokens with 2-second settlement times</p>
              </div>
              <div className="border border-gray-800 rounded-xl p-6">
                <div className="w-10 h-10 bg-purple-900/30 rounded-lg flex items-center justify-center mb-4">
                  <Coins className="h-5 w-5 text-purple-400" />
                </div>
                <h3 className="font-semibold mb-2">Staking Rewards</h3>
                <p className="text-sm text-gray-500">Stake your tokens and earn rewards</p>
              </div>
            </div>
          </div>
        ) : (
          /* Connected State */
          <div className="max-w-4xl mx-auto">
            {/* Portfolio Summary */}
            <div className="bg-gradient-to-r from-blue-900/30 to-purple-900/30 border border-gray-800 rounded-2xl p-6 mb-8">
              <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                <div>
                  <p className="text-sm text-gray-400 mb-1">Total Portfolio Value</p>
                  <h2 className="text-4xl font-bold">${getTotalValue()}</h2>
                </div>
                <div className="flex items-center gap-3">
                  <div className="bg-gray-900/50 rounded-lg px-4 py-2">
                    <p className="text-xs text-gray-500 mb-1">Address</p>
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-sm">{formatAddress(address || '')}</span>
                      <button onClick={handleCopy} className="text-gray-400 hover:text-white">
                        {copied ? <Check className="h-4 w-4 text-green-400" /> : <Copy className="h-4 w-4" />}
                      </button>
                    </div>
                  </div>
                  <button
                    onClick={() => { refreshBalances(); staking.refresh(); }}
                    className="p-3 bg-gray-800 hover:bg-gray-700 rounded-lg transition-colors"
                  >
                    <RefreshCw className={`h-5 w-5 ${staking.loading ? 'animate-spin' : ''}`} />
                  </button>
                </div>
              </div>
            </div>

            {/* Tab Navigation */}
            <div className="flex gap-2 mb-6 border-b border-gray-800 pb-2 overflow-x-auto">
              {[
                { id: 'assets' as const, label: 'Assets', icon: Wallet },
                { id: 'staking' as const, label: 'Staking', icon: Coins },
                { id: 'send' as const, label: 'Send', icon: Send },
                { id: 'receive' as const, label: 'Receive', icon: Download },
                { id: 'history' as const, label: 'History', icon: Clock },
              ].map(tab => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`flex items-center gap-2 px-4 py-2 rounded-lg transition-colors whitespace-nowrap ${
                    activeTab === tab.id
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-400 hover:text-white hover:bg-gray-800'
                  }`}
                >
                  <tab.icon className="h-4 w-4" />
                  {tab.label}
                </button>
              ))}
            </div>

            {/* Assets Tab */}
            {activeTab === 'assets' && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">Your Assets</h3>
                {balances.length === 0 ? (
                  <div className="border border-gray-800 rounded-xl p-8 text-center">
                    <p className="text-gray-500">No assets found</p>
                    <p className="text-sm text-gray-600 mt-2">
                      Use the faucet or receive tokens to get started
                    </p>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {balances.map((balance, index) => {
                      const { amount, symbol } = formatBalance(balance.amount, balance.denom);
                      return (
                        <div key={index} className="border border-gray-800 rounded-xl p-4 flex items-center justify-between hover:border-gray-700 transition-colors">
                          <div className="flex items-center gap-4">
                            <div className={`w-10 h-10 rounded-full flex items-center justify-center ${
                              symbol === 'HODL' ? 'bg-blue-900/50 text-blue-400' : 'bg-gray-800 text-gray-400'
                            }`}>
                              <span className="font-bold text-sm">{symbol.charAt(0)}</span>
                            </div>
                            <div>
                              <p className="font-semibold">{symbol}</p>
                              <p className="text-sm text-gray-500">{balance.denom}</p>
                            </div>
                          </div>
                          <div className="text-right">
                            <p className="font-semibold">{amount}</p>
                            <p className="text-sm text-gray-500">
                              ${symbol === 'HODL' ? amount : (parseFloat(amount.replace(/,/g, '')) * 0.01).toFixed(2)}
                            </p>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            )}

            {/* Staking Tab */}
            {activeTab === 'staking' && (
              <div className="space-y-6">
                {/* Loading State */}
                {staking.loading && (
                  <div className="flex items-center justify-center py-12">
                    <Loader2 className="h-8 w-8 text-blue-400 animate-spin" />
                    <span className="ml-3 text-gray-400">Loading staking data...</span>
                  </div>
                )}

                {/* Error State */}
                {staking.error && (
                  <div className="bg-red-900/20 border border-red-800 rounded-xl p-4 flex items-center gap-3">
                    <AlertCircle className="h-5 w-5 text-red-400" />
                    <span className="text-red-300">{staking.error}</span>
                    <button onClick={staking.refresh} className="ml-auto text-red-400 hover:text-red-300">
                      Retry
                    </button>
                  </div>
                )}

                {!staking.loading && (
                  <>
                    {/* Staking Overview */}
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                      <div className="bg-gradient-to-br from-green-900/30 to-green-800/20 border border-green-800/50 rounded-xl p-4">
                        <div className="flex items-center gap-2 text-green-400 mb-2">
                          <Coins className="h-4 w-4" />
                          <span className="text-sm">Total Staked</span>
                        </div>
                        <p className="text-2xl font-bold">{formatStakeAmount(staking.totalStaked)}</p>
                        <p className="text-sm text-gray-500">STAKE</p>
                      </div>
                      <div className="bg-gradient-to-br from-yellow-900/30 to-yellow-800/20 border border-yellow-800/50 rounded-xl p-4">
                        <div className="flex items-center gap-2 text-yellow-400 mb-2">
                          <Award className="h-4 w-4" />
                          <span className="text-sm">Pending Rewards</span>
                        </div>
                        <p className="text-2xl font-bold">{formatStakeAmount(staking.totalRewards)}</p>
                        <p className="text-sm text-gray-500">STAKE</p>
                      </div>
                      <div className="bg-gradient-to-br from-blue-900/30 to-blue-800/20 border border-blue-800/50 rounded-xl p-4">
                        <div className="flex items-center gap-2 text-blue-400 mb-2">
                          <TrendingUp className="h-4 w-4" />
                          <span className="text-sm">Current APR</span>
                        </div>
                        <p className="text-2xl font-bold">~12.5%</p>
                        <p className="text-sm text-gray-500">Annual</p>
                      </div>
                      <div className="bg-gradient-to-br from-purple-900/30 to-purple-800/20 border border-purple-800/50 rounded-xl p-4">
                        <div className="flex items-center gap-2 text-purple-400 mb-2">
                          <Clock className="h-4 w-4" />
                          <span className="text-sm">Unbonding</span>
                        </div>
                        <p className="text-2xl font-bold">{formatStakeAmount(staking.unbonding)}</p>
                        <p className="text-sm text-gray-500">STAKE</p>
                      </div>
                    </div>

                    {/* Claim Rewards Button */}
                    {parseFloat(staking.totalRewards) > 0 && (
                      <div className="bg-yellow-900/20 border border-yellow-800/50 rounded-xl p-4 flex items-center justify-between">
                        <div>
                          <p className="font-semibold text-yellow-400">Claim Your Rewards</p>
                          <p className="text-sm text-gray-400">You have {formatStakeAmount(staking.totalRewards)} STAKE in pending rewards</p>
                        </div>
                        <button
                          onClick={() => handleClaimRewards()}
                          disabled={staking.txStatus.loading}
                          className="bg-yellow-600 hover:bg-yellow-700 disabled:bg-yellow-800 disabled:cursor-not-allowed text-black font-semibold px-6 py-2 rounded-lg transition-colors flex items-center gap-2"
                        >
                          {staking.txStatus.loading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                          Claim All
                        </button>
                      </div>
                    )}

                    {/* My Delegations */}
                    <div>
                      <h3 className="text-lg font-semibold mb-4">My Delegations</h3>
                      {staking.delegations.length > 0 ? (
                        <div className="space-y-3">
                          {staking.delegations.map((del, index) => (
                            <div key={index} className="border border-gray-800 rounded-xl p-4 hover:border-gray-700 transition-colors">
                              <div className="flex items-center justify-between mb-3">
                                <div className="flex items-center gap-3">
                                  <div className="w-10 h-10 rounded-full bg-blue-900/50 flex items-center justify-center text-blue-400 font-bold">
                                    {del.validatorMoniker.charAt(0)}
                                  </div>
                                  <div>
                                    <p className="font-semibold">{del.validatorMoniker}</p>
                                    <p className="text-sm text-gray-500 font-mono">{formatAddress(del.validatorAddress)}</p>
                                  </div>
                                </div>
                                <div className="text-right">
                                  <p className="font-semibold">{formatStakeAmount(del.amount)} STAKE</p>
                                  <p className="text-sm text-green-400">+{formatStakeAmount(del.rewards)} rewards</p>
                                </div>
                              </div>
                              <div className="flex items-center gap-2 pt-3 border-t border-gray-800">
                                <button
                                  onClick={() => setSelectedValidator(del.validatorAddress)}
                                  className="flex-1 bg-green-600/20 hover:bg-green-600/30 text-green-400 py-2 rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2"
                                >
                                  <ArrowUpRight className="h-4 w-4" />
                                  Delegate More
                                </button>
                                <button
                                  onClick={() => setShowUndelegateModal(del.validatorAddress)}
                                  className="flex-1 bg-red-600/20 hover:bg-red-600/30 text-red-400 py-2 rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2"
                                >
                                  <ArrowDownRight className="h-4 w-4" />
                                  Undelegate
                                </button>
                                <button
                                  onClick={() => handleClaimRewards(del.validatorAddress)}
                                  disabled={parseFloat(del.rewards) === 0 || staking.txStatus.loading}
                                  className="flex-1 bg-yellow-600/20 hover:bg-yellow-600/30 disabled:opacity-50 disabled:cursor-not-allowed text-yellow-400 py-2 rounded-lg text-sm font-medium transition-colors"
                                >
                                  Claim
                                </button>
                              </div>

                              {/* Undelegate Modal */}
                              {showUndelegateModal === del.validatorAddress && (
                                <div className="mt-4 p-4 bg-gray-900 rounded-lg border border-gray-700">
                                  <h4 className="font-semibold mb-3">Undelegate from {del.validatorMoniker}</h4>
                                  <p className="text-sm text-gray-400 mb-3">
                                    Undelegated tokens will be available after 21 days unbonding period.
                                  </p>
                                  <input
                                    type="number"
                                    value={undelegateAmount}
                                    onChange={(e) => setUndelegateAmount(e.target.value)}
                                    placeholder="Amount to undelegate"
                                    max={parseFloat(del.amount) / 1000000}
                                    className="w-full bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 mb-3 focus:outline-none focus:ring-2 focus:ring-red-500"
                                  />
                                  <div className="flex gap-2">
                                    <button
                                      onClick={() => handleUndelegate(del.validatorAddress)}
                                      disabled={!undelegateAmount || staking.txStatus.loading}
                                      className="flex-1 bg-red-600 hover:bg-red-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white py-2 rounded-lg font-semibold transition-colors"
                                    >
                                      Confirm Undelegate
                                    </button>
                                    <button
                                      onClick={() => { setShowUndelegateModal(null); setUndelegateAmount(''); }}
                                      className="px-4 bg-gray-700 hover:bg-gray-600 text-white py-2 rounded-lg transition-colors"
                                    >
                                      Cancel
                                    </button>
                                  </div>
                                </div>
                              )}
                            </div>
                          ))}
                        </div>
                      ) : (
                        <div className="border border-gray-800 rounded-xl p-8 text-center">
                          <Coins className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                          <p className="text-gray-500">No active delegations</p>
                          <p className="text-sm text-gray-600 mt-2">Stake your tokens to earn rewards</p>
                        </div>
                      )}
                    </div>

                    {/* Delegate to Validator */}
                    <div>
                      <h3 className="text-lg font-semibold mb-4">Delegate to Validator</h3>
                      <div className="border border-gray-800 rounded-xl p-6">
                        <div className="space-y-4">
                          <div>
                            <label className="block text-sm text-gray-400 mb-2">Select Validator</label>
                            <select
                              value={selectedValidator}
                              onChange={(e) => setSelectedValidator(e.target.value)}
                              className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
                            >
                              <option value="">Choose a validator...</option>
                              {staking.validators.map((val, index) => (
                                <option key={index} value={val.operatorAddress}>
                                  {val.moniker} - {val.commission}% commission - {val.votingPower}% power
                                </option>
                              ))}
                            </select>
                          </div>
                          <div>
                            <label className="block text-sm text-gray-400 mb-2">Amount to Delegate</label>
                            <div className="relative">
                              <input
                                type="number"
                                value={delegateAmount}
                                onChange={(e) => setDelegateAmount(e.target.value)}
                                placeholder="0.00"
                                className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 pr-20 focus:outline-none focus:ring-2 focus:ring-blue-500"
                              />
                              <span className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500">STAKE</span>
                            </div>
                          </div>
                          <div className="bg-gray-900 border border-gray-700 rounded-lg p-4">
                            <div className="flex justify-between text-sm mb-2">
                              <span className="text-gray-400">Unbonding Period</span>
                              <span>21 days</span>
                            </div>
                            <div className="flex justify-between text-sm">
                              <span className="text-gray-400">Estimated APR</span>
                              <span className="text-green-400">~12.5%</span>
                            </div>
                          </div>
                          <button
                            onClick={handleDelegate}
                            disabled={!selectedValidator || !delegateAmount || staking.txStatus.loading}
                            className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white py-3 rounded-lg font-semibold transition-colors flex items-center justify-center gap-2"
                          >
                            {staking.txStatus.loading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                            Delegate
                          </button>
                        </div>
                      </div>
                    </div>

                    {/* All Validators */}
                    <div>
                      <div className="flex items-center justify-between mb-4">
                        <h3 className="text-lg font-semibold">All Validators ({staking.validators.length})</h3>
                        <button
                          onClick={staking.refresh}
                          className="text-sm text-gray-400 hover:text-white flex items-center gap-2"
                        >
                          <RefreshCw className={`h-4 w-4 ${staking.loading ? 'animate-spin' : ''}`} />
                          Refresh
                        </button>
                      </div>
                      {staking.validators.length > 0 ? (
                        <div className="border border-gray-800 rounded-xl overflow-hidden">
                          <table className="w-full">
                            <thead className="bg-gray-900/50">
                              <tr>
                                <th className="text-left px-4 py-3 text-sm text-gray-400 font-medium">Validator</th>
                                <th className="text-right px-4 py-3 text-sm text-gray-400 font-medium">Voting Power</th>
                                <th className="text-right px-4 py-3 text-sm text-gray-400 font-medium">Commission</th>
                                <th className="text-right px-4 py-3 text-sm text-gray-400 font-medium">Action</th>
                              </tr>
                            </thead>
                            <tbody>
                              {staking.validators.slice(0, 20).map((val, index) => (
                                <tr key={index} className="border-t border-gray-800 hover:bg-gray-900/30 transition-colors">
                                  <td className="px-4 py-3">
                                    <div className="flex items-center gap-3">
                                      <div className="w-8 h-8 rounded-full bg-blue-900/50 flex items-center justify-center text-blue-400 text-sm font-bold">
                                        {val.moniker.charAt(0)}
                                      </div>
                                      <div>
                                        <p className="font-medium">{val.moniker}</p>
                                        <p className="text-xs text-gray-500 font-mono">{formatAddress(val.operatorAddress)}</p>
                                      </div>
                                    </div>
                                  </td>
                                  <td className="px-4 py-3 text-right">{val.votingPower}%</td>
                                  <td className="px-4 py-3 text-right">{val.commission}%</td>
                                  <td className="px-4 py-3 text-right">
                                    <button
                                      onClick={() => setSelectedValidator(val.operatorAddress)}
                                      className="bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 px-4 py-1.5 rounded-lg text-sm font-medium transition-colors"
                                    >
                                      Delegate
                                    </button>
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      ) : (
                        <div className="border border-gray-800 rounded-xl p-8 text-center">
                          <p className="text-gray-500">No validators found</p>
                          <p className="text-sm text-gray-600 mt-2">Make sure the blockchain is running</p>
                        </div>
                      )}
                    </div>
                  </>
                )}
              </div>
            )}

            {/* Send Tab */}
            {activeTab === 'send' && (
              <div className="max-w-lg">
                <h3 className="text-lg font-semibold mb-4">Send Tokens</h3>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm text-gray-400 mb-2">Token</label>
                    <select
                      value={sendDenom}
                      onChange={(e) => setSendDenom(e.target.value)}
                      className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      {balances.length > 0 ? (
                        balances.map((b, i) => (
                          <option key={i} value={b.denom}>
                            {b.denom.replace('u', '').toUpperCase()} (Balance: {formatBalance(b.amount, b.denom).amount})
                          </option>
                        ))
                      ) : (
                        <option value="uhodl">HODL</option>
                      )}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm text-gray-400 mb-2">Recipient Address</label>
                    <input
                      type="text"
                      value={sendAddress}
                      onChange={(e) => setSendAddress(e.target.value)}
                      placeholder="hodl1..."
                      className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm text-gray-400 mb-2">Amount</label>
                    <input
                      type="number"
                      value={sendAmount}
                      onChange={(e) => setSendAmount(e.target.value)}
                      placeholder="0.00"
                      className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <div className="bg-gray-900 border border-gray-700 rounded-lg p-4">
                    <div className="flex justify-between text-sm mb-2">
                      <span className="text-gray-400">Network Fee</span>
                      <span>~0.005 HODL</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">Settlement Time</span>
                      <span className="text-green-400">~2 seconds</span>
                    </div>
                  </div>
                  <button
                    disabled={!sendAddress || !sendAmount}
                    className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white py-3 rounded-lg font-semibold transition-colors"
                  >
                    Send Tokens
                  </button>
                </div>
              </div>
            )}

            {/* Receive Tab */}
            {activeTab === 'receive' && (
              <div className="max-w-lg">
                <h3 className="text-lg font-semibold mb-4">Receive Tokens</h3>
                <div className="border border-gray-800 rounded-xl p-6">
                  <p className="text-sm text-gray-400 mb-4">
                    Share your address to receive tokens on the ShareHODL network
                  </p>
                  <div className="bg-gray-900 border border-gray-700 rounded-lg p-4 mb-4">
                    <p className="font-mono text-sm break-all">{address}</p>
                  </div>
                  <div className="flex gap-3">
                    <button
                      onClick={handleCopy}
                      className="flex-1 bg-blue-600 hover:bg-blue-700 text-white py-3 rounded-lg font-semibold transition-colors flex items-center justify-center gap-2"
                    >
                      {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                      {copied ? 'Copied!' : 'Copy Address'}
                    </button>
                  </div>
                  <div className="mt-6 p-4 bg-gray-900/50 rounded-lg border border-gray-800">
                    <h4 className="font-semibold mb-2 text-sm">Supported Tokens</h4>
                    <ul className="text-sm text-gray-400 space-y-1">
                      <li>- HODL (Native stablecoin)</li>
                      <li>- STAKE (Staking token)</li>
                      <li>- All listed equity tokens</li>
                      <li>- IBC tokens (cross-chain)</li>
                    </ul>
                  </div>
                </div>
              </div>
            )}

            {/* History Tab */}
            {activeTab === 'history' && (
              <div>
                <h3 className="text-lg font-semibold mb-4">Transaction History</h3>
                <div className="border border-gray-800 rounded-xl p-8 text-center">
                  <Clock className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                  <p className="text-gray-500">No transactions yet</p>
                  <p className="text-sm text-gray-600 mt-2">
                    Your transaction history will appear here
                  </p>
                  <a
                    href={`http://localhost:3003?search=${address}`}
                    className="inline-flex items-center gap-2 text-blue-400 hover:text-blue-300 mt-4 text-sm"
                  >
                    View in Explorer <ExternalLink className="h-4 w-4" />
                  </a>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Footer */}
        <div className="text-center text-gray-500 pt-12 mt-12 border-t border-gray-800">
          <p className="mb-2 text-gray-400">ShareWallet - Secure Asset Management</p>
          <p className="text-sm mb-4">
            Part of the ShareHODL blockchain platform
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
