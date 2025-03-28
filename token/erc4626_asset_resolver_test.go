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

var _ = Describe("ERC4626AssetResolver", func() {
	var erc4626AssetResolver *token.ERC4626AssetResolver

	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()

		erc4626AssetResolver = token.NewERC4626AssetResolver(rpcConfigurationResolver, http.DefaultClient)
	})

	When("the contract has an asset function on it", func() {
		It("returns the address of the asset", func() {
			vaultAddres := "0x68d30f47F19c07bCCEf4Ac7FAE2Dc12FCa3e0dC9"
			assetAddress := "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97"
			evmNode.RegisterETHCallCall("asset", vaultAddres, nil, func(_ string, _ []string) (rpc.MockEVMNodeRPCResult, *rpc.MockEVMNodeRPCError, error) {
				return rpc.NewMockEVMNodeRPCAddressResult(assetAddress), nil, nil
			})

			address, err := erc4626AssetResolver.ResolveAssetAddress(ctx, &config.ERC4626Account{
				OnchainAsset: config.OnchainAsset{
					ChainName: chainName,
				},
				VaultAddress: vaultAddres,
			})

			Expect(err).ToNot(HaveOccurred(), "resolving the asset address should not fail")
			Expect(address).ToNot(BeNil(), "the asset address should be returned")
			Expect(*address).To(Equal(assetAddress), "the asset address should be returned")
		})
	})

	When("the contract has no asset function", func() {
		It("returns nil for the asset address", func() {
			vaultAddress := "0xdAC17F958D2ee523a2206206994597C13D831ec7"
			evmNode.RegisterContractExistence(vaultAddress)

			address, err := erc4626AssetResolver.ResolveAssetAddress(ctx, &config.ERC4626Account{
				OnchainAsset: config.OnchainAsset{
					ChainName: chainName,
				},
				VaultAddress: vaultAddress,
			})

			Expect(err).ToNot(HaveOccurred(), "resolving the asset address should not fail")
			Expect(address).To(BeNil(), "the asset address should be nil")
		})
	})

	When("the account has no vault address", func() {
		It("rejects the request", func() {
			_, err := erc4626AssetResolver.ResolveAssetAddress(ctx, &config.ERC4626Account{
				OnchainAsset: config.OnchainAsset{
					ChainName: chainName,
				},
			})

			Expect(err).To(MatchError(ContainSubstring("vault address must be provided on the given onchain account")), "the error should be about the missing vault address")
		})
	})
})
