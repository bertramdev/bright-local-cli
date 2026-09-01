# bright-local-cli

A fast, read-only command-line client for the [BrightLocal Management API](https://developer.brightlocal.com/docs/management-apis).

It starts with the daily account lookups most useful from a terminal: locations, clients, and business categories. Responses are printed as formatted JSON, so they are easy to inspect or pipe into `jq`.

## Installation

Requires Go 1.26 or later.

```sh
go install github.com/bertramdev/bright-local-cli@latest
```

Or, from this directory:

```sh
go build -o bright-local .
```

## Authentication

Create a BrightLocal API key in the API page of your BrightLocal account, then provide it at runtime. Do not commit it.

```sh
export BRIGHTLOCAL_API_KEY='your-api-key'
```

You may instead pass `--api-key` to a single command.

## Usage

```sh
# List locations and clients
bright-local locations list --per-page 25 --query "Acme"
bright-local clients list --type client --page 2

# Read a single resource
bright-local locations get 12345
bright-local clients get 67890

# Business categories use an ISO 3166-1 alpha-2 country code
bright-local categories US --query "plumber"

# Access other documented read-only Management API endpoints
bright-local api get /manage/v1/clients --query page=2 --query num_per_page=50

# Read report and ranking data
bright-local rank-tracker reports history REPORT_ID --query page=1
bright-local search-grid runs list REPORT_ID KEYWORD_ID --query filter=active
bright-local search-grid rankings competitor REPORT_ID RUN_ID KEYWORD_ID COMPETITOR_ID
bright-local reputation reports reviews REPORT_ID
bright-local citation-builder list
bright-local reference time-options
```

## Commands

| Command | Description |
| --- | --- |
| `locations list` | List locations, with paging and free-text search. |
| `locations get <location-id>` | Get one location. |
| `clients list` | List clients, with paging, search, and type filtering. |
| `clients get <client-id>` | Get one client. |
| `categories <country>` | List business categories for a country. |
| `rank-tracker reports ...` | List reports and read rank-tracker history/results. |
| `search-grid reports\|runs\|rankings ...` | Read Local Search Grid reports, runs, competitors, and rankings. |
| `reputation reports ...` | List reports and read reputation reviews. |
| `citation-builder list\|get ...` | List or read Citation Builder campaigns. |
| `reference time-options` | Read available time options. |
| `reference white-label-profiles` | List white-label profiles. |
| `api get <path>` | GET another documented `/manage/v1/` endpoint. |

The CLI intentionally performs no create, update, or delete operations.

## Development

```sh
go test ./...
go run . --help
```
