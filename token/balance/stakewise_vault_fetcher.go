package balance

import (
	"context"
	"fmt"
	"math/big"

	synchttp "github.com/jrh3k5/cryptonabber-sync/http"
)

// StakewiseVaultFetcher will fetch a wallet's balance from a Stakewise vault.
type StakewiseVaultFetcher struct {
	nodeURL string
	doer    synchttp.Doer
}

func NewStakewiseVaultFetcher(nodeURL string, doer synchttp.Doer) *StakewiseVaultFetcher {
	return &StakewiseVaultFetcher{
		nodeURL: nodeURL,
		doer:    doer,
	}
}

func (s *StakewiseVaultFetcher) FetchBalance(ctx context.Context, tokenAddress string, walletAddress string) (*big.Int, error) {
	result, err := executeEthCall(ctx, s.doer, s.nodeURL, "getShares", tokenAddress, walletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to execute balanceOf: %w", err)
	}

	balance := big.NewInt(0)
	balance.SetString(result[2:], 16)
	return balance, nil
}
