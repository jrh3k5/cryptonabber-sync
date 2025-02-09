package token

import (
	"context"
	"fmt"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
	"github.com/jrh3k5/cryptonabber-sync/v3/config/chain"
	"github.com/jrh3k5/cryptonabber-sync/v3/config/rpc"
)

// ResolveRPCURL resolves the RPC URL for the given chain name.
func ResolveRPCURL(ctx context.Context, configurationResolver rpc.ConfigurationResolver, onchainAsset config.OnchainAsset, requiredChainType chain.Type) (string, error) {
	if onchainAsset.ChainName == "" {
		return "", fmt.Errorf("chain name is required")
	}

	rpcConfig, hasConfig, err := configurationResolver.ResolveConfiguration(ctx, onchainAsset.ChainName)
	if err != nil {
		return "", fmt.Errorf("failed to resolve RPC configuration: %w", err)
	} else if !hasConfig {
		return "", fmt.Errorf("no RPC configuration found for chain '%s'", onchainAsset.ChainName)
	} else if rpcConfig.ChainType != requiredChainType {
		return "", fmt.Errorf("RPC configuration for chain '%s' is not the required chain type of '%s'", onchainAsset.ChainName, requiredChainType)
	}

	return rpcConfig.RPCURL, nil
}
