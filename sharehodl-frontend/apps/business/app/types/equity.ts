// Anti-Dilution Types for Frontend
// These correspond to the blockchain x/equity/types definitions

export enum AntiDilutionType {
  None = 0,
  FullRatchet = 1,
  WeightedAverage = 2,
  BroadBasedWeightedAverage = 3,
}

export const AntiDilutionTypeLabels: Record<AntiDilutionType, string> = {
  [AntiDilutionType.None]: "None",
  [AntiDilutionType.FullRatchet]: "Full Ratchet",
  [AntiDilutionType.WeightedAverage]: "Weighted Average",
  [AntiDilutionType.BroadBasedWeightedAverage]: "Broad-Based Weighted Average",
};

export const AntiDilutionTypeDescriptions: Record<AntiDilutionType, string> = {
  [AntiDilutionType.None]: "No anti-dilution protection",
  [AntiDilutionType.FullRatchet]: "Adjusts to the lowest price in any down round - maximum protection",
  [AntiDilutionType.WeightedAverage]: "Adjusts based on weighted average of old and new prices",
  [AntiDilutionType.BroadBasedWeightedAverage]: "Includes all shares in calculation - more balanced protection",
};

export interface AntiDilutionProvision {
  companyId: string;
  classId: string;
  provisionType: AntiDilutionType;
  initialPrice: string;
  currentAdjustedPrice: string;
  triggerThreshold: string;
  adjustmentCap: string;
  minimumPrice: string;
  isActive: boolean;
  lastAdjustmentTime: string;
  totalAdjustments: number;
  createdAt: string;
}

export interface IssuanceRecord {
  id: string;
  companyId: string;
  classId: string;
  sharesIssued: string;
  issuePrice: string;
  totalRaised: string;
  timestamp: string;
  isDownRound: boolean;
  priceDropPercent: string;
  purpose: string;
}

export interface AntiDilutionAdjustment {
  id: string;
  companyId: string;
  classId: string;
  shareholder: string;
  originalShares: string;
  adjustedShares: string;
  sharesAdded: string;
  triggerPrice: string;
  adjustmentRatio: string;
  adjustmentType: AntiDilutionType;
  issuanceId: string;
  timestamp: string;
}

export interface ShareClass {
  companyId: string;
  classId: string;
  name: string;
  description: string;
  votingRights: boolean;
  dividendRights: boolean;
  liquidationPreference: boolean;
  conversionRights: boolean;
  authorizedShares: string;
  issuedShares: string;
  outstandingShares: string;
  parValue: string;
  transferable: boolean;
  hasAntiDilution: boolean;
  antiDilutionType: AntiDilutionType;
  createdAt: string;
  updatedAt: string;
}

export interface Shareholding {
  companyId: string;
  classId: string;
  owner: string;
  shares: string;
  vestedShares: string;
  lockedShares: string;
  costBasis: string;
  totalCost: string;
  acquiredAt: string;
  vestingSchedule?: VestingSchedule;
}

export interface VestingSchedule {
  startDate: string;
  cliffDate: string;
  endDate: string;
  totalShares: string;
  vestedShares: string;
  vestingPeriodMonths: number;
  cliffMonths: number;
}

export interface Company {
  id: string;
  name: string;
  symbol: string;
  description: string;
  founder: string;
  website: string;
  industry: string;
  country: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

// Cap Table Response Types

export interface ShareClassSummary {
  classId: string;
  name: string;
  authorizedShares: string;
  issuedShares: string;
  outstandingShares: string;
  holderCount: number;
  hasAntiDilution: boolean;
  antiDilutionType: AntiDilutionType;
  votingRights: boolean;
  dividendRights: boolean;
}

export interface ShareholderSummary {
  owner: string;
  totalShares: string;
  ownershipPercent: string;
  classId: string;
  votingPower: string;
}

export interface IssuanceStatistics {
  totalIssuances: number;
  totalSharesIssued: string;
  totalRaised: string;
  averageIssuePrice: string;
  lastIssuancePrice: string;
  downRoundCount: number;
}

export interface AntiDilutionSummary {
  protectedClasses: number;
  totalAdjustments: number;
  totalSharesAdjusted: string;
  activeProvisions: number;
}

export interface CompanyCapTable {
  company: Company;
  shareClasses: ShareClassSummary[];
  topHolders: ShareholderSummary[];
  issuanceStats: IssuanceStatistics;
  antiDilution: AntiDilutionSummary;
}

// API Request Types

export interface RegisterAntiDilutionRequest {
  companyId: string;
  classId: string;
  provisionType: AntiDilutionType;
  triggerThreshold: string;
  adjustmentCap: string;
  minimumPrice: string;
}

export interface UpdateAntiDilutionRequest {
  companyId: string;
  classId: string;
  isActive?: boolean;
  triggerThreshold?: string;
  adjustmentCap?: string;
  minimumPrice?: string;
}

export interface IssueSharesWithProtectionRequest {
  companyId: string;
  classId: string;
  recipient: string;
  shares: string;
  price: string;
  paymentDenom: string;
  purpose: string;
}

// Utility functions

export function formatShares(shares: string): string {
  const num = parseFloat(shares);
  if (num >= 1e9) return `${(num / 1e9).toFixed(2)}B`;
  if (num >= 1e6) return `${(num / 1e6).toFixed(2)}M`;
  if (num >= 1e3) return `${(num / 1e3).toFixed(2)}K`;
  return num.toLocaleString();
}

export function formatCurrency(amount: string, symbol = "$"): string {
  const num = parseFloat(amount);
  if (num >= 1e12) return `${symbol}${(num / 1e12).toFixed(2)}T`;
  if (num >= 1e9) return `${symbol}${(num / 1e9).toFixed(2)}B`;
  if (num >= 1e6) return `${symbol}${(num / 1e6).toFixed(2)}M`;
  if (num >= 1e3) return `${symbol}${(num / 1e3).toFixed(2)}K`;
  return `${symbol}${num.toFixed(2)}`;
}

export function formatPercent(value: string): string {
  const num = parseFloat(value);
  return `${num.toFixed(2)}%`;
}

export function formatDate(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function formatDateTime(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

// Calculate dilution impact
export function calculateDilutionImpact(
  currentShares: string,
  newShares: string
): { newOwnership: string; dilutionPercent: string } {
  const current = parseFloat(currentShares);
  const added = parseFloat(newShares);
  const total = current + added;

  const newOwnership = ((current / total) * 100).toFixed(2);
  const dilutionPercent = (100 - parseFloat(newOwnership)).toFixed(2);

  return { newOwnership, dilutionPercent };
}

// Calculate full ratchet adjustment
export function calculateFullRatchetAdjustment(
  originalShares: string,
  oldPrice: string,
  newPrice: string
): string {
  const shares = parseFloat(originalShares);
  const old = parseFloat(oldPrice);
  const current = parseFloat(newPrice);

  if (current >= old) return originalShares;

  const ratio = old / current;
  return (shares * ratio).toFixed(0);
}

// Calculate weighted average adjustment
export function calculateWeightedAverageAdjustment(
  originalShares: string,
  oldPrice: string,
  newPrice: string,
  newShares: string,
  totalOutstanding: string
): string {
  const shares = parseFloat(originalShares);
  const old = parseFloat(oldPrice);
  const current = parseFloat(newPrice);
  const issued = parseFloat(newShares);
  const outstanding = parseFloat(totalOutstanding);

  if (current >= old) return originalShares;

  // Weighted average price = (Old Price × Outstanding + New Price × New Shares) / (Outstanding + New Shares)
  const weightedPrice = (old * outstanding + current * issued) / (outstanding + issued);
  const ratio = old / weightedPrice;

  return (shares * ratio).toFixed(0);
}
