# Cooked Books CLI

A double-entry accounting system that works for you. Talk to your books from the terminal.

## Install

```bash
# macOS (Apple Silicon)
curl -L https://github.com/joshdholtz/cooked-books-cli/releases/latest/download/cooked-books-darwin-arm64 -o /usr/local/bin/cooked-books
chmod +x /usr/local/bin/cooked-books

# macOS (Intel)
curl -L https://github.com/joshdholtz/cooked-books-cli/releases/latest/download/cooked-books-darwin-amd64 -o /usr/local/bin/cooked-books
chmod +x /usr/local/bin/cooked-books

# Linux
curl -L https://github.com/joshdholtz/cooked-books-cli/releases/latest/download/cooked-books-linux-amd64 -o /usr/local/bin/cooked-books
chmod +x /usr/local/bin/cooked-books
```

## Setup

1. Generate an API token at [app.cookedbooks.ai/integrations](https://app.cookedbooks.ai/integrations) (Developer tab)
2. Run `cooked-books login` and paste your token

## Commands

```
cooked-books context                         Financial overview
cooked-books transactions [--status=new]     List transactions
cooked-books accounts [--type=expense]       Chart of accounts
cooked-books pnl [--start=2025-01-01]        Profit & Loss
cooked-books balance-sheet [--date=2025-12-31]  Balance Sheet
cooked-books invoices [--status=sent]        List invoices
cooked-books contacts                        List contacts
cooked-books rules                           Categorization rules
```

## Environment Variables

- `COOKED_BOOKS_API_URL` — Override the API base URL (default: `https://api.cookedbooks.ai`)
- `COOKED_BOOKS_TOKEN` — API token (alternative to `cooked-books login`)

## Build from source

```bash
go build -o cooked-books .
```
