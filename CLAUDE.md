# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AdGuard Home is a network-wide DNS server for blocking ads and tracking. It operates as a DNS forwarding server with filtering capabilities, a web UI dashboard, DHCP server, and automatic updates. This is a fork maintained by quydang04.

## Build & Development Commands

Prerequisites: Go 1.25+, Node.js v24.10.0+, npm 10.8+.

```sh
make init          # Set up git hooks (run once)
make               # Full build (deps + frontend + backend)
make quick-build   # Build without fetching deps (js-build + go-build)
```

### Backend (Go)

```sh
make go-deps       # Install Go dependencies
make go-build      # Build the Go binary (output: ./AdGuardHome)
make go-test       # Run all Go tests (race detector ON, count=2, shuffle=on, timeout=90s)
make go-lint       # Run all Go linters (requires: make go-tools first)
make go-tools      # Install linter binaries into ./bin/
make go-check      # go-tools + go-lint + go-test
make go-bench      # Run benchmarks
make go-fuzz       # Run fuzz tests
```

Run a single Go test:
```sh
go test -run TestName -race -count=1 ./internal/dnsforward/
```

### Frontend (React/TypeScript)

```sh
make js-deps       # npm ci
make js-build      # Production build (output embedded in ./build/)
make js-lint       # ESLint
make js-test       # Vitest
make js-test-e2e   # Playwright e2e tests
make js-typecheck  # tsc --noEmit
```

From `client/` directly:
```sh
npm run watch       # Dev build with watch
npm run watch:hot   # Webpack dev server with HMR
npm run test:watch  # Vitest in watch mode
```

### Release & Docker

```sh
make build-release SIGN=0 VERSION='1.0.0'
make build-docker
# Cross-compile: env GOOS='linux' GOARCH='arm64' make
```

### Other Linters

```sh
make txt-lint   # Text file linting
make md-lint    # Markdown linting
make sh-lint    # Shell script linting
```

## Architecture

### Two Entry Points (Build Tags)

- `main.go` (default): Current production code, calls `internal/home.Main()`. The frontend is embedded via `//go:embed build`.
- `main_next.go` (`-tags next`): Next-generation API rewrite, calls `internal/next/cmd.Main()`. Separate OpenAPI spec at `openapi/next.yaml`.

### Key Backend Packages (`internal/`)

| Package | Purpose |
|---|---|
| `home` | HTTP API handlers, app lifecycle, config loading, auth, TLS, service management |
| `dnsforward` | Core DNS forwarding server (wraps `dnsproxy`), request processing pipeline, access control, DNS64, upstream management |
| `filtering` | DNS request/response filtering engine, filter list management, hosts files, safe search, blocked services |
| `filtering/rulelist` | Rule list parsing and compilation |
| `filtering/safesearch` | Safe search enforcement for search engines |
| `filtering/hashprefix` | Hash-prefix lookups for safe browsing |
| `client` | Client identification, persistent/runtime client storage |
| `querylog` | DNS query logging with search |
| `stats` | DNS statistics collection and reporting |
| `dhcpd` | DHCP v4/v6 server (current implementation) |
| `dhcpsvc` | DHCP service (new implementation) |
| `configmigrate` | YAML config schema migrations (v1 through v31+) |
| `updater` | Auto-update mechanism |
| `next/` | Next-gen architecture: `cmd`, `configmgr`, `dnssvc`, `websvc` |

### Shared Utility Packages (`internal/agh*`)

- `agh` - Common interfaces (e.g., `ConfigModifier`)
- `aghalg` - Generic algorithms/data structures (sorted maps, null bools)
- `aghhttp` - HTTP helpers, JSON response utilities
- `aghnet` - Network utilities, hosts container, interface discovery
- `aghos` - OS abstraction, service management, file system
- `aghtest` - Test helpers, fake upstreams
- `aghtls` - TLS configuration helpers
- `aghslog` - Structured logging utilities

### Frontend (`client/`)

React 16 + Redux + TypeScript dashboard. Webpack-bundled, output goes to `build/` which is embedded in the Go binary. Uses i18next for localization with translations in `client/src/__locales/`.

### DNS Request Flow

1. DNS query arrives at `dnsforward.Server`
2. Before-request hooks run (access control, rate limiting)
3. Client identification (by IP, ClientID)
4. `filtering.DNSFilter` checks against filter lists, hosts files, safe search, safe browsing
5. DNS rewrites applied if matched
6. If not filtered, forwarded to configured upstream DNS servers via `dnsproxy`
7. Response processed (DNS64, CNAME filtering)
8. Query logged to `querylog`, stats recorded to `stats`

### Configuration

YAML config file with versioned schema. `configmigrate` handles upgrades between schema versions (each `vN.go` file is one migration step). Config version is tracked in the YAML file itself.

### OpenAPI

API specs live in `openapi/openapi.yaml` (current) and `openapi/next.yaml` (next-gen).

## Code Conventions

### Go

- **Banned imports**: `errors` (use `github.com/AdguardTeam/golibs/errors`), `log` (use `log/slog` + `slogutil`), `reflect`, `sort` (use `slices`), `unsafe`, `io/ioutil`
- **No underscores in Go filenames** except for build tags (`_linux.go`, `_test.go`, etc.)
- **HTTP methods**: Use `http.MethodGet` etc., never raw strings like `"GET"`
- **Formatting**: `gofumpt` with `--extra` flag
- **Linter suite**: govulncheck, gocyclo (max 10), gocognit (max 10-20 depending on package), ineffassign, unparam, errcheck, gosec, staticcheck (cross-OS matrix), shadow, fieldalignment, nilness, misspell
- **Test style**: Tests use `_internal_test.go` suffix (same package) or `_test.go` (external test package)
- **Errors**: Use `github.com/AdguardTeam/golibs/errors` for wrapping, prefer typed/sentinel errors
- **Logging**: Use `log/slog` and `github.com/AdguardTeam/golibs/logutil/slogutil`
- Follow [AdGuard Go code guidelines](https://github.com/AdguardTeam/CodeGuidelines/blob/master/Go/Go.md)

### Frontend

- ESLint with Airbnb config + Prettier
- Stylelint for CSS
- Vitest for unit tests, Playwright for e2e
