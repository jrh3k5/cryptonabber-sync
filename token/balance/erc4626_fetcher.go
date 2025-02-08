package balance

import (
	"context"
	"fmt"
	"math/big"

	"github.com/jrh3k5/cryptonabber-sync/v2/config"
	"github.com/jrh3k5/cryptonabber-sync/v2/config/chain"
	rpcconfig "github.com/jrh3k5/cryptonabber-sync/v2/config/rpc"
	synchttp "github.com/jrh3k5/cryptonabber-sync/v2/http"
	"github.com/jrh3k5/cryptonabber-sync/v2/http/json/rpc"
)

type ERC4262Fetcher struct {
	rpcConfigurationResolver rpcconfig.ConfigurationResolver
	doer                     synchttp.Doer
}

func NewERC4262Fetcher(rpcConfigurationResolver rpcconfig.ConfigurationResolver, doer synchttp.Doer) *ERC4262Fetcher {
	return &ERC4262Fetcher{
		rpcConfigurationResolver: rpcConfigurationResolver,
		doer:                     doer,
	}
}

func (e *ERC4262Fetcher) FetchBalance(ctx context.Context, onchainAccount *config.ERC4626Account) (*big.Int, error) {
	rpcNodeURL, err := resolveRPCURL(ctx, e.rpcConfigurationResolver, onchainAccount.OnchainAsset, chain.TypeEVM)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve RPC URL: %w", err)
	}

	sharesResult, err := rpc.ExecuteEthCallAddress(ctx, e.doer, rpcNodeURL, "getShares", onchainAccount.VaultAddress, onchainAccount.WalletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to execute getShares: %w", err)
	}

	fmt.Printf("shares: %s\n", sharesResult)

	sharesBalance := big.NewInt(0)
	sharesBalance.SetString(sharesResult[2:], 16)

	fmt.Printf("sharesBalance: %v\n", sharesBalance.Text(10))

	assetsResult, err := rpc.ExecuteEthCallUint256(ctx, e.doer, rpcNodeURL, "convertToAssets", onchainAccount.VaultAddress, sharesBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to execute convertToAssets: %w", err)
	}

	fmt.Printf("assets: %s\n", assetsResult)

	assetsBalance := big.NewInt(0)
	assetsBalance.SetString(assetsResult[2:], 16)

	fmt.Printf("assetsBalance: %v\n", assetsBalance.Text(10))

	return assetsBalance, nil
}
