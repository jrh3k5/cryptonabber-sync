# cryptonabber-sync
A utility used to synchronize YNAB account balances to token balances

## Usage

### Prerequisites

* You must have a [YNAB](https://ynab.com) account with a budget and accounts to set up
* You must have a registered OAuth client ID and secret as described [here](https://api.ynab.com/#oauth-applications).

### Executing the Program

You can either supply the OAuth credentials interactively by executing this application as:

```
/cryptonabber-sync --interactive
```

...or you can supply the OAuth credentials non-interactively by executing this application as:

```
/cryptonabber-sync --oauth-client-id=<client ID> --oauth-client-secret=<client secret>
```

You can provide the following optional arguments:

* `--dry-run`: specify this if you would like this tool to calculate balances, but not actually persist them to YNAB
* `--file`: by default, this application looks for a file called `config.yaml` in the local directory; if you would like to use a different filename or location, you can use this parameter to specify that

### Configuration

#### Configuration File Format

Your configuration file should have the following configuration:

```
ynab_budget_name: "<the name of the budget for which accounts are to be updated>"
rpc_configurations:
  - rpc_url: "<the URL of the RPC node>"
    chain_name: "<a shorthand reference for the RPC node; used in your YNAB account config, below>"
    chain_type: "evm"
ynab_accounts:
  - <configuration varies; see below>
```

##### YNAB Account Configuration

This tool supports the following types of assets to be evaluated:

* **ERC20**: a standalone token that implements the ERC20 standard
* **ERC4626 Vault**: a vault that implements the ERC4626 standard
* **ERC20 Wrapper**: a wrapper token that, through a function on the contract, expresses what the underlying wrapped asset is

###### ERC20 YNAB Account Configuration

The configuration block for evaluating the balance of an ERC20 token looks like:

```
- account_name: "<the name of the account in YNAB to be updated>"
  payee_name: "<the payee name to be recorded in YNAB>"
  transaction_category_name: "<the budget category under which the transaction is to be written in YNAB>"
  wallet_address: "<the address of the wallet that holds the ass
  address_type: "erc20"
  chain_name: "<the chain name of the RPC node to be used to read this token's information>"
  token_address: "<the address of the ERC20 asset>"
```

###### ERC4626 Vault YNAB Account Configuration

The configuration block for evaluating the balance of an ERC4626 vault token looks like:

```
- account_name: "<the name of the account in YNAB to be updated>"
  payee_name: "<the payee name to be recorded in YNAB>"
  transaction_category_name: "<the budget category under which the transaction is to be written in YNAB>"
  wallet_address: "<the address of the wallet that holds the ass
  address_type: "erc4626"
  chain_name: "<the chain name of the RPC node to be used to read this token's information>"
  vault_address: "<the address of the ERC4626 vault asset>"
```

You can also provide the following optional fields:

* `balance_function`: the optional name of the function to be called to get the wallet's balance of the vault asset; if not specified, the application will call `balanceOf`
* `backing_asset`: by default, the applicaiton will try to call `asset` on the given vault address to determine the underlying asset; however, you can specify a `backing_asset` element like so to describe what asset is backing the vault; for assets without a contract address, such as Ether on Ethereum, the `contract_address` field should be set to `""`:
```
backing_asset:
  contract_address: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
```

###### ERC20 Wrapper YNAB Account Configuration

The configuration block for evaluating the balance of an ERC20 wrapper token looks like:

```
- account_name: "<the name of the account in YNAB to be updated>"
  payee_name: "<the payee name to be recorded in YNAB>"
  transaction_category_name: "<the budget category under which the transaction is to be written in YNAB>"
  wallet_address: "<the address of the wallet that holds the ass
  address_type: "erc20_wrapper"
  chain_name: "<the chain name of the RPC node to be used to read this token's information>"
  token_address: "<the address of the ERC20 wrapper asset>"
  base_token_address_function: "<the name of the function to be called to get the address of the asset wrapped by this token>"
```

##### Fiat Value Evaluation

This tool currently only supports conversion of asset values into USD. This tool attempts to resolve an asset's quote from Coingecko using the asset's address (or the address of the underlying asset, for cases such as vaults or wrapping tokens).

## Privacy Policy

This application does not persist any information given to this application. It only uses the access granted to your account within YNAB to update account balances within YNAB to reflect ochain balances using the configuration you provide to the tool.

No data given to this application or read from YNAB is shared with any third parties.
