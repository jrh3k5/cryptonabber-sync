# cryptonabber-sync
A utility used to synchronize YNAB account balances to token balances

## Usage

Execute the compiled binary with a `--file` parameter that points the application at the configuration file it should use - e.g.:

```
/cryptonabber-sync --file="nabber.yaml"
```

## Configuration

### Prerequisites

To configure this tool, you must have obtained the following:

* A [YNAB](https://ynab.com) account with a budget and accounts to set up
* A [YNAB personal access token](https://api.ynab.com/#personal-access-tokens)

Your configuration file should have the following configuration:

```
ynab_pat: <your YNAB personal access token>
ynab_budget_name: <the name of the budget for which accounts are to be updated>
ynab_accounts:
  - account_name: <the name of the account whose balance is to be updated>
    wallet_address: <the address of the wallet whose token balance is to be used to update the balance of the account>
    token_address: <the address of the token whose balance is to be calculated>
    token_decimals: <the decimals value of the token>
    rpc_url: <the URL of the node to be used to query for wallet balances>
```

Note that this only syncs the balances of the account to the balances of the configured token and _will not_ convert it to a fiat currency value.

Thus, this tool is best used to synchronize your accounts to a token that is pegged to the value of the account's fiat currency.