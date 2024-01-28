package balance

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	synchttp "github.com/jrh3k5/cryptonabber-sync/http"
	"github.com/jrh3k5/cryptonabber-sync/http/json/rpc"
)

// ERC20Fetcher is a Fetcher implementation for EVM chains.
type ERC20Fetcher struct {
	nodeURL string
	doer    synchttp.Doer
}

// NewERC20Fetcher builds an ERC20Fetcher instance that communicates with the given node URL.
func NewERC20Fetcher(nodeURL string, doer synchttp.Doer) *ERC20Fetcher {
	return &ERC20Fetcher{
		nodeURL: nodeURL,
		doer:    doer,
	}
}

func (e *ERC20Fetcher) FetchBalance(ctx context.Context, tokenAddress string, walletAddress string) (*big.Int, error) {
	data := crypto.Keccak256Hash([]byte("balanceOf(address)")).String()[0:10] + "000000000000000000000000" + walletAddress[2:]

	rpcRequest := &rpc.Request{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "eth_call",
		Params: []any{
			map[string]string{
				"to":   tokenAddress,
				"data": data,
			},
			"latest",
		},
	}

	result, err := rpc.ExecuteRequest(ctx, e.doer, e.nodeURL, rpcRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	balance := big.NewInt(0)
	balance.SetString(result.Result[2:], 16)
	return balance, nil
}
