package balance

import (
	"context"
	"math/big"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
)

// Fetcher defines a way to fetch balance information.
type Fetcher[M config.OnchainAccount] interface {
	// FetchBalance gets the balance of the given token for the given wallet address.
	// If the given onchain account represents a wrapper around another asset (a wrapping asset, a vault),
	// the returned balance is the balance of the underlying asset.
	FetchBalance(ctx context.Context, onchainAccount M) (*big.Int, error)
}
