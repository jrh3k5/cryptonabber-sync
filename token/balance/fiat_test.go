package balance_test

import (
	"math/big"

	"github.com/jrh3k5/cryptonabber-sync/token/balance"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fiat", func() {
	Context("AsFiat", func() {
		DescribeTable("token decimals", func(decimals int, expectedFiatValue int) {
			Expect(balance.AsFiat(big.NewInt(int64(12345)), decimals, 1, 0.05)).To(Equal(int64(expectedFiatValue)), "the correct fiat amount should be calculated")
		},
			Entry("decimals of zero", 0, 1296225),
			Entry("decimals of one", 1, 129622),
			Entry("decimals of two", 2, 12962),
			Entry("decimals of three", 3, 1296))

		DescribeTable("dollar rate", func(dollarRate int, expectedFiatValue int) {
			Expect(balance.AsFiat(big.NewInt(int64(98765)), 0, int64(dollarRate), 0.5)).To(Equal(int64(expectedFiatValue)), "the correct fiat amount should be calculated")
		},
			Entry("dollar rate of zero", 0, 4938250),
			Entry("dollar rate of one", 1, 14814750),
			Entry("dollar rate of two", 2, 24691250))

		When("the cents rate is past two significant figures", func() {
			It("correctly calculates the fiat value", func() {
				Expect(balance.AsFiat(big.NewInt(int64(98765)), 0, 1, 0.0005)).To(Equal(int64(9881438)), "the correct fiat amount should be calculated")
			})
		})

		It("calculates a realistic ETH balance scenario", func() {
			tokenBalance := big.NewInt(int64(2410555693229900000))
			Expect(balance.AsFiat(tokenBalance, 18, 2493, 0.38)).To(Equal(int64(601043)), "the correct fiat amount should be calculated")
		})
	})
})
