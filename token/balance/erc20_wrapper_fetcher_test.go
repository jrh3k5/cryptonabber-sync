package balance_test

import (
	"context"
	"math/big"
	"sync"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
	"github.com/jrh3k5/cryptonabber-sync/v3/token/balance"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ERC20WrapperFetcher", func() {
	var erc20Fetcher *testERC20Fetcher
	var erc20WrapperFetcher *balance.ERC20WrapperFetcher

	BeforeEach(func() {
		erc20Fetcher = newTestERC20Fetcher()
		erc20WrapperFetcher = balance.NewERC20WrapperFetcher(erc20Fetcher)
	})

	Context("FetchBalance", func() {
		It("returns the correct balance", func() {
			erc20Fetcher.setBalance("0x123", big.NewInt(100))
			balance, err := erc20WrapperFetcher.FetchBalance(context.Background(), &config.ERC20WrapperAccount{
				ERC20Account: config.ERC20Account{
					TokenAddress: "0x123",
				},
			})

			Expect(err).ToNot(HaveOccurred(), "fetching the balance should not fail")
			Expect(balance).To(Equal(big.NewInt(100)), "the balance should be correct")
		})
	})
})

type testERC20Fetcher struct {
	balancesMutex sync.RWMutex
	balances      map[string]*big.Int
}

func newTestERC20Fetcher() *testERC20Fetcher {
	return &testERC20Fetcher{
		balances: make(map[string]*big.Int),
	}
}

func (t *testERC20Fetcher) FetchBalance(ctx context.Context, onchainAccount *config.ERC20Account) (*big.Int, error) {
	t.balancesMutex.RLock()
	defer t.balancesMutex.RUnlock()

	balance, hasBalance := t.balances[onchainAccount.TokenAddress]
	if !hasBalance {
		return big.NewInt(0), nil
	}

	return balance, nil
}

func (t *testERC20Fetcher) setBalance(address string, balance *big.Int) {
	t.balancesMutex.Lock()
	defer t.balancesMutex.Unlock()

	t.balances[address] = balance
}
