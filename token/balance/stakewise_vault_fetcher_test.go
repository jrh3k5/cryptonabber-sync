package balance_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jarcoal/httpmock"
	"github.com/jrh3k5/cryptonabber-sync/v2/http/json/rpc"
	"github.com/jrh3k5/cryptonabber-sync/v2/token/balance"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StakewiseVaultFetcher", func() {
	var sharesByContractAddress map[string]map[string]int64
	var assetsPerShares map[int64]int64
	var nodeURL string

	var fetcher *balance.StakewiseVaultFetcher

	var ctx context.Context

	BeforeEach(func() {
		sharesByContractAddress = make(map[string]map[string]int64)
		assetsPerShares = make(map[int64]int64)

		ctx = context.Background()

		nodeURL = "http://node.localhost.stakewise/rpc"
		httpmock.RegisterResponder(http.MethodPost, nodeURL, func(request *http.Request) (*http.Response, error) {
			requestBody := &rpc.Request{}
			if unmarshalErr := json.NewDecoder(request.Body).Decode(requestBody); unmarshalErr != nil {
				responseErr := fmt.Errorf("failed to unmarshal request body: %w", unmarshalErr)
				return httpmock.NewStringResponse(http.StatusInternalServerError, responseErr.Error()), responseErr
			}

			funcArgs := requestBody.Params[0].(map[string]any)
			targetAddress := funcArgs["to"].(string)
			data := funcArgs["data"].(string)

			if strings.HasPrefix(data, crypto.Keccak256Hash(([]byte("getShares(address)"))).String()[0:10]) {
				walletAddress := data[10+24:]

				var tokenBalance int64
				if walletShares, hasToken := sharesByContractAddress[targetAddress]; hasToken {
					tokenBalance = walletShares["0x"+walletAddress]
				}

				response := &rpc.Response{
					ID:      json.Number(fmt.Sprintf("%d", requestBody.ID)),
					JSONRPC: requestBody.JSONRPC,
					Result:  fmt.Sprintf("0x%x", tokenBalance),
				}

				return httpmock.NewJsonResponse(http.StatusOK, response)
			} else if strings.HasPrefix(data, crypto.Keccak256Hash([]byte("convertToAssets(uint256)")).String()[0:10]) {
				assetsHex := data[10+24:]

				sharesBigInt := big.NewInt(0)
				sharesBigInt.SetString(assetsHex, 16)

				var assetsBalance int64
				if shareAssets, hasAssets := assetsPerShares[sharesBigInt.Int64()]; hasAssets {
					assetsBalance = shareAssets
				}

				response := &rpc.Response{
					ID:      json.Number(fmt.Sprintf("%d", requestBody.ID)),
					JSONRPC: requestBody.JSONRPC,
					Result:  fmt.Sprintf("0x%x", assetsBalance),
				}

				return httpmock.NewJsonResponse(http.StatusOK, response)
			}

			return httpmock.NewStringResponse(http.StatusBadRequest, "unsupported request"), nil
		})

		fetcher = balance.NewStakewiseVaultFetcher(nodeURL, http.DefaultClient)
	})

	It("fetches the balance", func() {
		vaultContractAddress := "0xaab82B7F80b05C6f203409011C56ae818BF9b5F7"
		walletAddress := "0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5"

		sharesByContractAddress[vaultContractAddress] = map[string]int64{
			walletAddress: int64(3324834172173128020),
		}
		assetsPerShares[3324834172173128020] = 3358503866879544656

		fetchedBalance, err := fetcher.FetchBalance(ctx, vaultContractAddress, walletAddress)
		Expect(err).ToNot(HaveOccurred(), "retrieving the balance should not fail")
		Expect(fetchedBalance.Int64()).To(Equal(int64(3358503866879544656)), "the number of assets should be returned as the balance")
	})
})
