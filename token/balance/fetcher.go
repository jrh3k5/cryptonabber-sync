package balance

import (
	"context"
	"math/big"
)

// Fetcher defines a way to fetch balance information.
type Fetcher interface {
	// FetchBalance gets the balance of the given token for the given wallet address.
	FetchBalance(ctx context.Context, tokenAddress string, walletAddress string) (*big.Int, error)
}
