package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
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
	PersonalAccessToken string           `yaml:"ynab_pat"`
	BudgetName          string           `yaml:"ynab_budget_name"`
	Accounts            []*SyncedAccount `yaml:"ynab_accounts"`
}

type SyncedAccount struct {
	AccountName   string `yaml:"account_name"`
	WalletAddress string `yaml:"wallet_address"`
	TokenAddress  string `yaml:"token_address"`
	TokenDecimals int    `yaml:"token_decimals"`
	RPCURL        string `yaml:"rpc_url"`
}
