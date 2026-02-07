package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/sharehodl/sharehodl-blockchain/x/staking/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdStake(),
		CmdUnstake(),
		CmdClaimRewards(),
	)

	return cmd
}

// CmdStake returns the stake command
func CmdStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake [amount]",
		Short: "Stake HODL to earn rewards and increase tier",
		Long: `Stake HODL tokens to earn staking rewards and increase your tier level.

Tiers and thresholds:
- Holder: 100 HODL
- Keeper: 10,000 HODL
- Warden: 100,000 HODL
- Steward: 1,000,000 HODL
- Archon: 10,000,000 HODL
- Validator: 50,000,000 HODL

Example:
  sharehodld tx staking stake 40000000000000 --from validator
  (Stakes 40,000,000 HODL)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, ok := math.NewIntFromString(args[0])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[0])
			}

			msg := &types.MsgStake{
				Staker: clientCtx.GetFromAddress().String(),
				Amount: amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdUnstake returns the unstake command
func CmdUnstake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unstake [amount]",
		Short: "Unstake HODL (may cause tier demotion)",
		Long: `Unstake HODL tokens. This may cause a tier demotion if your remaining
stake falls below your current tier threshold.

Example:
  sharehodld tx staking unstake 1000000000000 --from validator
  (Unstakes 1,000,000 HODL)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, ok := math.NewIntFromString(args[0])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[0])
			}

			msg := &types.MsgUnstake{
				Staker: clientCtx.GetFromAddress().String(),
				Amount: amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdClaimRewards returns the claim rewards command
func CmdClaimRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-rewards",
		Short: "Claim pending staking rewards",
		Long: `Claim any pending staking rewards earned from your staked HODL.

Example:
  sharehodld tx staking claim-rewards --from validator`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgClaimRewards{
				Staker: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Helper to parse amount with unit suffix
func parseAmount(amountStr string) (math.Int, error) {
	// Try to parse as plain integer first
	amount, ok := math.NewIntFromString(amountStr)
	if ok {
		return amount, nil
	}

	// Try parsing with suffixes (e.g., "40M" = 40,000,000 HODL = 40,000,000,000,000 uhodl)
	var multiplier int64 = 1
	if len(amountStr) > 1 {
		suffix := amountStr[len(amountStr)-1]
		switch suffix {
		case 'K', 'k':
			multiplier = 1000
			amountStr = amountStr[:len(amountStr)-1]
		case 'M', 'm':
			multiplier = 1000000
			amountStr = amountStr[:len(amountStr)-1]
		case 'B', 'b':
			multiplier = 1000000000
			amountStr = amountStr[:len(amountStr)-1]
		}
	}

	baseAmount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return math.Int{}, fmt.Errorf("invalid amount format: %s", amountStr)
	}

	// Convert to micro units (uhodl) - multiply by 1,000,000
	microAmount := baseAmount * multiplier * 1000000
	return math.NewInt(microAmount), nil
}
