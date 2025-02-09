package token

import (
	"context"
	"fmt"
	"strings"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
	"github.com/jrh3k5/cryptonabber-sync/v3/config/chain"
	rpcconfig "github.com/jrh3k5/cryptonabber-sync/v3/config/rpc"
	synchttp "github.com/jrh3k5/cryptonabber-sync/v3/http"
	"github.com/jrh3k5/cryptonabber-sync/v3/http/json/rpc"
)

type ERC20WrapperAssetResolver struct {
	rpcConfigurationResolver rpcconfig.ConfigurationResolver
	doer                     synchttp.Doer
}

func NewERC20WrapperAssetResolver(rpcConfigurationResolver rpcconfig.ConfigurationResolver, doer synchttp.Doer) *ERC20WrapperAssetResolver {
	return &ERC20WrapperAssetResolver{
		rpcConfigurationResolver: rpcConfigurationResolver,
		doer:                     doer,
	}
}

func (e *ERC20WrapperAssetResolver) ResolveAssetAddress(ctx context.Context, onchainAccount *config.ERC20WrapperAccount) (*string, error) {
	rpcURL, err := ResolveRPCURL(ctx, e.rpcConfigurationResolver, onchainAccount.OnchainAsset, chain.TypeEVM)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve RPC URL: %w", err)
	}

	assetAddress, err := rpc.ExecuteEthCall(ctx, e.doer, rpcURL, onchainAccount.BaseTokenAddressFunction, onchainAccount.TokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to execute %s: %w", onchainAccount.BaseTokenAddressFunction, err)
	}

	substringAddress := strings.ReplaceAll(assetAddress, "000000000000000000000000", "")

	return &substringAddress, nil
}
