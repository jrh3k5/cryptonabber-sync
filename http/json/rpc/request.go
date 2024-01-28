package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	ID      int64  `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type Response struct {
	ID      json.Number `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Result  string      `json:"result"`
	Error   struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// ExecuteRequest executes the given RPC request and handles error checking.
func ExecuteRequest(ctx context.Context, requestURL string, rpcRequest *Request) (*Response, error) {
	requestBodyBytes, err := json.Marshal(rpcRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(requestBodyBytes))
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
			bodyText = fmt.Sprintf("failed to read request body: %v", bodyBytesErr)
		} else {
			bodyText = string(bodyBytes)
		}

		return nil, fmt.Errorf("unexpected response status code (%d); first %d bytes of body are: '%s'", response.StatusCode, bodySampleLimit, bodyText)
	}

	var result *Response
	if unmarshalErr := json.NewDecoder(response.Body).Decode(&result); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", unmarshalErr)
	}

	if result.Error.Code != 0 {
		return nil, fmt.Errorf("RPC error: code %d, message: '%s'", result.Error.Code, result.Error.Message)
	}

	return result, nil
}
