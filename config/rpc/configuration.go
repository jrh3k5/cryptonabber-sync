package rpc

import "github.com/jrh3k5/cryptonabber-sync/v2/config/chain"

// Configuration describes how to communciate with an RPC node.
type Configuration struct {
	RPCURL    string     `yaml:"rpc_url"`    // the URL of the RPC node
	ChainName string     `yaml:"chain_name"` // the name of the chain
	ChainType chain.Type `yaml:"chain_type"` // the type of the chain
}
