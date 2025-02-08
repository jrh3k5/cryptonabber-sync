package token

import (
	"context"

	"github.com/jrh3k5/cryptonabber-sync/v2/config"
)

// AssetResolver is a function that resolves the address for the onchain asset that represents the value of the token.
type AssetResolver[M config.OnchainAccount] interface {
	// ResolveAssetAddress resolves the address for the onchain asset that represents the value of the token.
	// If this returns nil, it means it is for an asset that has no contract address on the asset's network.
	ResolveAssetAddress(ctx context.Context, onchainAccount M) (*string, error)
}
