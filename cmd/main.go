package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	"github.com/jrh3k5/cryptonabber-sync/v2/coingecko"
	"github.com/jrh3k5/cryptonabber-sync/v2/config"
	"github.com/jrh3k5/cryptonabber-sync/v2/evm"
	"github.com/jrh3k5/cryptonabber-sync/v2/token/balance"
	"github.com/jrh3k5/oauth-cli/pkg/auth"
)

func main() {
	ctx := context.Background()

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
	ynabClient := ynab.NewClient(ynabURL, http.DefaultClient, oauthToken.AccessToken)

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

	coingeckoQuoteResolver := coingecko.NewHTTPQuoteResolver(http.DefaultClient)
	assetPlatformIDResolver := coingecko.NewSimpleAssetPlatformIDResolver()

	accountChangeSummaries := make(map[string]*changeSummary)

	for _, account := range syncConfig.Accounts {
		tokenDecimals, err := account.GetTokenDecimals()
		if err != nil {
			panic(fmt.Sprintf("unable to fetch account's token decimals: %v", err))
		}

		var categoryID string
		for _, categoryGroup := range categoryGroups {
			for _, category := range categoryGroup.Categories {
				if category.Name == account.TransactionCategoryName {
					categoryID = category.Id
					break
				}
			}
		}

		if categoryID == "" {
			panic(fmt.Sprintf("No category '%s' found in budget for account '%s'", account.TransactionCategoryName, account.AccountName))
		}

		var balanceFetcher balance.Fetcher
		switch account.GetAddressType() {
		case config.AddressTypeERC20:
			balanceFetcher = balance.NewERC20Fetcher(account.RPCURL, http.DefaultClient)
		case config.AddressTypeStakewiseVault:
			balanceFetcher = balance.NewStakewiseVaultFetcher(account.RPCURL, http.DefaultClient)
		default:
			panic(fmt.Sprintf("unhandled address type: %v", account.GetAddressType()))
		}

		tokenBalance, err := balanceFetcher.FetchBalance(ctx, account.TokenAddress, account.WalletAddress)
		if err != nil {
			panic(fmt.Sprintf("failed to retrieve balance of token '%s' for address '%s': %v", account.TokenAddress, account.WalletAddress, err))
		}

		var dollarRate int64
		var centsRate float64
		if account.HasQuote() {
			var quoteErr error
			dollarRate, centsRate, _, quoteErr = account.GetQuote()
			if quoteErr != nil {
				panic(fmt.Sprintf("failed to retrieve configured quote from account: %v", quoteErr))
			}
		} else {
			chainID, chainIDErr := evm.NewJSONRPCChainIDFetcher(account.RPCURL, http.DefaultClient).GetChainID(ctx)
			if chainIDErr != nil {
				panic(fmt.Sprintf("failed to retrieve chain ID for account '%s': %v", account.AccountName, chainIDErr))
			}

			assetPlatformID, assetPlatformIDErr := assetPlatformIDResolver.ResolveForChainID(ctx, chainID)
			if assetPlatformIDErr != nil {
				panic(fmt.Sprintf("failed to retrieve asset platform ID for account '%s': %v", account.AccountName, assetPlatformIDErr))
			}

			switch account.GetAddressType() {
			case config.AddressTypeERC20:
				var hasQuote bool
				var quoteErr error
				dollarRate, centsRate, hasQuote, quoteErr = coingeckoQuoteResolver.ResolveQuote(ctx, assetPlatformID, account.TokenAddress)
				if quoteErr != nil {
					panic(fmt.Sprintf("failed to get quote for token address '%s': %v", account.TokenAddress, quoteErr))
				} else if !hasQuote {
					panic(fmt.Sprintf("unable to resolve a quote for token address '%s'; please configure one explicitly for the account", account.TokenAddress))
				}
			case config.AddressTypeStakewiseVault:
				var quoteErr error
				dollarRate, centsRate, quoteErr = coingeckoQuoteResolver.ResolveETHQuote(ctx)
				if quoteErr != nil {
					panic(fmt.Sprintf("unable to resolve quote for ETH: %v", quoteErr))
				}
			}
		}

		currentBalance := balance.AsFiat(tokenBalance, tokenDecimals, dollarRate, centsRate) * 10 // YNAB stores cents as hundreds, not tens

		ynabAccount, err := getAccount(account.AccountName, accounts)
		if err != nil {
			panic(fmt.Sprintf("failed to find account: %v", err))
		}

		if accountDiff := currentBalance - int64(ynabAccount.Balance); accountDiff != 0 {
			updateAccount(ynabClient, budget.Id, ynabAccount.Id, categoryID, tokenBalance.Int64(), tokenDecimals, dollarRate, centsRate, accountDiff)

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
	for accountName, changeSummary := range accountChangeSummaries {
		fmt.Printf("  %s: $%d.%02d\n", accountName, changeSummary.dollars, changeSummary.cents)
	}
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
