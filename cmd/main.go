package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	"github.com/jrh3k5/cryptonabber-sync/v2/coingecko"
	"github.com/jrh3k5/cryptonabber-sync/v2/config"
	rpcconfig "github.com/jrh3k5/cryptonabber-sync/v2/config/rpc"
	"github.com/jrh3k5/cryptonabber-sync/v2/evm"
	"github.com/jrh3k5/cryptonabber-sync/v2/token"
	"github.com/jrh3k5/cryptonabber-sync/v2/token/balance"
	"github.com/jrh3k5/oauth-cli/pkg/auth"
)

func main() {
	ctx := context.Background()

	dryRun := dryRunEnabled()
	if dryRun {
		fmt.Println("Dry run is enabled; no writes will be made to YNAB")
	}

	oauthToken, err := auth.DefaultGetOAuthToken(ctx,
		"https://app.ynab.com/oauth/authorize",
		"https://api.ynab.com/oauth/token")
	if err != nil {
		panic(fmt.Sprintf("failed to get OAuth token: %v", err))
	}

	configFileLocation := getConfigFile()

	fmt.Printf("Reading configuration from '%s'\n", configFileLocation)

	syncConfig, err := config.FromFile(configFileLocation)
	if err != nil {
		panic(fmt.Sprintf("failed to read configuration file: %v", err))
	}

	ynabURL, err := url.Parse("https://api.ynab.com/v1/")
	if err != nil {
		// ??? how?
		panic(fmt.Sprintf("unable to parse hard-coded YNAB URL: %v", err))
	}

	httpClient := http.DefaultClient
	ynabClient := ynab.NewClient(ynabURL, httpClient, oauthToken.AccessToken)

	budget, err := getBudget(syncConfig.BudgetName, ynabClient)
	if err != nil {
		panic(fmt.Sprintf("failed to get budget: %v", err))
	}

	categoryGroups, err := ynabClient.CategoriesService.List(budget.Id)
	if err != nil {
		panic(fmt.Sprintf("failed to list categories: %v", err))
	}

	accounts, err := ynabClient.AccountsService.List(budget.Id)
	if err != nil {
		panic(fmt.Sprintf("failed to list accounts: %v", err))
	} else if len(accounts) == 0 {
		panic("no accounts found in budget")
	}

	coingeckoQuoteResolver := coingecko.NewHTTPQuoteResolver(httpClient)
	assetPlatformIDResolver := coingecko.NewSimpleAssetPlatformIDResolver()
	rpcConfigurationResolver := rpcconfig.NewDefaultConfigurationResolver(syncConfig.RPCConfigurations)

	erc20BalanceFetcher := balance.NewERC20Fetcher(rpcConfigurationResolver, httpClient)
	erc4626BalanceFetcher := balance.NewERC4262Fetcher(rpcConfigurationResolver, httpClient)
	erc20WrapperBalanceFetcher := balance.NewERC20WrapperFetcher(erc20BalanceFetcher)

	erc20AssetResolver := token.NewERC20AssetResolver()
	erc4626AssetResolver := token.NewERC4626AssetResolver(rpcConfigurationResolver, httpClient)
	erc20WrapperAssetResolver := token.NewERC20WrapperAssetResolver(rpcConfigurationResolver, httpClient)

	chainIDFetcher := evm.NewJSONRPCChainIDFetcher(rpcConfigurationResolver, httpClient)
	decimalsResolver := token.NewRPCDecimalsResolver(rpcConfigurationResolver, httpClient)

	accountChangeSummaries := make(map[string]*changeSummary)

	for accountIndex, account := range syncConfig.Accounts {
		addressType, err := account.GetAddressType()
		if err != nil {
			panic(fmt.Sprintf("failed to resolve address type for account at index %d: %v", accountIndex, err))
		}

		var tokenAddress *string
		var tokenBalance *big.Int
		var syncableAccount config.SyncableAccount
		var onchainAsset config.OnchainAsset
		switch addressType {
		case config.AddressTypeERC20:
			erc20Account, err := account.AsERC20Account()
			if err != nil {
				panic(fmt.Sprintf("failed to resolve ERC20 account at index %d: %v", accountIndex, err))
			}

			tokenAddress, err = erc20AssetResolver.ResolveAssetAddress(ctx, erc20Account)
			if err != nil {
				panic(fmt.Sprintf("failed to resolve token address for ERC20 account '%s': %v", syncableAccount.AccountName, err))
			}

			tokenBalance, err = erc20BalanceFetcher.FetchBalance(ctx, erc20Account)
			if err != nil {
				panic(fmt.Sprintf("failed to retrieve balance of ERC20 token '%s' for address '%s': %v", erc20Account.TokenAddress, erc20Account.WalletAddress, err))
			}

			syncableAccount = erc20Account.SyncableAccount
			onchainAsset = erc20Account.OnchainAsset
		case config.AddressTypeERC4626:
			erc4626Account, err := account.AsERC4626Account()
			if err != nil {
				panic(fmt.Sprintf("failed to resolve ERC4626 account at index %d: %v", accountIndex, err))
			}

			tokenAddress, err = erc4626AssetResolver.ResolveAssetAddress(ctx, erc4626Account)
			if err != nil {
				panic(fmt.Sprintf("failed to resolve token address for ERC4626 account '%s': %v", syncableAccount.AccountName, err))
			}

			tokenBalance, err = erc4626BalanceFetcher.FetchBalance(ctx, erc4626Account)
			if err != nil {
				panic(fmt.Sprintf("failed to retrieve balance of ERC4626 vault '%s' for address '%s': %v", erc4626Account.VaultAddress, erc4626Account.WalletAddress, err))
			}

			syncableAccount = erc4626Account.SyncableAccount
			onchainAsset = erc4626Account.OnchainAsset
		case config.AddressTypeERC20Wrapper:
			erc20WrapperAccount, err := account.AsERC20WrapperAccount()
			if err != nil {
				panic(fmt.Sprintf("failed to resolve ERC20Wrapper account at index %d: %v", accountIndex, err))
			}

			tokenAddress, err = erc20WrapperAssetResolver.ResolveAssetAddress(ctx, erc20WrapperAccount)
			if err != nil {
				panic(fmt.Sprintf("failed to resolve token address for ERC20Wrapper account '%s': %v", syncableAccount.AccountName, err))
			}

			tokenBalance, err = erc20WrapperBalanceFetcher.FetchBalance(ctx, erc20WrapperAccount)
			if err != nil {
				panic(fmt.Sprintf("failed to retrieve balance of ERC20Wrapper token '%s' for address '%s': %v", erc20WrapperAccount.TokenAddress, erc20WrapperAccount.WalletAddress, err))
			}

			syncableAccount = erc20WrapperAccount.SyncableAccount
			onchainAsset = erc20WrapperAccount.OnchainAsset
		default:
			panic(fmt.Sprintf("unsupported address type '%s' for account '%s'", addressType, syncableAccount.AccountName))
		}

		var categoryID string
		for _, categoryGroup := range categoryGroups {
			for _, category := range categoryGroup.Categories {
				if category.Name == syncableAccount.TransactionCategoryName {
					categoryID = category.Id
					break
				}
			}
		}

		if categoryID == "" {
			panic(fmt.Sprintf("No category '%s' found in budget for account '%s'", syncableAccount.TransactionCategoryName, syncableAccount.AccountName))
		}

		var dollarRate int64
		var centsRate float64

		chainID, chainIDErr := chainIDFetcher.GetChainID(ctx, onchainAsset.ChainName)
		if chainIDErr != nil {
			panic(fmt.Sprintf("failed to retrieve chain ID for account '%s': %v", syncableAccount.AccountName, chainIDErr))
		}

		assetPlatformID, assetPlatformIDErr := assetPlatformIDResolver.ResolveForChainID(ctx, chainID)
		if assetPlatformIDErr != nil {
			panic(fmt.Sprintf("failed to retrieve asset platform ID for account '%s': %v", syncableAccount.AccountName, assetPlatformIDErr))
		}

		if tokenAddress != nil {
			var hasQuote bool
			var quoteErr error
			dollarRate, centsRate, hasQuote, quoteErr = coingeckoQuoteResolver.ResolveQuote(ctx, assetPlatformID, *tokenAddress)
			if quoteErr != nil {
				panic(fmt.Sprintf("failed to get quote for token address '%s': %v", *tokenAddress, quoteErr))
			} else if !hasQuote {
				panic(fmt.Sprintf("unable to resolve a quote for token address '%s'; please configure one explicitly for the account", *tokenAddress))
			}
		} else {
			var quoteErr error
			dollarRate, centsRate, quoteErr = coingeckoQuoteResolver.ResolveETHQuote(ctx)
			if quoteErr != nil {
				panic(fmt.Sprintf("unable to resolve quote for ETH: %v", quoteErr))
			}
		}

		tokenDecimals, err := decimalsResolver.ResolveDecimals(ctx, onchainAsset, tokenAddress)
		if err != nil {
			panic(fmt.Sprintf("failed to resolve token decimals for account '%s': %v", syncableAccount.AccountName, err))
		}

		currentBalance := balance.AsFiat(tokenBalance, tokenDecimals, dollarRate, centsRate) * 10 // YNAB stores cents as hundreds, not tens

		ynabAccount, err := getAccount(syncableAccount.AccountName, accounts)
		if err != nil {
			panic(fmt.Sprintf("failed to find account: %v", err))
		}

		if accountDiff := currentBalance - int64(ynabAccount.Balance); accountDiff != 0 {
			if !dryRun {
				updateAccount(ynabClient, budget.Id, ynabAccount.Id, categoryID, tokenBalance.Int64(), tokenDecimals, dollarRate, centsRate, accountDiff)
			}

			accountDiffCents := accountDiff % 1000
			accountDiffDollars := (accountDiff - accountDiffCents) / 1000
			accountDiffCents = (accountDiffCents - (accountDiffCents % 10)) / 10
			accountChangeSummaries[ynabAccount.Name] = &changeSummary{
				dollars: accountDiffDollars,
				cents:   accountDiffCents,
			}
		} else {
			accountChangeSummaries[ynabAccount.Name] = &changeSummary{}
		}

	}

	fmt.Println("================")
	fmt.Printf("Updated %d accounts:\n", len(accountChangeSummaries))

	accountNames := make([]string, 0, len(accountChangeSummaries))
	for accountName := range accountChangeSummaries {
		accountNames = append(accountNames, accountName)
	}
	sort.Strings(accountNames)

	for _, accountName := range accountNames {
		changeSummary, _ := accountChangeSummaries[accountName]
		absCents := int(math.Abs(float64(changeSummary.cents)))
		fmt.Printf("  %s: $%d.%02d\n", accountName, changeSummary.dollars, absCents)
	}
}

func dryRunEnabled() bool {
	for _, osArg := range os.Args {
		if strings.HasPrefix(osArg, "--dry-run") {
			return true
		}
	}

	return false
}

func getAccount(desiredAccountName string, accounts []ynab.Account) (*ynab.Account, error) {
	var accountNames []string
	for _, account := range accounts {
		accountName := account.Name
		accountNames = append(accountNames, accountName)
		if accountName == desiredAccountName {
			return &account, nil
		}
	}

	sort.Strings(accountNames)

	return nil, fmt.Errorf("no account found for name '%s'; available accounts are: ['%s']", desiredAccountName, strings.Join(accountNames, "', '"))
}

func getBudget(desiredBudgetName string, client *ynab.Client) (*ynab.BudgetSummary, error) {
	budgets, err := client.BudgetService.List()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve budgets: %w", err)
	} else if len(budgets) == 0 {
		return nil, errors.New("no budgets found")
	}

	var budgetNames []string
	for _, budget := range budgets {
		budgetName := budget.Name
		budgetNames = append(budgetNames, budgetName)
		if budgetName == desiredBudgetName {
			return &budget, nil
		}
	}

	return nil, fmt.Errorf("Budget '%s' not found; available budget(s) are: ['%s']", desiredBudgetName, strings.Join(budgetNames, "', '"))
}

func getConfigFile() string {
	for _, osArg := range os.Args {
		if strings.HasPrefix(osArg, "--file=") {
			return strings.TrimPrefix(osArg, "--file=")
		}
	}

	return "config.yaml"
}

func updateAccount(client *ynab.Client, budgetID string, accountID string, categoryID string, tokenBalance int64, tokenDecimals int, dollarRate int64, centsRate float64, deltaDecicents int64) error {
	dateString := time.Now().Format("2006-01-02")

	tokenFractions := tokenBalance % int64(math.Pow10(tokenDecimals))
	wholeTokens := (tokenBalance - tokenFractions) / int64(math.Pow10(tokenDecimals))
	formattedTokenBalance := fmt.Sprintf("%d.%s", wholeTokens, fmt.Sprintf("%d", tokenFractions)[:2])
	formattedRate := fmt.Sprintf("$%d.%s", dollarRate, fmt.Sprintf("%.2f", centsRate)[2:])
	formattedTime := time.Now().Format("03:04 PM MST")

	_, err := client.TransactionsService.Create(budgetID, &ynab.SaveTransaction{
		AccountId:  accountID,
		Date:       dateString,
		Amount:     int(deltaDecicents),
		PayeeName:  "Balance Adjustment",
		CategoryId: categoryID,
		Memo:       fmt.Sprintf("%s @ %s (executed %v)", formattedTokenBalance, formattedRate, formattedTime),
	})
	if err != nil {
		return fmt.Errorf("failed to create adjustment transaction: %w", err)
	}

	return nil
}

type changeSummary struct {
	dollars int64
	cents   int64
}
