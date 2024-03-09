package balance

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	synchttp "github.com/jrh3k5/cryptonabber-sync/http"
	"github.com/jrh3k5/cryptonabber-sync/http/json/rpc"
)

// executeEthCallAddress calls the given method (assuming to accept a single address - the given wallet address) against
// the given contract address.
func executeEthCallAddress(ctx context.Context, doer synchttp.Doer, nodeURL string, methodName string, contractAddress string, walletAddress string) (string, error) {
	data := crypto.Keccak256Hash([]byte(methodName + "(address)")).String()[0:10] + "000000000000000000000000" + walletAddress[2:]

	rpcRequest := &rpc.Request{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "eth_call",
		Params: []any{
			map[string]string{
				"to":   contractAddress,
				"data": data,
			},
			"latest",
		},
	}

	result, err := rpc.ExecuteRequest(ctx, doer, nodeURL, rpcRequest)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}

	return result.Result, nil
}

// executeEthCallUint256 calls the given method, supplying the given input as uin256 input. The method called is assumed
// to return a value.
func executeEthCallUint256(ctx context.Context, doer synchttp.Doer, nodeURL string, methodName string, contractAddress string, input int64) (string, error) {
	data := crypto.Keccak256Hash([]byte(methodName + "(uint256)")).String()[0:10] + "000000000000000000000000" + fmt.Sprintf("%048x", input)

	rpcRequest := &rpc.Request{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "eth_call",
		Params: []any{
			map[string]string{
				"to":   contractAddress,
				"data": data,
			},
			"latest",
		},
	}

	result, err := rpc.ExecuteRequest(ctx, doer, nodeURL, rpcRequest)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}

	return result.Result, nil
}
