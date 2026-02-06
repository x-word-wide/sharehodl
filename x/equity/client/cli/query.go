package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for the equity module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "equity",
		Short:                      "Querying commands for the equity module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryCompany(),
		GetCmdQueryCompanies(),
		GetCmdQueryShareholdings(),
	)

	return cmd
}

// GetCmdQueryCompany returns the command to query a company by ID
func GetCmdQueryCompany() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "company [company-id]",
		Short: "Query a company by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			companyID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid company ID: %w", err)
			}

			// For now, just display a message since we don't have full gRPC support
			fmt.Printf("Querying company with ID: %d\n", companyID)
			fmt.Printf("Note: Full gRPC query support requires protobuf generation\n")
			fmt.Printf("Use REST API at: http://localhost:1317/sharehodl/equity/v1/company/%d\n", companyID)

			_ = clientCtx // Will be used when gRPC is fully implemented
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryCompanies returns the command to query all companies
func GetCmdQueryCompanies() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "companies",
		Short: "Query all companies",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			fmt.Println("Querying all companies...")
			fmt.Println("Note: Full gRPC query support requires protobuf generation")
			fmt.Println("Use REST API at: http://localhost:1317/sharehodl/equity/v1/companies")

			_ = clientCtx
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryShareholdings returns the command to query shareholdings
func GetCmdQueryShareholdings() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shareholdings [owner-address]",
		Short: "Query shareholdings by owner address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			owner := args[0]
			fmt.Printf("Querying shareholdings for: %s\n", owner)
			fmt.Println("Note: Full gRPC query support requires protobuf generation")
			fmt.Printf("Use REST API at: http://localhost:1317/sharehodl/equity/v1/shareholdings/%s\n", owner)

			_ = clientCtx
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
