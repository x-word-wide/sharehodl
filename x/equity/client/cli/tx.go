package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetTxCmd returns the transaction commands for the equity module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "equity",
		Short:                      "Equity transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCreateCompanyCmd(),
		NewTransferSharesCmd(),
		NewIssueSharesCmd(),
	)

	return cmd
}

// NewCreateCompanyCmd creates a new company
func NewCreateCompanyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-company [name] [symbol] [total-shares]",
		Short: "Create a new company listing",
		Long: `Create a new company listing on ShareHODL.

Note: Full transaction support requires protobuf generation.
For now, this command shows what would be sent.

Example:
  sharehodld tx equity create-company "Acme Corp" "ACME" 1000000 --from alice`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			symbol := args[1]
			totalShares, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid total shares: %w", err)
			}

			// Show what would be sent (full tx broadcast requires protobuf)
			fmt.Printf("Creating company listing:\n")
			fmt.Printf("  Creator: %s\n", clientCtx.GetFromAddress().String())
			fmt.Printf("  Name: %s\n", name)
			fmt.Printf("  Symbol: %s\n", symbol)
			fmt.Printf("  Total Shares: %d\n", totalShares)
			fmt.Println("\nNote: Full transaction broadcast requires protobuf message registration.")
			fmt.Println("The equity module keeper logic is implemented and functional.")

			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewTransferSharesCmd transfers shares between accounts
func NewTransferSharesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-shares [company-id] [class-id] [recipient] [shares]",
		Short: "Transfer shares to another account",
		Long: `Transfer equity shares from your account to another.

Note: Full transaction support requires protobuf generation.
For now, this command shows what would be sent.

Example:
  sharehodld tx equity transfer-shares 1 common hodl1abc... 100 --from alice`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			companyID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid company ID: %w", err)
			}

			classID := args[1]
			recipient := args[2]

			shares, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shares amount: %w", err)
			}

			fmt.Printf("Transferring shares:\n")
			fmt.Printf("  From: %s\n", clientCtx.GetFromAddress().String())
			fmt.Printf("  To: %s\n", recipient)
			fmt.Printf("  Company ID: %d\n", companyID)
			fmt.Printf("  Class ID: %s\n", classID)
			fmt.Printf("  Shares: %d\n", shares)
			fmt.Println("\nNote: Full transaction broadcast requires protobuf message registration.")
			fmt.Println("The equity module keeper logic is implemented and functional.")

			_ = math.NewInt(shares) // Verify math import is used

			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewIssueSharesCmd issues new shares
func NewIssueSharesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue-shares [company-id] [class-id] [recipient] [shares]",
		Short: "Issue new shares (company admin only)",
		Long: `Issue new equity shares to a recipient address.

Note: Full transaction support requires protobuf generation.
For now, this command shows what would be sent.

Example:
  sharehodld tx equity issue-shares 1 common hodl1abc... 1000 --from company-admin`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			companyID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid company ID: %w", err)
			}

			classID := args[1]
			recipient := args[2]

			shares, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shares amount: %w", err)
			}

			fmt.Printf("Issuing shares:\n")
			fmt.Printf("  Authority: %s\n", clientCtx.GetFromAddress().String())
			fmt.Printf("  Company ID: %d\n", companyID)
			fmt.Printf("  Class ID: %s\n", classID)
			fmt.Printf("  Recipient: %s\n", recipient)
			fmt.Printf("  Shares: %d\n", shares)
			fmt.Println("\nNote: Full transaction broadcast requires protobuf message registration.")
			fmt.Println("The equity module keeper logic is implemented and functional.")

			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
