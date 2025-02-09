package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/jrh3k5/cryptonabber-sync/v3/config/chain"
	rpcconfig "github.com/jrh3k5/cryptonabber-sync/v3/config/rpc"
	synchttp "github.com/jrh3k5/cryptonabber-sync/v3/http"
	"github.com/jrh3k5/cryptonabber-sync/v3/http/json/rpc"
)

// ChainIDFetcher describes a means of retrieving a chain ID.
type ChainIDFetcher interface {
	// GetChainID gets the chain ID.
	GetChainID(ctx context.Context) (*big.Int, error)
}

// JSONRPCChainIDFetcher is a ChainIDFetcher that uses JSON RPC calls
// to determine it.
type JSONRPCChainIDFetcher struct {
	rpcConfigurationResolver rpcconfig.ConfigurationResolver
	doer                     synchttp.Doer
}

func NewJSONRPCChainIDFetcher(rpcConfigurationResolver rpcconfig.ConfigurationResolver, doer synchttp.Doer) *JSONRPCChainIDFetcher {
	return &JSONRPCChainIDFetcher{
		rpcConfigurationResolver: rpcConfigurationResolver,
		doer:                     doer,
	}
}

func (j *JSONRPCChainIDFetcher) GetChainID(ctx context.Context, chainName string) (*big.Int, error) {
	rpcConfiguration, hasURL, err := j.rpcConfigurationResolver.ResolveConfiguration(ctx, chainName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve RPC URL: %w", err)
	} else if !hasURL {
		return nil, fmt.Errorf("no RPC URL found for chain '%s'", chainName)
	} else if rpcConfiguration.ChainType != chain.TypeEVM {
		return nil, fmt.Errorf("invalid chain type for chain '%s': %s", chainName, rpcConfiguration.ChainType)
	}

	rpcRequest := &rpc.Request{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "eth_chainId",
	}

	rpcResponse, err := rpc.ExecuteRequest(ctx, j.doer, rpcConfiguration.RPCURL, rpcRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to execute eth_chainId: %w", err)
	}

	result := rpcResponse.Result
	chainID := big.NewInt(0)
	chainID.SetString(result[2:], 16)
	return chainID, nil
}
