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

### Configuration Parameters

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

#### Fiat Value Evaluation

This tool currently only supports conversion of asset values into USD.

By default, this tool attempts to resolve an asset's quote from Coingecko. However, your asset may not exist in Coingecko, or it may have the incorrect value. To remedy that, you can provide an optional `quote` value like so:

```
ynab_accounts:
  - account_name: ...
    quote: "1.25"
```

This will be the exchange rate used to determine the fiat value of the asset that is stored into YNAB.

#### Token Type

By default, all listed tokens are assumed to be ERC20 tokens. Their `balanceOf` methods will be used to determine the amount of token the wallet address has.

However, not all onchain balances are ERC20 balances. This tool supports an optional `token_type` field on the configuration like so:

```
ynab_accounts:
  - account_name: ...
    address_type: "erc20"
```

The supported types are:

* `erc20`: a token conforming to the ERC20 standard
* `stakewise_vault`: a vault providing liquid staking options via Stakewise
  * In this case, the `token_address` field should set to the contract address of the Vault's contract address - e.g., for the Gensis vault, use the address `0xAC0F906E433d58FA868F936E8A43230473652885`
  * When this is used, then the resolved quote will be Ethereum's value. Decimals will be assumed to be 18 and do not need to be configured.
