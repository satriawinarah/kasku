# Kasku — Family Money Manager

A lightweight personal finance tracker for families. Husband and wife can log income and expenses across multiple wallets (salary, savings, business, etc.) and review monthly spending together.

## Features

- **Family account** — one account groups all members and wallets
- **Multi-user** — each person logs in separately; entries are attributed to the recorder
- **Wallets** — unlimited wallets per family: salary, savings, business, general
- **Ledger** — credit (income) and debit (expense) entries with category, date, and note
- **Monthly view** — filter the ledger by month with income/expense summaries
- **Invite flow** — admin generates a link to share with their partner; no email server needed
- **Session auth** — secure cookie-based sessions, bcrypt passwords
- **Single binary** — the entire app compiles to one binary; no external dependencies at runtime

## Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.23 |
| HTTP router | Chi v5 |
| HTML templates | Templ (type-safe Go HTML) |
| Frontend interactivity | HTMX 2.x (via CDN) |
| Styling | Tailwind CSS (via CDN) |
| Database | SQLite (modernc.org/sqlite — pure Go, no CGo) |
| Sessions | alexedwards/scs with custom SQLite store |
| Passwords | bcrypt |

## Quick Start

### Prerequisites

- Go 1.23+

### 1. Install dev tools

```bash
make tools
```

This installs `templ` (template compiler) and `air` (live reload).

### 2. Generate templates and build

```bash
make build
./bin/kasku
```

Open [http://localhost:8080](http://localhost:8080) and register your family account.

### 3. Development mode (live reload)

Open two terminals:

```bash
# Terminal 1: watch and recompile templ files
make templ-watch

# Terminal 2: hot reload the Go binary
make dev
```

## Configuration

All configuration is via environment variables. Defaults work out of the box for local use.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `DB_PATH` | `./kasku.db` | SQLite database file path |
| `SESSION_SECRET` | *(insecure default)* | 32+ byte secret for session cookies — **change in production** |
| `APP_URL` | `http://localhost:8080` | Base URL used in invite links |

## Database

The database is a single SQLite file (`kasku.db` by default). To back it up, copy the file:

```bash
cp kasku.db kasku-backup-$(date +%Y%m%d).db
```

Migrations run automatically on startup. They are idempotent (`IF NOT EXISTS`).

## Usage Flow

1. **Register** at `/register` — create your family name and your admin account
2. **Create wallets** at `/wallets` — e.g., "Monthly Salary" (type: salary), "Emergency Fund" (type: savings)
3. **Add entries** at `/ledger` — record income (credit) and expenses (debit) with categories
4. **Invite partner** at `/family` — generates a link your partner uses to set their password
5. **View dashboard** — see the month's income, expenses, and net balance at a glance

## Makefile Targets

```
make build        # Compile to ./bin/kasku
make run          # Build and run
make dev          # Hot reload with air (requires make templ-watch running)
make templ-watch  # Watch and recompile templ files on change
make templ-gen    # One-shot templ generation
make test         # Run all tests
make fmt          # Format Go and templ files
make tidy         # Run go mod tidy
make clean        # Remove build artifacts
make build-prod   # Cross-compile optimised Linux binary
```

## Project Structure

See [docs/BACKEND.md](docs/BACKEND.md) for the Go package architecture and [docs/FRONTEND.md](docs/FRONTEND.md) for the template and HTMX patterns.
