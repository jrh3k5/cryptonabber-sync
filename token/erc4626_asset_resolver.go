package token

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
	"github.com/jrh3k5/cryptonabber-sync/v3/config/chain"
	rpcconfig "github.com/jrh3k5/cryptonabber-sync/v3/config/rpc"
	synchttp "github.com/jrh3k5/cryptonabber-sync/v3/http"
	"github.com/jrh3k5/cryptonabber-sync/v3/http/json/rpc"
)

type ERC4626AssetResolver struct {
	rpcConfigurationResolver rpcconfig.ConfigurationResolver
	doer                     synchttp.Doer
}

func NewERC4626AssetResolver(rpcConfigurationResolver rpcconfig.ConfigurationResolver, doer synchttp.Doer) *ERC4626AssetResolver {
	return &ERC4626AssetResolver{
		rpcConfigurationResolver: rpcConfigurationResolver,
		doer:                     doer,
	}
}

func (r *ERC4626AssetResolver) ResolveAssetAddress(ctx context.Context, onchainAccount *config.ERC4626Account) (*string, error) {
	if onchainAccount == nil {
		return nil, errors.New("onchain account must be provided")
	}

	if backingAsset := onchainAccount.BackingAsset; backingAsset != nil {
		return backingAsset.ContractAddress, nil
	}

	if onchainAccount.VaultAddress == "" {
		return nil, errors.New("vault address must be provided on the given onchain account")
	}

	nodeURL, err := ResolveRPCURL(ctx, r.rpcConfigurationResolver, onchainAccount.OnchainAsset, chain.TypeEVM)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve node URL for asset %s: %w", onchainAccount.OnchainAsset, err)
	}

	assetAddress, err := rpc.ExecuteEthCall(ctx, r.doer, nodeURL, "asset", onchainAccount.VaultAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve asset for ERC4626 account: %w", err)
	}

	substringAddress := strings.ReplaceAll(assetAddress, "000000000000000000000000", "")

	return &substringAddress, nil
}
