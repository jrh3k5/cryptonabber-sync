package balance

import (
	"context"
	"math/big"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
)

// ERC20WrapperFetcher is a Fetcher implementation for ERC20Wrapper chains.
type ERC20WrapperFetcher struct {
	erc20Fetcher Fetcher[*config.ERC20Account]
}

// NewERC20WrapperFetcher creates a new ERC20WrapperFetcher.
func NewERC20WrapperFetcher(erc20Fetcher Fetcher[*config.ERC20Account]) *ERC20WrapperFetcher {
	return &ERC20WrapperFetcher{
		erc20Fetcher: erc20Fetcher,
	}
}

func (f *ERC20WrapperFetcher) FetchBalance(ctx context.Context, onchainAccount *config.ERC20WrapperAccount) (*big.Int, error) {
	return f.erc20Fetcher.FetchBalance(ctx, &onchainAccount.ERC20Account)
}
