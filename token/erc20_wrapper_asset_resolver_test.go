package token_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
	"github.com/jrh3k5/cryptonabber-sync/v3/http/json/rpc"
	"github.com/jrh3k5/cryptonabber-sync/v3/token"
)

var _ = Describe("ERC20WrapperAssetResolver", func() {
	var erc4626AssetResolver *token.ERC20WrapperAssetResolver

	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()

		erc4626AssetResolver = token.NewERC20WrapperAssetResolver(rpcConfigurationResolver, http.DefaultClient)
	})

	Describe("ResolveAssetAddress", func() {
		It("returns the address for the onchain asset that represents the value of the token", func() {
			tokenAddress := "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97"
			baseTokenAddress := "0x388C818CA8B9251b393131C08a736A67ccB19297"
			baseTokenFunction := "baseToken"

			evmNode.RegisterETHCallCall(baseTokenFunction, tokenAddress, nil, func(_ string, _ []string) (rpc.MockEVMNodeRPCResult, *rpc.MockEVMNodeRPCError, error) {
				return rpc.NewMockEVMNodeRPCAddressResult(baseTokenAddress), nil, nil
			})

			address, err := erc4626AssetResolver.ResolveAssetAddress(ctx, &config.ERC20WrapperAccount{
				ERC20Account: config.ERC20Account{
					OnchainAsset: config.OnchainAsset{
						ChainName: chainName,
					},
					TokenAddress: tokenAddress,
				},
				BaseTokenAddressFunction: baseTokenFunction,
			})

			Expect(err).To(BeNil(), "resolving the asset address should not fail")
			Expect(*address).To(Equal(baseTokenAddress), "the asset address should be the base token address")
		})
	})
})
