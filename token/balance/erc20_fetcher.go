package balance

import (
	"context"
	"fmt"
	"math/big"

	synchttp "github.com/jrh3k5/cryptonabber-sync/http"
)

// ERC20Fetcher is a Fetcher implementation for EVM chains.
type ERC20Fetcher struct {
	nodeURL string
	doer    synchttp.Doer
}

// NewERC20Fetcher builds an ERC20Fetcher instance that communicates with the given node URL.
func NewERC20Fetcher(nodeURL string, doer synchttp.Doer) *ERC20Fetcher {
	return &ERC20Fetcher{
		nodeURL: nodeURL,
		doer:    doer,
	}
}

func (e *ERC20Fetcher) FetchBalance(ctx context.Context, tokenAddress string, walletAddress string) (*big.Int, error) {
	result, err := executeEthCall(ctx, e.doer, e.nodeURL, "balanceOf", tokenAddress, walletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to execute balanceOf: %w", err)
	}

	balance := big.NewInt(0)
	balance.SetString(result[2:], 16)
	return balance, nil
}
