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
	"github.com/jrh3k5/cryptonabber-sync/v2/token"
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
	rpcNodeURL, err := token.ResolveRPCURL(ctx, e.rpcConfigurationResolver, onchainAccount.OnchainAsset, chain.TypeEVM)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve RPC URL: %w", err)
	}

	sharesResult, err := rpc.ExecuteEthCall(ctx, e.doer, rpcNodeURL, onchainAccount.BalanceFunctionName, onchainAccount.VaultAddress, rpc.Arg("address", onchainAccount.WalletAddress))
	if err != nil {
		return nil, fmt.Errorf("failed to execute getShares: %w", err)
	}

	sharesBalance := big.NewInt(0)
	sharesBalance.SetString(sharesResult[2:], 16)

	assetsResult, err := rpc.ExecuteEthCall(ctx, e.doer, rpcNodeURL, "convertToAssets", onchainAccount.VaultAddress, rpc.Arg("uint256", sharesBalance))
	if err != nil {
		return nil, fmt.Errorf("failed to execute convertToAssets: %w", err)
	}

	assetsBalance := big.NewInt(0)
	assetsBalance.SetString(assetsResult[2:], 16)

	return assetsBalance, nil
}
