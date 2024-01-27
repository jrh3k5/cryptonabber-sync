package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/jrh3k5/cryptonabber-sync/config"
	"github.com/jrh3k5/cryptonabber-sync/token/balance"
)

func main() {
	ctx := context.Background()

	var configFileLocation string
	flag.StringVar(&configFileLocation, "file", "", "the location of the file to be read in for configuration")

	flag.Parse()

	fmt.Printf("Reading configuration from '%s'\n", configFileLocation)

	syncConfig, err := config.FromFile(configFileLocation)
	if err != nil {
		panic(fmt.Sprintf("Failed to read configuration file: %v", err))
	}

	for _, account := range syncConfig.Accounts {
		balanceFetcher := balance.NewEVMFetcher(account.RPCURL)
		balance, err := balanceFetcher.FetchBalance(ctx, account.TokenAddress, account.WalletAddress)
		if err != nil {
			panic(fmt.Sprintf("Failed to retrieve balance of token '%s' for address '%s': %v", account.TokenAddress, account.WalletAddress, err))
		}
		fmt.Printf("Balance of token '%s' for wallet '%s' is %v\n", account.TokenAddress, account.WalletAddress, balance)
	}
}
