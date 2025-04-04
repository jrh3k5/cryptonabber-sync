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
			vaultAddress := "0x68d30f47F19c07bCCEf4Ac7FAE2Dc12FCa3e0dC9"
			assetAddress := "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97"
			evmNode.RegisterETHCallCall("asset", vaultAddress, nil, func(_ string, _ []string) (rpc.MockEVMNodeRPCResult, *rpc.MockEVMNodeRPCError, error) {
				return rpc.NewMockEVMNodeRPCAddressResult(assetAddress), nil, nil
			})

			address, err := erc4626AssetResolver.ResolveAssetAddress(ctx, &config.ERC4626Account{
				OnchainAsset: config.OnchainAsset{
					ChainName: chainName,
				},
				VaultAddress: vaultAddress,
			})

			Expect(err).ToNot(HaveOccurred(), "resolving the asset address should not fail")
			Expect(address).ToNot(BeNil(), "the asset address should be returned")
			Expect(*address).To(Equal(assetAddress), "the asset address should be returned")
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

	When("the account has a backing asset specified", func() {
		It("returns the contract address of the configured backing asset", func() {
			assetAddress := "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97"
			address, err := erc4626AssetResolver.ResolveAssetAddress(ctx, &config.ERC4626Account{
				OnchainAsset: config.OnchainAsset{
					ChainName: chainName,
				},
				BackingAsset: &config.ERC4626BackingAsset{
					ContractAddress: &assetAddress,
				},
			})

			Expect(err).ToNot(HaveOccurred(), "resolving the asset address should not fail")
			Expect(address).ToNot(BeNil(), "the asset address should be returned")
			Expect(*address).To(Equal(assetAddress), "the asset address should be returned")
		})

		When("the backing asset has no contract address", func() {
			It("returns nil for the contract address to accomodate assets such as ETH", func() {
				address, err := erc4626AssetResolver.ResolveAssetAddress(ctx, &config.ERC4626Account{
					OnchainAsset: config.OnchainAsset{
						ChainName: chainName,
					},
					BackingAsset: &config.ERC4626BackingAsset{},
				})

				Expect(err).ToNot(HaveOccurred(), "resolving the asset address should not fail")
				Expect(address).To(BeNil(), "the asset address should be nil")
			})
		})
	})
})
