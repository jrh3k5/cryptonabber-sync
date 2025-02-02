package balance_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/jrh3k5/cryptonabber-sync/v2/http/json/rpc"
	"github.com/jrh3k5/cryptonabber-sync/v2/token/balance"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ERC20Fetcher", func() {
	var balancesByContractAddress map[string]map[string]int64
	var nodeURL string
	var fetcher *balance.ERC20Fetcher

	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()

		nodeURL = "http://node.localhost/rpc"
		balancesByContractAddress = make(map[string]map[string]int64)
		httpmock.RegisterResponder(http.MethodPost, nodeURL, func(request *http.Request) (*http.Response, error) {
			requestBody := &rpc.Request{}
			if unmarshalErr := json.NewDecoder(request.Body).Decode(requestBody); unmarshalErr != nil {
				responseErr := fmt.Errorf("failed to unmarshal request body: %w", unmarshalErr)
				return httpmock.NewStringResponse(http.StatusInternalServerError, responseErr.Error()), responseErr
			}

			funcArgs := requestBody.Params[0].(map[string]any)
			targetAddress := funcArgs["to"].(string)
			data := funcArgs["data"].(string)
			ownerAddress := data[10+24:]

			var tokenBalance int64
			if walletBalances, hasToken := balancesByContractAddress[targetAddress]; hasToken {
				tokenBalance = walletBalances["0x"+ownerAddress]
			}

			response := &rpc.Response{
				ID:      json.Number(fmt.Sprintf("%d", requestBody.ID)),
				JSONRPC: requestBody.JSONRPC,
				Result:  fmt.Sprintf("0x%x", tokenBalance),
			}

			return httpmock.NewJsonResponse(http.StatusOK, response)
		})

		fetcher = balance.NewERC20Fetcher(nodeURL, http.DefaultClient)
	})

	It("fetches and retrieves the correct balance", func() {
		contractAddress := "0xa83114A443dA1CecEFC50368531cACE9F37fCCcb"
		walletAddress := "0x2870d53DcAc4763D6b0C030fbE0555405B09CDb3"
		balance := int64(478932974)

		balancesByContractAddress[contractAddress] = map[string]int64{
			walletAddress: balance,
		}

		retrievedBalance, err := fetcher.FetchBalance(ctx, contractAddress, walletAddress)
		Expect(err).ToNot(HaveOccurred(), "getting the balance should not fail")
		Expect(retrievedBalance.Int64()).To(Equal(balance), "the correct balance should be returned")
	})
})
