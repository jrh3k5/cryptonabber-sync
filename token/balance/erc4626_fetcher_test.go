package balance_test

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jrh3k5/cryptonabber-sync/v3/config"
	"github.com/jrh3k5/cryptonabber-sync/v3/http/json/rpc"
	"github.com/jrh3k5/cryptonabber-sync/v3/token/balance"
)

var _ = Describe("Erc4626Fetcher", func() {
	var fetcher *balance.ERC4262Fetcher

	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()

		fetcher = balance.NewERC4262Fetcher(rpcConfigurationResolver, http.DefaultClient)
	})

	Context("FetchBalance", func() {
		It("resolves the balance from the user's shares", func() {
			sharesBalance := big.NewInt(100)
			assets := big.NewInt(200)

			walletAddress := "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97"
			vaultAddress := "0x192850F437160A09A48a056Df3C2dacc68769d34"

			evmNode.RegisterETHCallCall("getShares", vaultAddress, []string{"address"}, func(_ string, params []string) (rpc.MockEVMNodeRPCResult, *rpc.MockEVMNodeRPCError, error) {
				if len(params) != 1 {
					return nil, nil, fmt.Errorf("expected 1 parameter, got %d", len(params))
				}

				expectedAddress := strings.TrimPrefix(walletAddress, "0x")
				if params[0] != expectedAddress {
					return nil, nil, fmt.Errorf("expected address '%s', got '%s'", expectedAddress, params[0])
				}

				return rpc.NewMockEVMNodeRPCNumericResult(sharesBalance), nil, nil
			})

			evmNode.RegisterETHCallCall("convertToAssets", vaultAddress, []string{"uint256"}, func(_ string, params []string) (rpc.MockEVMNodeRPCResult, *rpc.MockEVMNodeRPCError, error) {
				if len(params) != 1 {
					return nil, nil, fmt.Errorf("expected 1 parameter, got %d", len(params))
				}

				inputBigInt := new(big.Int)
				inputBigInt.SetString(params[0], 16)

				if inputBigInt.Cmp(sharesBalance) != 0 {
					return nil, nil, fmt.Errorf("expected shares balance '%s', got '%s'", sharesBalance.Text(10), inputBigInt.Text(10))
				}

				return rpc.NewMockEVMNodeRPCNumericResult(assets), nil, nil
			})

			erc4626Account := &config.ERC4626Account{
				OnchainAsset: config.OnchainAsset{
					ChainName: chainName,
				},
				OnchainWallet: config.OnchainWallet{
					WalletAddress: walletAddress,
				},
				VaultAddress:        vaultAddress,
				BalanceFunctionName: "getShares",
			}

			balance, err := fetcher.FetchBalance(ctx, erc4626Account)
			Expect(err).ToNot(HaveOccurred(), "fetching the balance should not fail")

			Expect(balance).To(Equal(assets), "the balance should be correct")
		})
	})
})
