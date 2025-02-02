package coingecko

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	synchttp "github.com/jrh3k5/cryptonabber-sync/v2/http"
)

// QuoteResolver can be used to resolve quotes.
type QuoteResolver interface {
	// ResolveETHQuote resolves a quote for Ether.
	// It returns dollars and cents. The cents is expressed as a floating value relative to whole dollars to account for assets that may
	// have a value of less than 1 cent per whole token.
	// This method enforces the assumption that there is always a price of ETH available.
	ResolveETHQuote(ctx context.Context) (int64, float64, error)

	// ResolveQuote resolves a quote for the given contract address on the given asset platform.
	// It returns dollars and cents. The cents is expressed as a floating value relative to whole dollars to account for assets that may
	// have a value of less than 1 cent per whole token.
	// The returned bool is true if a quote was found; false if not.
	ResolveQuote(ctx context.Context, assetPlatformID string, contractAddress string) (int64, float64, bool, error)
}

// HTTPQuoteResolver is a QuoteResolver that uses Coingecko's HTTP API to resolve quotes.
type HTTPQuoteResolver struct {
	doer synchttp.Doer
}

func NewHTTPQuoteResolver(doer synchttp.Doer) *HTTPQuoteResolver {
	return &HTTPQuoteResolver{
		doer: doer,
	}
}

func (q *HTTPQuoteResolver) ResolveETHQuote(ctx context.Context) (int64, float64, error) {
	requestURL := "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd&include_market_cap=false&include_24hr_vol=false&include_24hr_change=false&include_last_updated_at=false&precision=false"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to build request: %w", err)
	}

	response, err := q.doer.Do(request)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to execute request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		statusErr := synchttp.BuildUnexpectedStatusErr(response)
		return 0, 0, statusErr
	}
	defer func() {
		_ = response.Body.Close()
	}()

	responseBody := make(map[string]map[string]json.Number)
	if unmarshalErr := json.NewDecoder(response.Body).Decode(&responseBody); unmarshalErr != nil {
		return 0, 0, fmt.Errorf("failed to unmarshal response body: %w", unmarshalErr)
	}

	ethereumPrices, hasEthereum := responseBody["ethereum"]
	if !hasEthereum {
		return 0, 0, errors.New("ethereum not returned in response")
	}

	usdPrice, hasUSD := ethereumPrices["usd"]
	if !hasUSD {
		return 0, 0, errors.New("USD price not found in Ethereum prices")
	}

	dollars, centsRatio, _, err := q.parsePrice(usdPrice)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse price: %w", err)
	}

	return dollars, centsRatio, nil
}

func (q *HTTPQuoteResolver) ResolveQuote(ctx context.Context, assetPlatformID string, contractAddress string) (int64, float64, bool, error) {
	requestURL := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/contract/%s", assetPlatformID, contractAddress)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to build request: %w", err)
	}

	response, err := q.doer.Do(request)
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to execute request: %w", err)
	} else if response.StatusCode == http.StatusNotFound {
		return 0, 0, false, nil
	} else if response.StatusCode != http.StatusOK {
		statusErr := synchttp.BuildUnexpectedStatusErr(response)
		return 0, 0, false, statusErr
	}
	defer func() {
		_ = response.Body.Close()
	}()

	responseBody := &coinDetailsResponse{}
	if unmarshalErr := json.NewDecoder(response.Body).Decode(responseBody); unmarshalErr != nil {
		return 0, 0, false, fmt.Errorf("failed to unmarshal response body: %w", unmarshalErr)
	}

	if responseBody.MarketData == nil || responseBody.MarketData.CurrentPrice == nil {
		return 0, 0, false, nil
	}

	usdPrice, hasUSD := responseBody.MarketData.CurrentPrice["usd"]
	if !hasUSD {
		return 0, 0, false, nil
	}

	return q.parsePrice(usdPrice)
}

func (*HTTPQuoteResolver) parsePrice(price json.Number) (int64, float64, bool, error) {
	usdPriceString := price.String()
	if usdPriceString == "" {
		return 0, 0, false, nil
	}

	splitUSD := strings.Split(usdPriceString, ".")
	dollars, parseErr := strconv.ParseInt(splitUSD[0], 10, 64)
	if parseErr != nil {
		return 0, 0, false, fmt.Errorf("failed to parse USD dollars '%s': %w", splitUSD[0], parseErr)
	}

	if len(splitUSD) == 1 {
		return dollars, 0, true, nil
	}

	centsString := splitUSD[1]
	cents, parseErr := strconv.ParseInt(centsString, 10, 64)
	if parseErr != nil {
		return 0, 0, false, fmt.Errorf("failed to parse USD cents '%s': %w", splitUSD[1], parseErr)
	}

	centsRatio := float64(cents) / math.Pow10(len(centsString))

	return dollars, centsRatio, true, nil
}

type coinDetailsResponse struct {
	MarketData *coinDetailMarketData `json:"market_data"`
}

type coinDetailMarketData struct {
	CurrentPrice map[string]json.Number `json:"current_price"`
}
