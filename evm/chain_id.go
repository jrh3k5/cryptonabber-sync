package evm

import (
	"context"
	"fmt"
	"math/big"

	synchttp "github.com/jrh3k5/cryptonabber-sync/http"
	"github.com/jrh3k5/cryptonabber-sync/http/json/rpc"
)

// ChainIDFetcher describes a means of retrieving a chain ID.
type ChainIDFetcher interface {
	// GetChainID gets the chain ID.
	GetChainID(ctx context.Context) (*big.Int, error)
}

// JSONRPCChainIDFetcher is a ChainIDFetcher that uses JSON RPC calls
// to determine it.
type JSONRPCChainIDFetcher struct {
	nodeURL string
	doer    synchttp.Doer
}

func NewJSONRPCChainIDFetcher(nodeURL string, doer synchttp.Doer) *JSONRPCChainIDFetcher {
	return &JSONRPCChainIDFetcher{
		nodeURL: nodeURL,
		doer:    doer,
	}
}

func (j *JSONRPCChainIDFetcher) GetChainID(ctx context.Context) (*big.Int, error) {
	rpcRequest := &rpc.Request{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "eth_chainId",
	}

	rpcResponse, err := rpc.ExecuteRequest(ctx, j.doer, j.nodeURL, rpcRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to execute eth_chainId: %w", err)
	}

	result := rpcResponse.Result
	chainID := big.NewInt(0)
	chainID.SetString(result[2:], 16)
	return chainID, nil
}
