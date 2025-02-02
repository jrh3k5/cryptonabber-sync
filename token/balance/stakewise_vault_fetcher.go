package balance

import (
	"context"
	"fmt"
	"math/big"

	synchttp "github.com/jrh3k5/cryptonabber-sync/v2/http"
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
	sharesResult, err := executeEthCallAddress(ctx, s.doer, s.nodeURL, "getShares", tokenAddress, walletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to execute getShares: %w", err)
	}

	sharesBalance := big.NewInt(0)
	sharesBalance.SetString(sharesResult[2:], 16)

	assetsResult, err := executeEthCallUint256(ctx, s.doer, s.nodeURL, "convertToAssets", tokenAddress, sharesBalance.Int64())
	if err != nil {
		return nil, fmt.Errorf("failed to execute convertToAssets: %w", err)
	}

	assetsBalance := big.NewInt(0)
	assetsBalance.SetString(assetsResult[2:], 16)

	return assetsBalance, nil
}
