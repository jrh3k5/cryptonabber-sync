package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/jrh3k5/cryptonabber-sync/v2/config/rpc"
	"gopkg.in/yaml.v3"
)

type AddressType string

type AccountProperties map[string]any

const (
	AddressTypeERC20    AddressType = "erc20"         // describes an ERC20 token
	AddressTypeERC4626  AddressType = "erc4626"       // describes an ERC4626 vault
	AddressTypeCompound AddressType = "erc20_wrapper" // describes an ERC20 wrapper

	fieldAccountName             = "account_name"
	fieldAddressType             = "address_type"
	fieldBaseTokenAddresFunction = "base_token_address_function"
	fieldPayeeName               = "payee_name"
	fieldTokenAddress            = "token_address"
	fieldTransactionCategoryName = "transaction_category_name"
	fieldVaultAddress            = "vault_address"
	fieldWalletAddress           = "wallet_address"
)

// FromFile builds a SyncConfig out of the contents of a YAML file at the given location.
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

// SyncConfig is the overall configuration for the application.
type SyncConfig struct {
	BudgetName        string              `yaml:"ynab_budget_name"`
	Accounts          []AccountProperties `yaml:"ynab_accounts"`
	RPCConfigurations []rpc.Configuration `yaml:"rpc_configurations"`
}

// GetAddressType resolves the type of the address represented by the account properties
func (a AccountProperties) GetAddressType() (AddressType, error) {
	addressTypeString, hasProp, err := a.stringProperty(fieldAddressType)
	if err != nil {
		return AddressTypeERC20, fmt.Errorf("unable to resolve address type: %w", err)
	} else if !hasProp {
		return AddressTypeERC20, nil
	}

	return AddressType(addressTypeString), nil
}

// AsERC20Account resolves the account properties into an ERC20 account
func (a AccountProperties) AsERC20Account() (*ERC20Account, error) {
	addressType, err := a.GetAddressType()
	if err != nil {
		return nil, err
	} else if addressType != AddressTypeERC20 {
		return nil, fmt.Errorf("invalid address type: %s", addressType)
	}

	syncableAccount, err := a.asSyncableAccount()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve syncable account: %w", err)
	}

	onchainWallet, err := a.asOnchainWallet()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve onchain wallet: %w", err)
	}

	onChainAsset, err := a.asOnchainAsset()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve onchain asset: %w", err)
	}

	tokenAddress, hasTokenAddress, err := a.stringProperty(fieldTokenAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve token address: %w", err)
	} else if !hasTokenAddress {
		return nil, errors.New("token address is required")
	}

	return &ERC20Account{
		SyncableAccount: *syncableAccount,
		OnchainWallet:   *onchainWallet,
		OnchainAsset:    *onChainAsset,
		TokenAddress:    tokenAddress,
	}, nil
}

// AsERC4626Account resolves the account properties into an ERC4626 account
func (a AccountProperties) AsERC4626Account() (*ERC4626Account, error) {
	addressType, err := a.GetAddressType()
	if err != nil {
		return nil, err
	} else if addressType != AddressTypeERC4626 {
		return nil, fmt.Errorf("invalid address type: %s", addressType)
	}

	syncableAccount, err := a.asSyncableAccount()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve syncable account: %w", err)
	}

	onchainWallet, err := a.asOnchainWallet()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve onchain wallet: %w", err)
	}

	onchainAsset, err := a.asOnchainAsset()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve onchain asset: %w", err)
	}

	vaultAddress, hasVaultAddress, err := a.stringProperty(fieldVaultAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve vault address: %w", err)
	} else if !hasVaultAddress {
		return nil, errors.New("vault address is required")
	}

	return &ERC4626Account{
		SyncableAccount: *syncableAccount,
		OnchainWallet:   *onchainWallet,
		OnchainAsset:    *onchainAsset,
		VaultAddress:    vaultAddress,
	}, nil
}

// AsERC20WrapperAccount resolves the account properties into an ERC20 wrapper account
func (a AccountProperties) AsERC20WrapperAccount() (*ERC20WrapperAccount, error) {
	addressType, err := a.GetAddressType()
	if err != nil {
		return nil, err
	} else if addressType != AddressTypeCompound {
		return nil, fmt.Errorf("invalid address type: %s", addressType)
	}

	erc20Account, err := a.AsERC20Account()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve erc20 account: %w", err)
	}

	baseTokenAddressFunctionName, hasBaseTokenAddressFunctionName, err := a.stringProperty(fieldBaseTokenAddresFunction)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve base token address function name: %w", err)
	} else if !hasBaseTokenAddressFunctionName {
		return nil, errors.New("base token address function name is required")
	}

	return &ERC20WrapperAccount{
		ERC20Account:             *erc20Account,
		BaseTokenAddressFunction: baseTokenAddressFunctionName,
	}, nil
}

func (a AccountProperties) asOnchainAsset() (*OnchainAsset, error) {
	chainName, hasChainName, err := a.stringProperty("chain_name")
	if err != nil {
		return nil, fmt.Errorf("unable to resolve chain name: %w", err)
	} else if !hasChainName {
		return nil, errors.New("chain name is required")
	}

	return &OnchainAsset{
		ChainName: chainName,
	}, nil
}

func (a AccountProperties) asSyncableAccount() (*SyncableAccount, error) {
	accountName, hasAccountName, err := a.stringProperty(fieldAccountName)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve account name: %w", err)
	} else if !hasAccountName {
		return nil, errors.New("account name is required")
	}

	payeeName, hasPayeeName, err := a.stringProperty(fieldPayeeName)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve payee name: %w", err)
	} else if !hasPayeeName {
		return nil, errors.New("payee name is required")
	}

	transactionCategoryName, hasTransactionCategoryName, err := a.stringProperty(fieldTransactionCategoryName)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve transaction category name: %w", err)
	} else if !hasTransactionCategoryName {
		return nil, errors.New("transaction category name is required")
	}

	return &SyncableAccount{
		AccountName:             accountName,
		PayeeName:               payeeName,
		TransactionCategoryName: transactionCategoryName,
	}, nil
}

func (a AccountProperties) asOnchainWallet() (*OnchainWallet, error) {
	walletAddress, hasWalletAddress, err := a.stringProperty(fieldWalletAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve wallet address: %w", err)
	} else if !hasWalletAddress {
		return nil, errors.New("wallet address is required")
	}

	return &OnchainWallet{
		WalletAddress: walletAddress,
	}, nil
}

func (a AccountProperties) stringProperty(propertyName string) (string, bool, error) {
	propetyAny, hasProperty := a[propertyName]
	if !hasProperty {
		return "", false, nil
	}

	propertyString, isString := propetyAny.(string)
	if !isString {
		return "", false, fmt.Errorf("invalid property type for '%s': %v", propertyName, propetyAny)
	}

	return propertyString, true, nil
}

// OnchainAccount is a marker interface to declare when an instance of onchain account is needed
type OnchainAccount interface {
	isOnchainAccount()
}

type SyncableAccount struct {
	AccountName             string // the name of the account in YNAB
	PayeeName               string // the name of the payee to which the transction should be attributed in YNAB
	TransactionCategoryName string // the name of the YNAB category under which the transaction is to be classified
}

// OnchainWallet describes a wallet that is onchain.
type OnchainWallet struct {
	WalletAddress string // the address to which the asset belongs onchain
}

// OnchainAsset is the descriptor of an asset's onchain presence.
type OnchainAsset struct {
	ChainName string // the name of the string on which the asset resides, corresponding to an RPC configuration's chain name
}

// ERC20Account defines the properties needed to resolve the balance of an ERC20 token
type ERC20Account struct {
	SyncableAccount
	OnchainAsset
	OnchainWallet

	TokenAddress string // the address of the ERC20
}

func (*ERC20Account) isOnchainAccount() {}

// ERC4626Account defines the properties needed to resolve the balance of an ERC4626 vault
type ERC4626Account struct {
	SyncableAccount
	OnchainAsset
	OnchainWallet

	VaultAddress string // the address of the ERC4626 vault
}

func (*ERC4626Account) isOnchainAccount() {}

// ERC20WrapperAccount defines the properties needed to resolve the balance of an ERC20 wrapper
type ERC20WrapperAccount struct {
	ERC20Account

	BaseTokenAddressFunction string // the name of the function that returns the address of the base token wrapped by this token
}

func (*ERC20WrapperAccount) isOnchainAccount() {}
