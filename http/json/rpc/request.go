package rpc

import "encoding/json"

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
