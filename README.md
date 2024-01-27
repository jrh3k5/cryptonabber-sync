# cryptonabber-sync
A utility used to synchronize YNAB account balances to token balances

## Usage

Execute the compiled binary with two parameters:

* `--access-token`: the YNAB personal access token to be used to communicate with YNAB
* `--file`: the location of the configuration file the app should use

Example:

```
/cryptonabber-sync --access-token=FAKE507c706 --file="nabber.yaml"
```

## Configuration

### Prerequisites

To configure this tool, you must have obtained the following:

* A [YNAB](https://ynab.com) account with a budget and accounts to set up
* A [YNAB personal access token](https://api.ynab.com/#personal-access-tokens)

Your configuration file should have the following configuration:

```
ynab_budget_name: "<the name of the budget for which accounts are to be updated>"
ynab_accounts:
  - account_name: "<the name of the account whose balance is to be updated>"
    wallet_address: "<the address of the wallet whose token balance is to be used to update the balance of the account>"
    token_address: "<the address of the token whose balance is to be calculated>"
    token_decimals: <the decimals value of the token>
    rpc_url: "<the URL of the node to be used to query for wallet balances>"
    payee_name: "<the name of the payee to be entered into YNAB when the adjustment is made>"
    transaction_category_name: "<the name of the budget category under which the transaction is to be classified>"
```

Note that this only syncs the balances of the account to the balances of the configured token and _will not_ convert it to a fiat currency value.

Thus, this tool is best used to synchronize your accounts to a token that is pegged to the value of the account's fiat currency.