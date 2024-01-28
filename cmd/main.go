package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	"github.com/jrh3k5/cryptonabber-sync/config"
	"github.com/jrh3k5/cryptonabber-sync/token/balance"
)

func main() {
	ctx := context.Background()

	var accessToken string
	flag.StringVar(&accessToken, "access-token", "", "the personal access token to be used to communicate with YNAB")

	var configFileLocation string
	flag.StringVar(&configFileLocation, "file", "", "the location of the file to be read in for configuration")

	flag.Parse()

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
	ynabClient := ynab.NewClient(ynabURL, http.DefaultClient, accessToken)

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

	accountChangeSummaries := make(map[string]*changeSummary)

	for _, account := range syncConfig.Accounts {
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

		balanceFetcher := balance.NewEVMFetcher(account.RPCURL, http.DefaultClient)
		balance, err := balanceFetcher.FetchBalance(ctx, account.TokenAddress, account.WalletAddress)
		if err != nil {
			panic(fmt.Sprintf("failed to retrieve balance of token '%s' for address '%s': %v", account.TokenAddress, account.WalletAddress, err))
		}

		tokenDecimals := account.TokenDecimals

		var dollars int64
		var cents int64
		switch tokenDecimals {
		case 0:
			dollars = balance.Int64()
		default:
			balanceInt := balance.Int64()
			cents = balanceInt % int64(math.Pow10(tokenDecimals))
			dollars = (balanceInt - cents) / int64(math.Pow10(tokenDecimals))

			if tokenDecimals > 2 {
				centsDivisor := int64(math.Pow10(tokenDecimals - 2))
				cents = (cents - (cents % int64(centsDivisor))) / centsDivisor
			}
		}

		account, err := getAccount(account.AccountName, accounts)
		if err != nil {
			panic(fmt.Sprintf("failed to find account: %v", err))
		}

		currentBalance := (dollars * 1000) + (cents * 10) // YNAB stores cents as hundreds, not tens
		if accountDiff := currentBalance - int64(account.Balance); accountDiff != 0 {
			updateAccount(ynabClient, budget.Id, account.Id, categoryID, accountDiff)

			accountDiffCents := accountDiff % 1000
			accountDiffDollars := (accountDiff - accountDiffCents) / 1000
			accountDiffCents = (accountDiffCents - (accountDiffCents % 10)) / 10
			accountChangeSummaries[account.Name] = &changeSummary{
				dollars: accountDiffDollars,
				cents:   accountDiffCents,
			}
		} else {
			accountChangeSummaries[account.Name] = &changeSummary{}
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

func updateAccount(client *ynab.Client, budgetID string, accountID string, categoryID string, deltaDecicents int64) error {
	dateString := time.Now().Format("2006-01-02")
	_, err := client.TransactionsService.Create(budgetID, &ynab.SaveTransaction{
		AccountId:  accountID,
		Date:       dateString,
		Amount:     int(deltaDecicents),
		PayeeName:  "Balance Adjustment",
		CategoryId: categoryID,
		Memo:       fmt.Sprintf("Balance adjustment executed %v", time.Now().Format(time.RFC3339)),
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
