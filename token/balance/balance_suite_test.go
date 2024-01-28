package balance_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jarcoal/httpmock"
)

func TestBalance(t *testing.T) {
	BeforeSuite(func() {
		httpmock.Activate()
		DeferCleanup(httpmock.DeactivateAndReset)
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Balance Suite")
}
