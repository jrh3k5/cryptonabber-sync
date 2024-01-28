package coingecko

import (
	"context"
	"fmt"
	"math/big"
)

// AssetPlatformIDResolver is a resolver for Coingecko asset platform IDs.
type AssetPlatformIDResolver interface {
	// ResolveForChainID tries to resolve a Coingecko asset platform ID for the given chain ID
	ResolveForChainID(ctx context.Context, chainID *big.Int) (string, error)
}

// SimpleAssetPlatformIDResolver resolves asset platform IDs based on a hardcoded list.
type SimpleAssetPlatformIDResolver struct {
}

func NewSimpleAssetPlatformIDResolver() *SimpleAssetPlatformIDResolver {
	return &SimpleAssetPlatformIDResolver{}
}

func (*SimpleAssetPlatformIDResolver) ResolveForChainID(ctx context.Context, chainID *big.Int) (string, error) {
	if !chainID.IsInt64() {
		return "", fmt.Errorf("unsupported value type for chain ID: %v", chainID)
	}

	switch chainID.Int64() {
	case 1:
		return "ethereum", nil
	case 137:
		return "polygon-pos", nil
	case 8453:
		return "base", nil
	case 42161:
		return "arbitrum-one", nil
	case 43114:
		return "avalanche", nil
	}

	return "", fmt.Errorf("unsupported chain ID value: %d", chainID.Int64())
}
