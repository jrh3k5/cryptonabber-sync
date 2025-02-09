package evm_test

import (
	"testing"

	"github.com/jrh3k5/cryptonabber-sync/v2/config/chain"
	rpcconfig "github.com/jrh3k5/cryptonabber-sync/v2/config/rpc"
	"github.com/jrh3k5/cryptonabber-sync/v2/http/json/rpc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var evmNode *rpc.MockEVMNode
var rpcConfigurationResolver rpcconfig.ConfigurationResolver
var chainName = "ethereum"

func TestEvm(t *testing.T) {
	BeforeSuite(func() {
		evmNode = rpc.StartMockEVMNode()

		rpcConfigurationResolver = rpcconfig.NewDefaultConfigurationResolver([]rpcconfig.Configuration{
			{
				RPCURL:    evmNode.URL(),
				ChainName: chainName,
				ChainType: chain.TypeEVM,
			},
		})

		DeferCleanup(evmNode.Stop)
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Evm Suite")
}
