package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	synchttp "github.com/jrh3k5/cryptonabber-sync/http"
)

// QuoteResolver can be used to resolve quotes.
type QuoteResolver interface {
	// ResolveQuote resolves a quote for the given contract address on the given asset platform.
	// It returns dollars and cents. The cents may not be whole cents (e.g., a quote of 102.43789473 will have cents of 43789473).
	// The returned bool is true if a quote was found; false if not.
	ResolveQuote(ctx context.Context, assetPlatformID string, contractAddress string) (int64, int64, bool, error)
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

func (q *HTTPQuoteResolver) ResolveQuote(ctx context.Context, assetPlatformID string, contractAddress string) (int64, int64, bool, error) {
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

	usdPriceString := usdPrice.String()
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

	cents, parseErr := strconv.ParseInt(splitUSD[1], 10, 64)
	if parseErr != nil {
		return 0, 0, false, fmt.Errorf("failed to parse USD cents '%s': %w", splitUSD[1], parseErr)
	}

	return dollars, cents, true, nil
}

type coinDetailsResponse struct {
	MarketData *coinDetailMarketData `json:"market_data"`
}

type coinDetailMarketData struct {
	CurrentPrice map[string]json.Number `json:"current_price"`
}
