package balance_test

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
	"github.com/jrh3k5/cryptonabber-sync/v3/http/json/rpc"
	"github.com/jrh3k5/cryptonabber-sync/v3/token/balance"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ERC20Fetcher", func() {
	var fetcher *balance.ERC20Fetcher

	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()

		fetcher = balance.NewERC20Fetcher(rpcConfigurationResolver, http.DefaultClient)
	})

	It("fetches and retrieves the correct balance", func() {
		contractAddress := "0xa83114A443dA1CecEFC50368531cACE9F37fCCcb"
		walletAddress := "0x2870d53DcAc4763D6b0C030fbE0555405B09CDb3"
		balance := big.NewInt(478932974)

		evmNode.RegisterETHCallCall("balanceOf", contractAddress, []string{"address"}, func(_ string, params []string) (rpc.MockEVMNodeRPCResult, *rpc.MockEVMNodeRPCError, error) {
			if len(params) != 1 {
				return nil, nil, fmt.Errorf("expected 1 parameter, got %d: %s", len(params), strings.Join(params, ", "))
			}

			expectedAddress := strings.TrimPrefix(walletAddress, "0x")
			if params[0] != expectedAddress {
				return nil, nil, fmt.Errorf("expected address to be %s, got %s", expectedAddress, params[0])
			}

			return rpc.NewMockEVMNodeRPCNumericResult(balance), nil, nil
		})

		retrievedBalance, err := fetcher.FetchBalance(ctx, &config.ERC20Account{
			TokenAddress: contractAddress,
			OnchainAsset: config.OnchainAsset{
				ChainName: chainName,
			},
			OnchainWallet: config.OnchainWallet{
				WalletAddress: walletAddress,
			},
		})

		Expect(err).ToNot(HaveOccurred(), "getting the balance should not fail")
		Expect(retrievedBalance).To(Equal(balance), "the correct balance should be returned")
	})
})
