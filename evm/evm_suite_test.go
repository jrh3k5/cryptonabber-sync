package evm_test

import (
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEvm(t *testing.T) {
	BeforeSuite(func() {
		httpmock.Activate()
		DeferCleanup(httpmock.DeactivateAndReset)
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Evm Suite")
}
