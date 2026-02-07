/**
 * Position Checker Service
 *
 * Checks if a user has active positions that would prevent unstaking:
 * - Active escrows (as sender or recipient)
 * - Active disputes (as involved party)
 * - Active loans (as borrower or lender)
 * - Active P2P trades (when implemented)
 *
 * SECURITY NOTE: This is a frontend check only. For production,
 * these checks should also be enforced at the blockchain level
 * via an ante handler or message server hook.
 */

import { API_BASE } from './stakingStore';

// Position types that block unstaking
export interface ActivePosition {
  type: 'escrow' | 'dispute' | 'loan' | 'p2p' | 'lending_stake';
  id: string;
  description: string;
  amount?: number;
}

export interface PositionCheckResult {
  canUnstake: boolean;
  blockedBy: ActivePosition[];
  error?: string;
}

/**
 * Escrow status values that indicate an active position
 */
const ACTIVE_ESCROW_STATUSES = [
  'Pending',
  'Funded',
  'Disputed',
  'UnderReview'
];

/**
 * Dispute status values that indicate an active position
 */
const ACTIVE_DISPUTE_STATUSES = [
  'Open',
  'UnderReview',
  'Voting',
  'Appealed'
];

/**
 * Loan status values that indicate an active position
 */
const ACTIVE_LOAN_STATUSES = [
  'Pending',
  'Active',
  'PartiallyRepaid'
];

/**
 * Check if user has any active escrows
 * NOTE: This endpoint doesn't exist yet - will return empty until implemented
 */
async function checkActiveEscrows(address: string): Promise<ActivePosition[]> {
  const positions: ActivePosition[] = [];

  try {
    // TODO: When escrow query endpoint is implemented, use:
    // GET /sharehodl/escrow/v1/user/{address}/escrows
    const response = await fetch(`${API_BASE}/sharehodl/escrow/v1/user/${address}/escrows`);

    if (response.ok) {
      const data = await response.json();
      const escrows = data.escrows || [];

      for (const escrow of escrows) {
        if (ACTIVE_ESCROW_STATUSES.includes(escrow.status)) {
          positions.push({
            type: 'escrow',
            id: escrow.id,
            description: `Active escrow #${escrow.id} (${escrow.status})`,
            amount: parseFloat(escrow.amount) / 1_000_000
          });
        }
      }
    }
  } catch {
    // Endpoint not available yet - return empty
    console.log('Escrow query endpoint not available');
  }

  return positions;
}

/**
 * Check if user has any active disputes
 * NOTE: This endpoint doesn't exist yet - will return empty until implemented
 */
async function checkActiveDisputes(address: string): Promise<ActivePosition[]> {
  const positions: ActivePosition[] = [];

  try {
    // TODO: When dispute query endpoint is implemented, use:
    // GET /sharehodl/escrow/v1/user/{address}/disputes
    const response = await fetch(`${API_BASE}/sharehodl/escrow/v1/user/${address}/disputes`);

    if (response.ok) {
      const data = await response.json();
      const disputes = data.disputes || [];

      for (const dispute of disputes) {
        if (ACTIVE_DISPUTE_STATUSES.includes(dispute.status)) {
          positions.push({
            type: 'dispute',
            id: dispute.id,
            description: `Active dispute #${dispute.id} (${dispute.status})`,
            amount: parseFloat(dispute.value || '0') / 1_000_000
          });
        }
      }
    }
  } catch {
    // Endpoint not available yet - return empty
    console.log('Dispute query endpoint not available');
  }

  return positions;
}

/**
 * Check if user has any active loans (as borrower or lender)
 * NOTE: This endpoint doesn't exist yet - will return empty until implemented
 */
async function checkActiveLoans(address: string): Promise<ActivePosition[]> {
  const positions: ActivePosition[] = [];

  try {
    // TODO: When lending query endpoint is implemented, use:
    // GET /sharehodl/lending/v1/user/{address}/loans
    const response = await fetch(`${API_BASE}/sharehodl/lending/v1/user/${address}/loans`);

    if (response.ok) {
      const data = await response.json();
      const loans = data.loans || [];

      for (const loan of loans) {
        if (ACTIVE_LOAN_STATUSES.includes(loan.status)) {
          const role = loan.borrower === address ? 'borrower' : 'lender';
          positions.push({
            type: 'loan',
            id: loan.id,
            description: `Active loan #${loan.id} as ${role} (${loan.status})`,
            amount: parseFloat(loan.principal) / 1_000_000
          });
        }
      }
    }
  } catch {
    // Endpoint not available yet - return empty
    console.log('Lending query endpoint not available');
  }

  return positions;
}

/**
 * Check if user has lender/borrower stake (lending trust ceiling)
 * NOTE: This endpoint doesn't exist yet - will return empty until implemented
 */
async function checkLendingStake(address: string): Promise<ActivePosition[]> {
  const positions: ActivePosition[] = [];

  try {
    // Check lender stake
    const lenderResponse = await fetch(`${API_BASE}/sharehodl/lending/v1/lender_stake/${address}`);
    if (lenderResponse.ok) {
      const data = await lenderResponse.json();
      if (data.stake && parseFloat(data.stake.amount) > 0) {
        positions.push({
          type: 'lending_stake',
          id: 'lender',
          description: 'Active lender stake (Trust Ceiling)',
          amount: parseFloat(data.stake.amount) / 1_000_000
        });
      }
    }

    // Check borrower stake
    const borrowerResponse = await fetch(`${API_BASE}/sharehodl/lending/v1/borrower_stake/${address}`);
    if (borrowerResponse.ok) {
      const data = await borrowerResponse.json();
      if (data.stake && parseFloat(data.stake.amount) > 0) {
        positions.push({
          type: 'lending_stake',
          id: 'borrower',
          description: 'Active borrower stake (Trust Ceiling)',
          amount: parseFloat(data.stake.amount) / 1_000_000
        });
      }
    }
  } catch {
    // Endpoints not available yet - return empty
    console.log('Lending stake query endpoints not available');
  }

  return positions;
}

/**
 * Check if user has any active P2P trades
 * NOTE: P2P module not implemented yet - will return empty
 */
async function checkActiveP2PTrades(address: string): Promise<ActivePosition[]> {
  const positions: ActivePosition[] = [];

  try {
    // TODO: When P2P module is implemented, use:
    // GET /sharehodl/p2p/v1/user/{address}/trades
    const response = await fetch(`${API_BASE}/sharehodl/p2p/v1/user/${address}/trades`);

    if (response.ok) {
      const data = await response.json();
      const trades = data.trades || [];

      for (const trade of trades) {
        if (['Open', 'Pending', 'InProgress'].includes(trade.status)) {
          positions.push({
            type: 'p2p',
            id: trade.id,
            description: `Active P2P trade #${trade.id} (${trade.status})`,
            amount: parseFloat(trade.amount) / 1_000_000
          });
        }
      }
    }
  } catch {
    // P2P module not available yet - return empty
    console.log('P2P query endpoint not available');
  }

  return positions;
}

/**
 * Check all position types to determine if user can unstake
 *
 * @param address - User's ShareHODL address
 * @returns PositionCheckResult with canUnstake flag and blocking positions
 */
export async function checkCanUnstake(address: string): Promise<PositionCheckResult> {
  try {
    // Run all checks in parallel for performance
    const [escrows, disputes, loans, lendingStakes, p2pTrades] = await Promise.all([
      checkActiveEscrows(address),
      checkActiveDisputes(address),
      checkActiveLoans(address),
      checkLendingStake(address),
      checkActiveP2PTrades(address)
    ]);

    const blockedBy = [
      ...escrows,
      ...disputes,
      ...loans,
      ...lendingStakes,
      ...p2pTrades
    ];

    return {
      canUnstake: blockedBy.length === 0,
      blockedBy
    };
  } catch (error) {
    console.error('Error checking positions:', error);
    // On error, allow unstaking but log the issue
    // In production, you might want to be more restrictive
    return {
      canUnstake: true,
      blockedBy: [],
      error: 'Failed to check positions - proceeding with caution'
    };
  }
}

/**
 * Format blocking positions into a user-friendly message
 */
export function formatBlockingMessage(positions: ActivePosition[]): string {
  if (positions.length === 0) return '';

  const lines = positions.map(p => `- ${p.description}`);
  return `You cannot unstake while you have active positions:\n${lines.join('\n')}\n\nPlease resolve these before unstaking.`;
}
