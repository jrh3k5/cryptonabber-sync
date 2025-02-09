package token

import (
	"context"

	"github.com/jrh3k5/cryptonabber-sync/v2/config"
)

type DecimalsResolver interface {
	ResolveDecimals(ctx context.Context, onchainAsset config.OnchainAsset, tokenAddress *string) (int, error)
}
