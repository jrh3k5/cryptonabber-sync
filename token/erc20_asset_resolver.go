package token

import (
	"context"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
)

type ERC20AssetResolver struct {
}

func NewERC20AssetResolver() *ERC20AssetResolver {
	return &ERC20AssetResolver{}
}

func (r *ERC20AssetResolver) ResolveAssetAddress(_ context.Context, onchainAccount *config.ERC20Account) (*string, error) {
	return &onchainAccount.TokenAddress, nil
}
