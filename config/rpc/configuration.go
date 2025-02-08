package rpc

import "github.com/jrh3k5/cryptonabber-sync/v2/config/chain"

// Configuration describes how to communciate with an RPC node.
type Configuration struct {
	RPCURL    string     // the URL of the RPC node
	ChainName string     // the name of the chain
	ChainType chain.Type // the type of the chain
}
