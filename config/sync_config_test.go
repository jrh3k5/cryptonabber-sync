package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jrh3k5/cryptonabber-sync/config"
)

var _ = Describe("SyncConfig", func() {
	Context("SyncedAccount", func() {
		Context("GetQuote", func() {
			When("the quote is whole dollars", func() {
				It("returns the correct quote", func() {
					quoteStr := "12345"
					cfg := &config.SyncedAccount{
						Quote: &quoteStr,
					}

					dollars, cents, hasQuote, err := cfg.GetQuote()
					Expect(err).ToNot(HaveOccurred(), "getting the quote should not fail")
					Expect(hasQuote).To(BeTrue(), "a quote should be returned")
					Expect(cents).To(Equal(float64(0)), "there should be no cents in the quote")
					Expect(dollars).To(Equal(int64(12345)), "the correct dollar amount should be returned")
				})
			})

			When("the quote has cents", func() {
				It("returns the correct quote", func() {
					quoteStr := "98765.054321"
					cfg := &config.SyncedAccount{
						Quote: &quoteStr,
					}

					dollars, cents, hasQuote, err := cfg.GetQuote()
					Expect(err).ToNot(HaveOccurred(), "getting the quote should not fail")
					Expect(hasQuote).To(BeTrue(), "a quote should be returned")
					Expect(cents).To(Equal(float64(0.054321)), "the correct cents amount should be returned")
					Expect(dollars).To(Equal(int64(98765)), "the correct dollar amount should be returned")
				})
			})

			When("the quote is not provided", func() {
				It("returns no quote", func() {
					cfg := &config.SyncedAccount{}
					_, _, hasQuote, err := cfg.GetQuote()
					Expect(err).ToNot(HaveOccurred(), "getting the quote should not fail")
					Expect(hasQuote).To(BeFalse(), "there should be no quote")
				})
			})
		})
	})
})
