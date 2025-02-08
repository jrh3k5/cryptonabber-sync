package rpc

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jarcoal/httpmock"
)

const (
	lengthFunctionSelector   = 10
	lengthFunctionArgPadding = 24
)

type MockEVMNodeCallHandler func(functionSelector string, params []string) (MockEVMNodeRPCResult, *MockEVMNodeRPCError, error)

type MockEVMNode struct {
	handlers map[string]map[string]MockEVMNodeCallHandler // token address -> function selector -> handler
}

func StartMockEVMNode() *MockEVMNode {
	httpmock.Activate()

	evmNode := &MockEVMNode{
		handlers: make(map[string]map[string]MockEVMNodeCallHandler),
	}

	httpmock.RegisterResponder(http.MethodPost, evmNode.URL(), func(request *http.Request) (*http.Response, error) {
		requestBody := &Request{}
		if unmarshalErr := json.NewDecoder(request.Body).Decode(requestBody); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal request body: %w", unmarshalErr)
		}

		if requestBody.Method != "eth_call" {
			return nil, fmt.Errorf("unexpected request method: %s", requestBody.Method)
		}

		if len(requestBody.Params) != 2 {
			return nil, fmt.Errorf("unexpected number of request parameters: %d", len(requestBody.Params))
		}

		funcArgs := requestBody.Params[0].(map[string]any)
		targetAddress, targetAddressIsString := funcArgs["to"].(string)
		if !targetAddressIsString {
			return nil, fmt.Errorf("unexpected request parameter type: %T", funcArgs["to"])
		}

		contractHandlers, hasContract := evmNode.handlers[targetAddress]
		if !hasContract {
			return httpmock.NewJsonResponse(http.StatusOK, &Response{
				ID:      json.Number(fmt.Sprintf("%d", requestBody.ID)),
				JSONRPC: "2.0",
				Result:  "0x",
			})
		}

		data, dataIsString := funcArgs["data"].(string)
		if !dataIsString {
			return nil, fmt.Errorf("unexpected request parameter type: %T", funcArgs["data"])
		}

		functionSelector := data[0:10]
		handler, hasHandler := contractHandlers[functionSelector]
		if !hasHandler {
			// simulate a lack of the function being defined
			return httpmock.NewJsonResponse(http.StatusOK, &Response{
				ID:      json.Number(fmt.Sprintf("%d", requestBody.ID)),
				JSONRPC: "2.0",
				Error:   ResponseError{Code: -32000, Message: "execution reverted"},
			})
		}

		// It's a no-arg call
		var functionResult MockEVMNodeRPCResult
		var rpcError *MockEVMNodeRPCError
		if len(data) < lengthFunctionSelector+lengthFunctionArgPadding {
			var err error
			functionResult, rpcError, err = handler(data[0:10], nil)
			if err != nil {
				return nil, fmt.Errorf("failed to handle function call for selector '%s': %w", &functionSelector, err)
			}
		} else {
			parameter := data[lengthFunctionSelector+lengthFunctionArgPadding:]
			// for now, only support single-arg function calls
			var err error
			functionResult, rpcError, err = handler(data[0:10], []string{parameter})
			if err != nil {
				return nil, fmt.Errorf("failed to handle function call for selector '%s': %w", &functionSelector, err)
			}
		}

		response := &Response{
			ID:      json.Number(fmt.Sprintf("%d", requestBody.ID)),
			JSONRPC: requestBody.JSONRPC,
			Result:  functionResult.ReturnValue(),
		}

		if rpcError != nil {
			response.Error = ResponseError{
				Code:    rpcError.Code,
				Message: rpcError.Message,
			}
		}

		return httpmock.NewJsonResponse(http.StatusOK, response)
	})

	return evmNode
}

func (m *MockEVMNode) RegisterContractExistence(address string) {
	m.handlers[address] = make(map[string]MockEVMNodeCallHandler)
}

func (m *MockEVMNode) RegisterFunctionCall(
	functionName string,
	targetAddress string,
	parameterTypes []string,
	callHandler MockEVMNodeCallHandler,
) error {
	functionSignature := fmt.Sprintf("%s(%s)", functionName, strings.Join(parameterTypes, ","))
	functionSelector := crypto.Keccak256Hash(([]byte(functionSignature))).String()[0:10]

	contractHandlers, hasContract := m.handlers[targetAddress]
	if !hasContract {
		contractHandlers = make(map[string]MockEVMNodeCallHandler)
		m.handlers[targetAddress] = contractHandlers
	}

	contractHandlers[functionSelector] = callHandler

	return nil
}

func (m *MockEVMNode) Stop() {
	httpmock.DeactivateAndReset()
}

func (m *MockEVMNode) URL() string {
	return "http://node.localhost/rpc"
}

type MockEVMNodeRPCError struct {
	Code    int64
	Message string
}

type MockEVMNodeRPCResult interface {
	ReturnValue() string
}

type MockEVMNodeRPCAddressResult struct {
	Address string
}

func NewMockEVMNodeRPCAddressResult(address string) *MockEVMNodeRPCAddressResult {
	return &MockEVMNodeRPCAddressResult{Address: address}
}

func (a MockEVMNodeRPCAddressResult) ReturnValue() string {
	return "0x000000000000000000000000" + strings.TrimPrefix(a.Address, "0x")
}

type MockEVMNodeRPCNumericResult struct {
	Number *big.Int
}

func NewMockEVMNodeRPCNumericResult(number *big.Int) *MockEVMNodeRPCNumericResult {
	return &MockEVMNodeRPCNumericResult{Number: number}
}

func (n MockEVMNodeRPCNumericResult) ReturnValue() string {
	return fmt.Sprintf("0x%x", n.Number)
}
