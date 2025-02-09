package token_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
	"github.com/jrh3k5/cryptonabber-sync/v3/token"
)

var _ = Describe("ERC20AssetResolver", func() {
	var erc20AssetResolver *token.ERC20AssetResolver

	BeforeEach(func() {
		erc20AssetResolver = token.NewERC20AssetResolver()
	})

	Describe("ResolveAssetAddress", func() {
		It("should return the token address", func() {
			tokenAddress := "0x1234"
			address, err := erc20AssetResolver.ResolveAssetAddress(nil, &config.ERC20Account{
				TokenAddress: tokenAddress,
			})
			Expect(err).To(BeNil())
			Expect(*address).To(Equal(tokenAddress))
		})
	})
})
