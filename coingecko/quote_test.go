package coingecko_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/jarcoal/httpmock"
	"github.com/jrh3k5/cryptonabber-sync/coingecko"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTPQuoteResolver", func() {
	var resolver *coingecko.HTTPQuoteResolver

	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()

		resolver = coingecko.NewHTTPQuoteResolver(http.DefaultClient)
	})

	Context("ResolveQuote", func() {
		var usdPriceByAssetPlatformAndContract map[string]map[string]string

		BeforeEach(func() {
			usdPriceByAssetPlatformAndContract = make(map[string]map[string]string)

			urlRegexp := regexp.MustCompile(`https\:\/\/api\.coingecko\.com\/api\/v3\/coins\/([a-z\-]+)\/contract\/(.+)`)
			httpmock.RegisterRegexpResponder(http.MethodGet, urlRegexp, func(request *http.Request) (*http.Response, error) {
				matches := urlRegexp.FindStringSubmatch(request.URL.String())
				if len(matches) != 3 {
					return httpmock.NewStringResponse(http.StatusBadRequest, fmt.Sprintf("cannot parse request URL: %s", request.URL.String())), nil
				}

				assetPlatformID := matches[1]
				contractAddress := matches[2]

				byContract, hasPlatform := usdPriceByAssetPlatformAndContract[assetPlatformID]
				if !hasPlatform {
					return httpmock.NewStringResponse(http.StatusNotFound, fmt.Sprintf("asset platform ID '%s' is not set up for the test", assetPlatformID)), nil
				}

				price, hasContract := byContract[contractAddress]
				if !hasContract {
					return httpmock.NewStringResponse(http.StatusNotFound, fmt.Sprintf("contracct address '%s' is not set up for the test", contractAddress)), nil
				}

				jsonResponse := fmt.Sprintf(`{ "market_data": { "current_price": { "usd": %s } } }`, price)
				return httpmock.NewStringResponse(http.StatusOK, jsonResponse), nil
			})
		})

		When("the quote is only in whole dollars", func() {
			It("successfully parses the quote data", func() {
				assetPlatformID := "test-whole-dollars"
				contractAddress := "0xwholedollars"
				usdPriceByAssetPlatformAndContract[assetPlatformID] = map[string]string{
					contractAddress: "12345",
				}

				dollars, cents, hasQuote, err := resolver.ResolveQuote(ctx, assetPlatformID, contractAddress)
				Expect(err).ToNot(HaveOccurred(), "getting the quote should not fail")
				Expect(hasQuote).To(BeTrue(), "there should be a quote retrieved")
				Expect(cents).To(Equal(float64(0)), "there should be no cents in the quote")
				Expect(dollars).To(Equal(int64(12345)), "the dollar amount should be successfully parsed out")
			})
		})

		When("the quote is in dollars and cents", func() {
			It("successfully parses the quote data", func() {
				assetPlatformID := "test-dollars-and-cents"
				contractAddress := "0xdollarsandcents"
				usdPriceByAssetPlatformAndContract[assetPlatformID] = map[string]string{
					contractAddress: "54321.6789",
				}

				dollars, cents, hasQuote, err := resolver.ResolveQuote(ctx, assetPlatformID, contractAddress)
				Expect(err).ToNot(HaveOccurred(), "getting the quote should not fail")
				Expect(hasQuote).To(BeTrue(), "there should be a quote retrieved")
				Expect(cents).To(Equal(float64(0.6789)), "there cents amount should be successfully parsed out")
				Expect(dollars).To(Equal(int64(54321)), "the dollar amount should be successfully parsed out")
			})
		})

		When("there is no quote data available", func() {
			It("indicates that no quote data was found", func() {
				_, _, hasQuote, err := resolver.ResolveQuote(ctx, "not-found", "0xnuh-uh")
				Expect(err).ToNot(HaveOccurred(), "resolving the quote should not fail")
				Expect(hasQuote).To(BeFalse(), "no quote should have been resolved")
			})
		})
	})

	Context("ResolveETHQuote", func() {
		It("resolves the ETH quote value", func() {
			httpmock.RegisterResponder(http.MethodGet, `=~^https:\/\/api\.coingecko\.com/api\/v3\/simple/price\?ids=ethereum&vs_currencies=usd(&.*)?`,
				httpmock.NewStringResponder(http.StatusOK, `{ "ethereum" : { "usd": 2273.14 } } }`))

			ethDollars, ethCents, err := resolver.ResolveETHQuote(ctx)
			Expect(err).ToNot(HaveOccurred(), "resolving the ETH quote should not fail")
			Expect(ethCents).To(Equal(float64(0.14)), "the correct cents value should be resolved")
			Expect(ethDollars).To(Equal(int64(2273)), "the correct dollar value should be resolved")
		})
	})
})
