package config

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type AddressType string

const (
	AddressTypeERC20          AddressType = "erc20"
	AddressTypeStakewiseVault AddressType = "stakewise_vault"
)

func FromFile(fileLocation string) (*SyncConfig, error) {
	file, err := os.ReadFile(fileLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", fileLocation, err)
	}

	syncConfig := &SyncConfig{}
	if unmarshalErr := yaml.Unmarshal(file, syncConfig); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", unmarshalErr)
	}

	return syncConfig, nil
}

type SyncConfig struct {
	BudgetName string           `yaml:"ynab_budget_name"`
	Accounts   []*SyncedAccount `yaml:"ynab_accounts"`
}

type SyncedAccount struct {
	AccountName             string  `yaml:"account_name"`
	WalletAddress           string  `yaml:"wallet_address"`
	TokenAddress            string  `yaml:"token_address"`
	TokenDecimals           *int    `yaml:"token_decimals"`
	RPCURL                  string  `yaml:"rpc_url"`
	PayeeName               string  `yaml:"payee_name"`
	TransactionCategoryName string  `yaml:"transaction_category_name"`
	Quote                   *string `yaml:"quote"`
	AddressType             *string `yaml:"address_type"`
}

// GetAddressType gets the type of the address.
func (s *SyncedAccount) GetAddressType() AddressType {
	if s.AddressType == nil {
		return AddressTypeERC20
	}

	return AddressType(*s.AddressType)
}

// GetQuote gets, if configurd, the quote to be used to estimate the fiat value of the asset.
// The cents are returned as a ratio of cents to whole dollars to allow for assets that are worth
// less than 1 cent per whole token.
func (s *SyncedAccount) GetQuote() (int64, float64, bool, error) {
	if !s.HasQuote() {
		return 0, 0, false, nil
	}

	quoteStr := *s.Quote
	splitQuote := strings.Split(quoteStr, ".")
	dollars, parseErr := strconv.ParseInt(splitQuote[0], 10, 64)
	if parseErr != nil {
		return 0, 0, false, fmt.Errorf("failed to parse dollar value ('%s') of quote: %w", splitQuote[0], parseErr)
	}

	if len(splitQuote) == 1 {
		return dollars, 0, true, nil
	}

	centsString := splitQuote[1]
	cents, parseErr := strconv.ParseInt(centsString, 10, 64)
	if parseErr != nil {
		return 0, 0, false, fmt.Errorf("failed to parse cents value ('%s') of quote: %q", splitQuote[1], parseErr)
	}

	centsRatio := float64(cents) / math.Pow10(len(centsString))

	return dollars, centsRatio, true, nil
}

// GetTokenDecimals gets the token decimals for the asset to be matched in the account balance.
func (s *SyncedAccount) GetTokenDecimals() (int, error) {
	if s.TokenDecimals != nil {
		return *s.TokenDecimals, nil
	}

	if s.GetAddressType() == AddressTypeStakewiseVault {
		return 18, nil
	}

	return 0, errors.New("no decimals configured, and decimals cannot be inferred")
}

// HasQuote determines if this account has a quote value configured. It does not validate the configured value.
func (s *SyncedAccount) HasQuote() bool {
	return s.Quote != nil && *s.Quote != ""
}
