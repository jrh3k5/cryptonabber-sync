package token_test

import (
	"context"
	"math/big"
	"net/http"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
	"github.com/jrh3k5/cryptonabber-sync/v3/http/json/rpc"
	"github.com/jrh3k5/cryptonabber-sync/v3/token"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RPCDecimalsResolver", func() {
	var resolver *token.RPCDecimalsResolver

	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()

		resolver = token.NewRPCDecimalsResolver(rpcConfigurationResolver, http.DefaultClient)
	})

	Describe("ResolveDecimals", func() {
		When("the request is for ETH", func() {
			It("returns 18", func() {
				decimals, err := resolver.ResolveDecimals(ctx, config.OnchainAsset{ChainName: chainName}, nil)
				Expect(err).NotTo(HaveOccurred(), "resolving the decimals should not fail")
				Expect(decimals).To(BeNumerically("==", 18), "the correct decimals should be returned")
			})

			When("the request is for an ERC20", func() {
				It("returns the correct decimals", func() {
					tokenAddress := "0xfedE2dB34D22b46C2014Ba1d57fE31134012e1b2"

					evmNode.RegisterETHCallCall("decimals", tokenAddress, nil, func(_ string, _ []string) (rpc.MockEVMNodeRPCResult, *rpc.MockEVMNodeRPCError, error) {
						return rpc.NewMockEVMNodeRPCNumericResult(big.NewInt(6)), nil, nil
					})

					decimals, err := resolver.ResolveDecimals(ctx, config.OnchainAsset{ChainName: chainName}, &tokenAddress)
					Expect(err).NotTo(HaveOccurred(), "resolving the decimals should not fail")
					Expect(decimals).To(BeNumerically("==", 6), "the correct decimals should be returned")
				})
			})
		})
	})
})
