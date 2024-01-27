package balance

import (
	"context"
	"math/big"
)

type Fetcher interface {
	FetchBalance(ctx context.Context, tokenAddress string, walletAddress string) (*big.Int, error)
}
