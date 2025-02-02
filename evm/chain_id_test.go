package evm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/jrh3k5/cryptonabber-sync/v2/evm"
	"github.com/jrh3k5/cryptonabber-sync/v2/http/json/rpc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ChainID", func() {
	var nodeURL string
	var fetcher *evm.JSONRPCChainIDFetcher
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()

		nodeURL = "http://rpc.localhost/chain_id"
		fetcher = evm.NewJSONRPCChainIDFetcher(nodeURL, http.DefaultClient)
	})

	It("returns the chain ID", func() {
		chainID := int64(3378394983)

		jsonResponder, err := httpmock.NewJsonResponder(http.StatusOK, &rpc.Response{
			ID:      json.Number("1"),
			JSONRPC: "2.0",
			Result:  "0x" + fmt.Sprintf("%x", chainID),
		})
		Expect(err).ToNot(HaveOccurred(), "building the responder should not fail")
		httpmock.RegisterResponder(http.MethodPost, nodeURL, jsonResponder)

		retrievedChainID, err := fetcher.GetChainID(ctx)
		Expect(err).ToNot(HaveOccurred(), "getting the chain ID should not fail")
		Expect(retrievedChainID.Int64()).To(Equal(chainID), "the correct chain ID should be retrieved")
	})
})
