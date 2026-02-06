package upgrades

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// UpgradeName defines the on-chain upgrade name
const (
	// V1_0_0 is the initial release
	V1_0_0 = "v1.0.0"

	// V1_1_0 adds escrow and lending modules
	V1_1_0 = "v1.1.0"

	// V1_2_0 adds governance enhancements
	V1_2_0 = "v1.2.0"

	// V2_0_0 is the major upgrade with breaking changes
	V2_0_0 = "v2.0.0"
)

// Upgrade contains the upgrade info
type Upgrade struct {
	// UpgradeName is the name of the upgrade
	UpgradeName string

	// CreateUpgradeHandler creates the upgrade handler
	CreateUpgradeHandler func(mm *module.Manager, configurator module.Configurator) upgradetypes.UpgradeHandler

	// StoreUpgrades contains the store migrations
	StoreUpgrades storetypes.StoreUpgrades
}

// CreateV1_0_0UpgradeHandler creates upgrade handler for v1.0.0
// This is the initial release, no migrations needed
func CreateV1_0_0UpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		fmt.Println("Executing v1.0.0 upgrade...")
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

// CreateV1_1_0UpgradeHandler creates upgrade handler for v1.1.0
// This adds the escrow and lending modules
func CreateV1_1_0UpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		fmt.Println("Executing v1.1.0 upgrade...")
		fmt.Println("  - Adding escrow module")
		fmt.Println("  - Adding lending module")

		// Run migrations for new modules
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

// CreateV1_2_0UpgradeHandler creates upgrade handler for v1.2.0
// This adds governance enhancements (vote snapshots, execution)
func CreateV1_2_0UpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		fmt.Println("Executing v1.2.0 upgrade...")
		fmt.Println("  - Enhancing governance with vote snapshots")
		fmt.Println("  - Adding proposal execution queue")

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

// CreateV2_0_0UpgradeHandler creates upgrade handler for v2.0.0
// This is a major upgrade with breaking changes
func CreateV2_0_0UpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		fmt.Println("Executing v2.0.0 major upgrade...")
		fmt.Println("  - Protocol parameter updates")
		fmt.Println("  - State migrations")

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

// GetAllUpgrades returns all upgrade definitions
func GetAllUpgrades() []Upgrade {
	return []Upgrade{
		{
			UpgradeName:          V1_0_0,
			CreateUpgradeHandler: CreateV1_0_0UpgradeHandler,
			StoreUpgrades:        storetypes.StoreUpgrades{},
		},
		{
			UpgradeName:          V1_1_0,
			CreateUpgradeHandler: CreateV1_1_0UpgradeHandler,
			StoreUpgrades: storetypes.StoreUpgrades{
				Added: []string{"escrow", "lending"},
			},
		},
		{
			UpgradeName:          V1_2_0,
			CreateUpgradeHandler: CreateV1_2_0UpgradeHandler,
			StoreUpgrades: storetypes.StoreUpgrades{
				Added: []string{"governance"},
			},
		},
		{
			UpgradeName:          V2_0_0,
			CreateUpgradeHandler: CreateV2_0_0UpgradeHandler,
			StoreUpgrades:        storetypes.StoreUpgrades{},
		},
	}
}

// GetUpgrade returns the upgrade definition for a given name
func GetUpgrade(name string) (Upgrade, bool) {
	for _, upgrade := range GetAllUpgrades() {
		if upgrade.UpgradeName == name {
			return upgrade, true
		}
	}
	return Upgrade{}, false
}
