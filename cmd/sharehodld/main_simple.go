package main

import (
	"fmt"
	"os"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/sharehodl/sharehodl-blockchain/app"
)

func main() {
	fmt.Println("ShareHODL Blockchain v1.0.0")
	fmt.Println("🏦 Building the Future of Global Equity Markets")
	fmt.Println("")

	// Test that our app can be created successfully
	logger := log.NewLogger(os.Stdout)
	db := dbm.NewMemDB()
	
	sharehodlApp := app.NewShareHODLApp(
		logger, 
		db, 
		nil, 
		true, 
		servertypes.AppOptions{},
	)
	
	fmt.Printf("✅ ShareHODL blockchain initialized successfully!\n")
	fmt.Printf("   App Name: %s\n", sharehodlApp.Name())
	fmt.Printf("   Ready for custom modules implementation\n")
	fmt.Printf("\n🚀 Next: Implementing HODL stablecoin module...\n")
}