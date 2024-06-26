<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the GitHub issue reference in the following format:

* (<tag>) \#<issue-number> message

Tag must include `sql` if having any changes relate to schema

The issue numbers will later be link-ified during the release process,
so you do not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Schema Breaking" for breaking SQL Schema.
"API Breaking" for breaking API.

If any PR belong to multiple types of change, reference it into all types with only ticket id, no need description (convention)

Ref: https://keepachangelog.com/en/1.0.0/
-->

<!--
Templates for Unreleased:

## Unreleased

### Features

### Improvements

### Bug Fixes

### Schema Breaking

### API Breaking
-->

# Changelog

## Unreleased

## v1.2.4 - 2024-06-03

### Improvements

- (backend) [#24](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/24) Expose method `GetBankDenomsMetadata`

## v1.2.3 - 2024-05-06

### Bug Fixes

- (rpc) [#23](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/23) Fix cache staking validators not reloaded correctly

## v1.2.2 - 2024-05-06

### Improvements

- (rpc) [#22](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/22) `be_getValidators` returns more necessary information

## v1.2.1 - 2024-05-06

### Features

- (rpc) [#20](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/20) Add new API `be_getRecentBlocks` to fetch recent blocks

### Improvements

- (rpc) [#18](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/18) `be_getBlockByNumber` returns proposer
- (rpc) [#19](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/19) `be_getChainInfo` returns Be-RPC version

### Bug Fixes

- (rpc) [#21](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/21) Limit page size API `be_getRecentBlocks`

## v1.1.7 - 2024-05-05

### Improvements

- (rpc) [#17](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/17) Add query module params for `auth` and `ibc-transfer` module

### Bug Fixes

- (rpc) [#16](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/16) Fix query gov module params

## v1.1.6 - 2024-05-05

### Improvements

- (rpc) [#15](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/15) Returns EVM/Wasm tx action & sig info in `be_getBlockByNumber` response

## v1.1.5 - 2024-05-05

### Bug Fixes

- (rpc) [#14](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/14) Fix RPC crash due to wrong tx consumption logic (p2)

## v1.1.4 - 2024-05-03

### Bug Fixes

- (rpc) [#13](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/13) Fix validators not listed correctly

## v1.1.3 - 2024-05-03

### Bug Fixes

- (rpc) [#12](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/12) Fix RPC crash due to wrong tx consumption logic

## v1.1.2 - 2024-04-20

### Improvements

- (rpc) [#7](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/7) Improve IBC messages translation & add IBC packet info into message parser
- (rpc) [#8](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/8) + [#9](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/9) Add IBC packet information into `be_getTransactionsInBlockRange`
- (rpc) [#10](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/10) Refactor export function to reduce chance of crashing entire handler
- (rpc) [#11](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/11) Return transaction value in `be_getTransactionsInBlockRange` response

## v1.1.1 - 2024-04-14

### Improvements

- (rpc) [#6](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/6) `be_getTransactionsInBlockRange` response include some wasm tx information

## v1.1.0 - 2024-04-14

### Improvements

- (rpc) [#3](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/3) `be_getAccount` response include account proto type
- (rpc) [#5](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/5) `be_getTransactionsInBlockRange` response include some evm tx information

### API Breaking

- (rpc) [#4](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/4) Contract token balance involvers extractor

## v1.0.3 - 2024-04-08

### Features

- (rpc) [#2](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/2) Add endpoint `be_getLatestBlockNumber`

## v1.0.2 - 2024-04-05

### Improvements

- (parser) [#1](https://github.com/bcdevtools/block-explorer-rpc-cosmos/pull/1) Add message signer into involvers list
