package balance

import (
	"math"
	"math/big"
)

// AsFiat will convert the given token balance into a fiat value, expressed as the number of cents (e.g., $123.45 fiat of a token balance will be returned as 12345).
func AsFiat(tokenBalance *big.Int, tokenDecimals int, dollarRate int64, centsRate float64) int64 {
	if tokenDecimals == 0 {
		return computeFiatValue(tokenBalance.Int64(), dollarRate, centsRate)
	}

	balanceInt := tokenBalance.Int64()
	decimalsDivisor := int64(math.Pow10(tokenDecimals))
	fractionalTokens := balanceInt % int64(math.Pow10(tokenDecimals))
	wholeTokens := (balanceInt - fractionalTokens) / decimalsDivisor

	// Reduce the granularity of the fractional tokens to avoid integer overflow
	fractionalTokenLimitFactor := 6
	fractionalTokenLimit := int64(math.Pow10(fractionalTokenLimitFactor))
	if fractionalTokens > fractionalTokenLimit {
		fractionalTokens = (fractionalTokens - (fractionalTokens % fractionalTokenLimit)) / fractionalTokenLimit
		decimalsDivisor = int64(math.Pow10(tokenDecimals - fractionalTokenLimitFactor))
	}

	wholeTokenFiat := computeFiatValue(wholeTokens, dollarRate, centsRate)
	fractionalTokenFiat := computeFiatValue(fractionalTokens, dollarRate, centsRate)
	centsFiat := (fractionalTokenFiat - (fractionalTokenFiat % decimalsDivisor)) / decimalsDivisor

	return wholeTokenFiat + centsFiat
}

func computeFiatValue(amount int64, dollarRate int64, centsRate float64) int64 {
	dollarTotal := amount * dollarRate
	centsTotal := int64(centsRate * float64(amount) * 100)
	return (dollarTotal * 100) + centsTotal
}
