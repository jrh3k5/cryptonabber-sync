package rpc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	synchttp "github.com/jrh3k5/cryptonabber-sync/v2/http"
)

type EthCallArgument struct {
	Type  string
	Value any
}

func Arg(argType string, v any) *EthCallArgument {
	return &EthCallArgument{
		Type:  argType,
		Value: v,
	}
}

// ExecuteEthCallNoArg calls the given method (assuming to accept no arguments) against the given contract address.
func ExecuteEthCall(
	ctx context.Context,
	doer synchttp.Doer,
	nodeURL string,
	methodName string,
	contractAddress string,
	args ...*EthCallArgument,
) (string, error) {
	if len(args) > 1 {
		return "", errors.New("only one argument is supported for eth_call calls")
	}

	argTypes := make([]string, len(args))
	argValues := make([]string, len(args))

	for i, arg := range args {
		argTypes[i] = arg.Type
		switch v := arg.Value.(type) {
		case string:
			if argTypes[i] != "address" {
				return "", errors.New("only address arguments are supported for eth_call string parameters")
			}
			argValues[i] = v[2:]
		case *big.Int:
			argValues[i] = fmt.Sprintf("%040x", v)
		}
	}

	data := crypto.Keccak256Hash([]byte(methodName + "(" + strings.Join(argTypes, ",") + ")")).String()[0:10]
	if len(args) > 0 {
		data += "000000000000000000000000" + argValues[0]
	}

	rpcRequest := &Request{
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

	result, err := ExecuteRequest(ctx, doer, nodeURL, rpcRequest)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}

	return result.Result, nil
}
