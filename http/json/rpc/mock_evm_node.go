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

// MockEVMNodeRPCResult is the result of an RPC call.
type MockEVMRPCMethodCallHandler func(methodName string) (MockEVMNodeRPCResult, *MockEVMNodeRPCError, error)

// MockEVMNodeETHCallCallHandler is a function that handles a function call.
type MockEVMNodeETHCallCallHandler func(functionSelector string, params []string) (MockEVMNodeRPCResult, *MockEVMNodeRPCError, error)

// MockEVMNode is a mock EVM node.
type MockEVMNode struct {
	methodCallHandles map[string]MockEVMRPCMethodCallHandler              // method name -> handler
	ethCallHandlers   map[string]map[string]MockEVMNodeETHCallCallHandler // token address -> function selector -> handler
}

// StartMockEVMNode starts a mock EVM node.
func StartMockEVMNode() *MockEVMNode {
	httpmock.Activate()

	evmNode := &MockEVMNode{
		methodCallHandles: make(map[string]MockEVMRPCMethodCallHandler),
		ethCallHandlers:   make(map[string]map[string]MockEVMNodeETHCallCallHandler),
	}

	httpmock.RegisterResponder(http.MethodPost, evmNode.URL(), func(request *http.Request) (*http.Response, error) {
		requestBody := &Request{}
		if unmarshalErr := json.NewDecoder(request.Body).Decode(requestBody); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal request body: %w", unmarshalErr)
		}

		methodHandler, hasMethodHandler := evmNode.methodCallHandles[requestBody.Method]
		if hasMethodHandler {
			methodResult, rpcError, err := methodHandler(requestBody.Method)
			if err != nil {
				return nil, fmt.Errorf("failed to handle method call: %w", err)
			}

			response := &Response{
				ID:      json.Number(fmt.Sprintf("%d", requestBody.ID)),
				JSONRPC: requestBody.JSONRPC,
				Result:  methodResult.ReturnValue(),
			}

			if rpcError != nil {
				response.Error = ResponseError{
					Code:    rpcError.Code,
					Message: rpcError.Message,
				}
			}

			return httpmock.NewJsonResponse(http.StatusOK, response)
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

		contractHandlers, hasContract := evmNode.ethCallHandlers[targetAddress]
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

// RegisterContractExistence registers the existence of a contract.
func (m *MockEVMNode) RegisterContractExistence(address string) {
	m.ethCallHandlers[address] = make(map[string]MockEVMNodeETHCallCallHandler)
}

// RegisterRPCMethodCall registers a function call handler for calls to RPC methods.
// For eth_call, use RegisterETHCallCall.
func (m *MockEVMNode) RegisterRPCMethodCall(
	methodName string,
	callHandler MockEVMRPCMethodCallHandler,
) error {
	m.methodCallHandles[methodName] = callHandler

	return nil
}

// RegisterETHCallCall registers a function call handler for calls to functions using eth_call.
func (m *MockEVMNode) RegisterETHCallCall(
	functionName string,
	targetAddress string,
	parameterTypes []string,
	callHandler MockEVMNodeETHCallCallHandler,
) error {
	functionSignature := fmt.Sprintf("%s(%s)", functionName, strings.Join(parameterTypes, ","))
	functionSelector := crypto.Keccak256Hash(([]byte(functionSignature))).String()[0:10]

	contractHandlers, hasContract := m.ethCallHandlers[targetAddress]
	if !hasContract {
		contractHandlers = make(map[string]MockEVMNodeETHCallCallHandler)
		m.ethCallHandlers[targetAddress] = contractHandlers
	}

	contractHandlers[functionSelector] = callHandler

	return nil
}

// Stop stops the mock EVM node.
func (m *MockEVMNode) Stop() {
	httpmock.DeactivateAndReset()
}

// URL returns the URL of the mock EVM node.
func (m *MockEVMNode) URL() string {
	return "http://node.localhost/rpc"
}

// MockEVMNodeRPCError is the error result of an RPC call.
type MockEVMNodeRPCError struct {
	Code    int64
	Message string
}

// MockEVMNodeRPCResult is the result of an RPC call.
type MockEVMNodeRPCResult interface {
	// ReturnValue returns the value of the result.
	ReturnValue() string
}

// MockEVMNodeRPCAddressResult is the result of an RPC call that returns an address.
type MockEVMNodeRPCAddressResult struct {
	Address string
}

// NewMockEVMNodeRPCAddressResult builds a MockEVMNodeRPCAddressResult instance.
func NewMockEVMNodeRPCAddressResult(address string) *MockEVMNodeRPCAddressResult {
	return &MockEVMNodeRPCAddressResult{Address: address}
}

func (a MockEVMNodeRPCAddressResult) ReturnValue() string {
	return "0x000000000000000000000000" + strings.TrimPrefix(a.Address, "0x")
}

// MockEVMNodeRPCNumericResult is the result of an RPC call that returns a numeric value.
type MockEVMNodeRPCNumericResult struct {
	Number *big.Int
}

// NewMockEVMNodeRPCNumericResult builds a MockEVMNodeRPCNumericResult instance.
func NewMockEVMNodeRPCNumericResult(number *big.Int) *MockEVMNodeRPCNumericResult {
	return &MockEVMNodeRPCNumericResult{Number: number}
}

func (n MockEVMNodeRPCNumericResult) ReturnValue() string {
	return fmt.Sprintf("0x%x", n.Number)
}
