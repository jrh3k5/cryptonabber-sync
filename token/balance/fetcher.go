package balance

import (
	"context"
	"math/big"

	"github.com/jrh3k5/cryptonabber-sync/v2/config"
)

// Fetcher defines a way to fetch balance information.
type Fetcher[M config.OnchainAccount] interface {
	// FetchBalance gets the balance of the given token for the given wallet address.
	FetchBalance(ctx context.Context, onchainAccount M) (*big.Int, error)
}
