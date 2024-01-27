package balance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
)

type EVMFetcher struct {
	nodeURL string
}

func NewEVMFetcher(nodeURL string) *EVMFetcher {
	return &EVMFetcher{
		nodeURL: nodeURL,
	}
}

func (e *EVMFetcher) FetchBalance(ctx context.Context, tokenAddress string, walletAddress string) (*big.Int, error) {
	data := crypto.Keccak256Hash([]byte("balanceOf(address)")).String()[0:10] + "000000000000000000000000" + walletAddress[2:]

	postBody, err := json.Marshal(map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "eth_call",
		"params": []interface{}{
			map[string]string{
				"to":   tokenAddress,
				"data": data,
			},
			"latest",
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, e.nodeURL, bytes.NewBuffer(postBody))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		bodySampleLimit := int64(200)

		var bodyText string
		bodyBytes, bodyBytesErr := io.ReadAll(io.LimitReader(response.Body, bodySampleLimit))
		if bodyBytesErr != nil {
			bodyText = fmt.Sprintf("failed to read request body: %w", bodyBytesErr)
		} else {
			bodyText = string(bodyBytes)
		}

		return nil, fmt.Errorf("unexpected response status code (%d); first %d bytes of body are: '%s'", response.StatusCode, bodySampleLimit, bodyText)
	}

	var result *ethRPCResult
	if unmarshalErr := json.NewDecoder(response.Body).Decode(&result); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", unmarshalErr)
	}

	if result.Error.Code != 0 {
		return nil, fmt.Errorf("RPC error: code %d, message: '%s'", result.Error.Code, result.Error.Message)
	}

	balance := big.NewInt(0)
	balance.SetString(result.Result[2:], 16)
	return balance, nil
}

type ethRPCResult struct {
	Result string `json:"result"`
	Error  struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
