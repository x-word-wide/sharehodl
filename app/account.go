package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountRetriever implements the Account and AccountRetriever interfaces.
type AccountRetriever struct{}

func (AccountRetriever) GetAccount(_ client.Context, _ sdk.AccAddress) (client.Account, error) {
	return nil, nil
}

func (AccountRetriever) GetAccountWithHeight(clientCtx client.Context, addr sdk.AccAddress) (client.Account, int64, error) {
	return nil, 0, nil
}

func (AccountRetriever) EnsureExists(_ client.Context, _ sdk.AccAddress) error {
	return nil
}

func (AccountRetriever) GetAccountNumberSequence(_ client.Context, _ sdk.AccAddress) (accNum uint64, accSeq uint64, err error) {
	return 0, 0, nil
}

func (a AccountRetriever) GetAddress() sdk.AccAddress  { return nil }
func (a AccountRetriever) GetPubKey() cryptotypes.PubKey { return nil }
func (a AccountRetriever) GetAccountNumber() uint64     { return 0 }
func (a AccountRetriever) GetSequence() uint64          { return 0 }