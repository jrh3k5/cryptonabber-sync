package balance_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jrh3k5/cryptonabber-sync/v3/config/chain"
	rpcconfig "github.com/jrh3k5/cryptonabber-sync/v3/config/rpc"
	"github.com/jrh3k5/cryptonabber-sync/v3/http/json/rpc"
)

var evmNode *rpc.MockEVMNode
var rpcConfigurationResolver rpcconfig.ConfigurationResolver
var chainName = "ethereum"

func TestBalance(t *testing.T) {
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
	RunSpecs(t, "Balance Suite")
}
