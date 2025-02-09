package token

import (
	"context"
	"fmt"
	"math/big"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
	"github.com/jrh3k5/cryptonabber-sync/v3/config/chain"
	rpcconfig "github.com/jrh3k5/cryptonabber-sync/v3/config/rpc"
	synchttp "github.com/jrh3k5/cryptonabber-sync/v3/http"
	"github.com/jrh3k5/cryptonabber-sync/v3/http/json/rpc"
)

// RPCDecimalsResolver resolves the decimals of tokens using RPC calls.
type RPCDecimalsResolver struct {
	rpcConfigurationResolver rpcconfig.ConfigurationResolver
	doer                     synchttp.Doer
}

// NewRPCDecimalsResolver builds an RPCDecimalsResolver.
func NewRPCDecimalsResolver(rpcConfigurationResolver rpcconfig.ConfigurationResolver, doer synchttp.Doer) *RPCDecimalsResolver {
	return &RPCDecimalsResolver{
		rpcConfigurationResolver: rpcConfigurationResolver,
		doer:                     doer,
	}
}

func (r *RPCDecimalsResolver) ResolveDecimals(ctx context.Context, onchainAsset config.OnchainAsset, tokenAddress *string) (int, error) {
	if tokenAddress == nil {
		// Hard-code support for ETH
		return 18, nil
	}

	rpcURL, err := ResolveRPCURL(ctx, r.rpcConfigurationResolver, onchainAsset, chain.TypeEVM)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve RPC URL: %w", err)
	}

	result, err := rpc.ExecuteEthCall(ctx, r.doer, rpcURL, "decimals", *tokenAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve decimals: %w", err)
	}

	decimalsBigInt := new(big.Int)
	decimalsBigInt.SetString(result[2:], 16)

	return int(decimalsBigInt.Int64()), nil
}
