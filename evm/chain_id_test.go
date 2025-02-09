package evm_test

import (
	"context"
	"math/big"
	"net/http"

	"github.com/jrh3k5/cryptonabber-sync/v2/evm"
	"github.com/jrh3k5/cryptonabber-sync/v2/http/json/rpc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ChainID", func() {
	var fetcher *evm.JSONRPCChainIDFetcher

	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()

		fetcher = evm.NewJSONRPCChainIDFetcher(rpcConfigurationResolver, http.DefaultClient)
	})

	It("returns the chain ID", func() {
		chainID := int64(3378394983)

		evmNode.RegisterRPCMethodCall("eth_chainId", func(methodName string) (rpc.MockEVMNodeRPCResult, *rpc.MockEVMNodeRPCError, error) {
			return rpc.NewMockEVMNodeRPCNumericResult(big.NewInt(chainID)), nil, nil
		})

		retrievedChainID, err := fetcher.GetChainID(ctx, chainName)
		Expect(err).ToNot(HaveOccurred(), "getting the chain ID should not fail")
		Expect(retrievedChainID.Int64()).To(Equal(chainID), "the correct chain ID should be retrieved")
	})
})
