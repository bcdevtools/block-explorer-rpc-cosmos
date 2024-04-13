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
