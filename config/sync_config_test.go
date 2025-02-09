package config_test

import (
	"bytes"

	"github.com/jrh3k5/cryptonabber-sync/v2/config"
	"github.com/jrh3k5/cryptonabber-sync/v2/config/chain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("SyncConfig", func() {
	Context("RPC configurations", func() {
		It("successfully deserializes the RPC configuration", func() {
			rpcConfigurationYAML := map[string]any{
				"rpc_url":    "http://localhost:8545",
				"chain_name": "ethereum",
				"chain_type": "evm",
			}

			configBytes, err := yaml.Marshal(map[string]any{
				"rpc_configurations": []any{rpcConfigurationYAML},
			})
			Expect(err).ToNot(HaveOccurred(), "serializing the RPC configuration should not fail")

			syncConfig, err := config.FromYAML(bytes.NewBuffer(configBytes))

			Expect(err).ToNot(HaveOccurred(), "deserializing the RPC configuration should not fail")
			Expect(syncConfig.RPCConfigurations).To(HaveLen(1), "there should be one RPC configuration")

			rpcConfiguration := syncConfig.RPCConfigurations[0]
			Expect(rpcConfiguration.RPCURL).To(Equal("http://localhost:8545"), "the RPC URL should be successfully parsed")
			Expect(rpcConfiguration.ChainName).To(Equal("ethereum"), "the chain name should be successfully parsed")
			Expect(rpcConfiguration.ChainType).To(Equal(chain.TypeEVM), "the chain type should be successfully parsed")
		})
	})

	Context("ynab_accounts", func() {
		Context("ERC20 accounts", func() {
			It("successfully deserializes the ERC20 account", func() {
				erc20AccountYAML := map[string]any{
					"account_name":              "Test ERC20 Account",
					"payee_name":                "Test ERC20 Payee",
					"transaction_category_name": "Test ERC20 Transaction Category",
					"wallet_address":            "0x1234567890123456789012345678901234567890",
					"address_type":              "erc20",
					"chain_name":                "ethereum",
					"token_address":             "0x4567890123456789012345678901234567890",
				}

				accountsYAML := map[string]any{
					"ynab_accounts": []any{erc20AccountYAML},
				}

				yamlBytes, err := yaml.Marshal(accountsYAML)
				Expect(err).ToNot(HaveOccurred(), "serializing the ERC20 account should not fail")

				syncConfig, err := config.FromYAML(bytes.NewBuffer(yamlBytes))

				Expect(err).ToNot(HaveOccurred(), "deserializing the ERC20 account should not fail")
				Expect(syncConfig.Accounts).To(HaveLen(1), "there should be one ERC20 account")

				account := syncConfig.Accounts[0]

				Expect(account.GetAddressType()).To(Equal(config.AddressTypeERC20), "the address type should be successfully parsed")

				erc20Account, err := account.AsERC20Account()
				Expect(err).ToNot(HaveOccurred(), "resolving the ERC20 account should not fail")

				Expect(erc20Account.AccountName).To(Equal("Test ERC20 Account"), "the account name should be successfully parsed")
				Expect(erc20Account.PayeeName).To(Equal("Test ERC20 Payee"), "the payee name should be successfully parsed")
				Expect(erc20Account.TransactionCategoryName).To(Equal("Test ERC20 Transaction Category"), "the transaction category name should be successfully parsed")
				Expect(erc20Account.WalletAddress).To(Equal("0x1234567890123456789012345678901234567890"), "the wallet address should be successfully parsed")
				Expect(erc20Account.TokenAddress).To(Equal("0x4567890123456789012345678901234567890"), "the token address should be successfully parsed")
				Expect(erc20Account.ChainName).To(Equal("ethereum"), "the chain name should be successfully parsed")
			})
		})

		Context("ERC462 accounts", func() {
			var erc4626AccountYAML map[string]any

			BeforeEach(func() {
				erc4626AccountYAML = map[string]any{
					"account_name":              "Test ERC462 Account",
					"payee_name":                "Test ERC462 Payee",
					"transaction_category_name": "Test ERC462 Transaction Category",
					"wallet_address":            "0x1234567890123456789012345678901234567890",
					"address_type":              "erc4626",
					"chain_name":                "ethereum",
					"vault_address":             "0x4567890123456789012345678901234567890",
				}
			})

			It("successfully deserializes the ERC462 account", func() {
				yamlBytes, err := yaml.Marshal(map[string]any{
					"ynab_accounts": []any{erc4626AccountYAML},
				})
				Expect(err).ToNot(HaveOccurred(), "serializing the ERC462 account should not fail")

				syncConfig, err := config.FromYAML(bytes.NewBuffer(yamlBytes))
				Expect(err).ToNot(HaveOccurred(), "deserializing the ERC462 account should not fail")

				Expect(syncConfig.Accounts).To(HaveLen(1), "there should be one ERC462 account")

				account := syncConfig.Accounts[0]

				Expect(account.GetAddressType()).To(Equal(config.AddressTypeERC4626), "the address type should be successfully parsed")

				erc4626Account, err := account.AsERC4626Account()
				Expect(err).ToNot(HaveOccurred(), "resolving the ERC462 account should not fail")

				Expect(erc4626Account.AccountName).To(Equal("Test ERC462 Account"), "the account name should be successfully parsed")
				Expect(erc4626Account.PayeeName).To(Equal("Test ERC462 Payee"), "the payee name should be successfully parsed")
				Expect(erc4626Account.TransactionCategoryName).To(Equal("Test ERC462 Transaction Category"), "the transaction category name should be successfully parsed")
				Expect(erc4626Account.WalletAddress).To(Equal("0x1234567890123456789012345678901234567890"), "the wallet address should be successfully parsed")
				Expect(erc4626Account.ChainName).To(Equal("ethereum"), "the chain name should be successfully parsed")
				Expect(erc4626Account.VaultAddress).To(Equal("0x4567890123456789012345678901234567890"), "the vault address should be successfully parsed")
				Expect(erc4626Account.BalanceFunctionName).To(Equal("balanceOf"), "the balance function name should be the default value")
			})

			When("a balance function is provided", func() {
				BeforeEach(func() {
					erc4626AccountYAML["balance_function"] = "getShares"
				})

				It("populates the account entity with that function name", func() {
					yamlBytes, err := yaml.Marshal(map[string]any{
						"ynab_accounts": []any{erc4626AccountYAML},
					})
					Expect(err).ToNot(HaveOccurred(), "serializing the ERC462 account should not fail")

					syncConfig, err := config.FromYAML(bytes.NewBuffer(yamlBytes))
					Expect(err).ToNot(HaveOccurred(), "deserializing the ERC462 account should not fail")

					Expect(syncConfig.Accounts).To(HaveLen(1), "there should be one ERC462 account")

					account := syncConfig.Accounts[0]

					Expect(account.GetAddressType()).To(Equal(config.AddressTypeERC4626), "the address type should be successfully parsed")

					erc4626Account, err := account.AsERC4626Account()
					Expect(err).ToNot(HaveOccurred(), "resolving the ERC462 account should not fail")

					Expect(erc4626Account.BalanceFunctionName).To(Equal("getShares"), "the balance function name should be successfully parsed")
				})
			})
		})

		Context("ERC20 wrapper accounts", func() {
			It("successfully deserializes the ERC20 wrapper account", func() {
				erc20WrapperAccountYAML := map[string]any{
					"account_name":                "Test ERC20 Wrapper Account",
					"payee_name":                  "Test ERC20 Wrapper Payee",
					"transaction_category_name":   "Test ERC20 Wrapper Transaction Category",
					"wallet_address":              "0x1234567890123456789012345678901234567890",
					"address_type":                "erc20_wrapper",
					"chain_name":                  "ethereum",
					"token_address":               "0x4567890123456789012345678901234567890",
					"base_token_address_function": "0x7890123456789012345678901234567890",
				}

				yamlBytes, err := yaml.Marshal(map[string]any{
					"ynab_accounts": []any{erc20WrapperAccountYAML},
				})
				Expect(err).ToNot(HaveOccurred(), "serializing the ERC20 wrapper account should not fail")

				syncConfig, err := config.FromYAML(bytes.NewBuffer(yamlBytes))
				Expect(err).ToNot(HaveOccurred(), "deserializing the ERC20 wrapper account should not fail")

				Expect(syncConfig.Accounts).To(HaveLen(1), "there should be one ERC20 wrapper account")

				account := syncConfig.Accounts[0]

				Expect(account.GetAddressType()).To(Equal(config.AddressTypeERC20Wrapper), "the address type should be successfully parsed")

				erc20WrapperAccount, err := account.AsERC20WrapperAccount()
				Expect(err).ToNot(HaveOccurred(), "resolving the ERC20 wrapper account should not fail")

				Expect(erc20WrapperAccount.AccountName).To(Equal("Test ERC20 Wrapper Account"), "the account name should be successfully parsed")
				Expect(erc20WrapperAccount.PayeeName).To(Equal("Test ERC20 Wrapper Payee"), "the payee name should be successfully parsed")
				Expect(erc20WrapperAccount.TransactionCategoryName).To(Equal("Test ERC20 Wrapper Transaction Category"), "the transaction category name should be successfully parsed")
				Expect(erc20WrapperAccount.WalletAddress).To(Equal("0x1234567890123456789012345678901234567890"), "the wallet address should be successfully parsed")
				Expect(erc20WrapperAccount.ChainName).To(Equal("ethereum"), "the chain name should be successfully parsed")
				Expect(erc20WrapperAccount.TokenAddress).To(Equal("0x4567890123456789012345678901234567890"), "the token address should be successfully parsed")
				Expect(erc20WrapperAccount.BaseTokenAddressFunction).To(Equal("0x7890123456789012345678901234567890"), "the base token address function should be successfully parsed")
			})
		})
	})
})
