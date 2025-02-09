package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	synchttp "github.com/jrh3k5/cryptonabber-sync/v3/http"
)

type Request struct {
	ID      int64  `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type Response struct {
	ID      json.Number   `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Result  string        `json:"result"`
	Error   ResponseError `json:"error"`
}

type ResponseError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type RPCCallError struct {
	Code    int64
	Message string
}

func NewRPCCallError(code int64, message string) *RPCCallError {
	return &RPCCallError{
		Code:    code,
		Message: message,
	}
}

func (e *RPCCallError) Error() string {
	return fmt.Sprintf("RPC error: code %d, message: '%s'", e.Code, e.Message)
}

// ExecuteRequest executes the given RPC request and handles error checking.
// This can return RPCCallError if the RPC call returned an error in the JSON response.
func ExecuteRequest(ctx context.Context, doer synchttp.Doer, requestURL string, rpcRequest *Request) (*Response, error) {
	requestBodyBytes, err := json.Marshal(rpcRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	response, err := doer.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		statusErr := synchttp.BuildUnexpectedStatusErr(response)
		return nil, statusErr
	}

	var result *Response
	if unmarshalErr := json.NewDecoder(response.Body).Decode(&result); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", unmarshalErr)
	}

	if result.Error.Code != 0 {
		return nil, NewRPCCallError(result.Error.Code, result.Error.Message)
	}

	return result, nil
}
