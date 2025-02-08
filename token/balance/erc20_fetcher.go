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

// ERC20Fetcher is a Fetcher implementation for EVM chains.
type ERC20Fetcher struct {
	rpcConfigurationResolver rpcconfig.ConfigurationResolver
	doer                     synchttp.Doer
}

// NewERC20Fetcher builds an ERC20Fetcher instance that communicates with the given node URL.
func NewERC20Fetcher(rpcConfigurationResolver rpcconfig.ConfigurationResolver, doer synchttp.Doer) *ERC20Fetcher {
	return &ERC20Fetcher{
		rpcConfigurationResolver: rpcConfigurationResolver,
		doer:                     doer,
	}
}

func (e *ERC20Fetcher) FetchBalance(ctx context.Context, onchainAccount *config.ERC20Account) (*big.Int, error) {
	rpcURL, err := resolveRPCURL(ctx, e.rpcConfigurationResolver, onchainAccount.OnchainAsset, chain.TypeEVM)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve RPC URL: %w", err)
	}

	result, err := rpc.ExecuteEthCallAddress(ctx, e.doer, rpcURL, "balanceOf", onchainAccount.TokenAddress, onchainAccount.WalletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to execute balanceOf: %w", err)
	}

	balance := big.NewInt(0)
	balance.SetString(result[2:], 16)

	return balance, nil
}
