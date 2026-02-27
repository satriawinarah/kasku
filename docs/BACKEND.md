# Backend Architecture

## Overview

Kasku is a Go monolith. There is no separate API server and no JavaScript build step. All HTML is rendered server-side using the Templ template engine. HTMX handles dynamic page updates by requesting HTML fragments from the same Go server.

## Package Layout

```
internal/
├── config/        # Environment-based configuration
├── db/            # SQLite connection, pragmas, migration runner
│   └── migrations/
│       └── 001_initial.sql
├── model/         # Plain Go structs (family, user, wallet, ledger)
├── store/         # SQL query implementations (one file per table)
├── service/       # Business logic (auth, wallet, ledger)
├── session/       # Custom SCS SQLite session store
├── handler/       # HTTP handlers + middleware
└── router/        # Route definitions
```

## Dependency Flow

```
main.go
  └── config.Load()
  └── db.Open() → db.RunMigrations()
  └── session.NewManager()
  └── store.New()          ← wraps *sql.DB
  └── service.*            ← wraps stores
  └── handler.*            ← wraps services + stores + session
  └── router.New()         ← wires handlers to routes
```

There are no circular dependencies. Each layer only depends on the layer below it.

## Database

### Schema (4 tables + sessions)

| Table | Purpose |
|-------|---------|
| `sessions` | SCS session data (token, BLOB, expiry) |
| `families` | Top-level group — name only |
| `users` | Family members; `role` is `admin` or `member` |
| `wallets` | Named account with a running balance and type |
| `ledger` | Credit/debit entries; balance updated by triggers |

### Balance Accounting

`wallets.balance` is maintained automatically by two SQLite triggers:

- `ledger_after_insert` — adds (credit) or subtracts (debit) the entry amount
- `ledger_after_delete` — reverses the effect when an entry is deleted

Application code never manually calculates balances; it reads `wallets.balance` directly.

### SQLite Settings

```
PRAGMA journal_mode = WAL;    -- concurrent reads during writes
PRAGMA foreign_keys = ON;     -- enforce referential integrity
PRAGMA busy_timeout = 5000;   -- wait up to 5 s before returning SQLITE_BUSY
```

Connection pool is capped at `MaxOpenConns(1)` because SQLite allows only one concurrent writer.

## Session Management

The `alexedwards/scs` library manages sessions. The off-the-shelf `scs/sqlite3store` uses the CGo-based `mattn/go-sqlite3` driver, which conflicts with `modernc.org/sqlite` (pure Go). Instead, `internal/session/session.go` implements the three-method `scs.Store` interface directly against the existing `*sql.DB`:

```go
Find(token string) ([]byte, bool, error)
Commit(token string, b []byte, expiry time.Time) error
Delete(token string) error
```

A background goroutine runs every 5 minutes to purge expired session rows.

## Authentication

- Passwords hashed with `bcrypt` (default cost)
- Sessions use a `kasku_session` cookie (HttpOnly, SameSite=Strict)
- `session.RenewToken()` is called on login and registration to prevent session fixation
- `RequireAuth` middleware loads the authenticated user and their family into `context.Context`

### Invite Flow

1. Admin POSTs to `/family/invite` with invitee name + email
2. A `crypto/rand` 32-byte hex token is stored in `users.invite_token`
3. Admin copies the link `{APP_URL}/invite/{token}` and shares it out-of-band
4. Invitee visits the link, sets a password; the token is cleared
5. No email server is required

## Service Layer

Services contain logic that spans multiple stores or requires validation:

| Service | Key Operations |
|---------|---------------|
| `AuthService` | `Register`, `Login`, `GenerateInvite`, `AcceptInvite` |
| `WalletService` | `Create`, `Update` (with validation) |
| `LedgerService` | `AddEntry`, `DeleteEntry` (with family ownership checks) |

Validation errors are returned as `*service.ValidationError` which carries per-field messages. Handlers pattern-match on this type to re-render forms with inline errors.

## Handler Pattern

Every handler struct is constructed with explicit dependencies:

```go
type LedgerHandler struct {
    session *scs.SessionManager
    ledger  *service.LedgerService
    wallets *store.WalletStore
    lstore  *store.LedgerStore
}
```

Handlers use two helpers to decide what to render:

- `handler.IsHTMX(r)` — returns `true` when `HX-Request: true` header is present
- `handler.Render(w, r, status, component)` — writes the templ component

When HTMX is detected, handlers return HTML fragments only. Otherwise they return full pages.

## HTMX OOB Swap After Ledger Insert

`POST /ledger` returns multiple HTML fragments in one response:

1. **Primary target** (`#ledger-content`): updated ledger table rows
2. `HX-Trigger: closeModal` header — closes the add-entry modal
3. No OOB for the summary cards; a full ledger list reload shows updated totals

## Route Table

| Method | Path | Handler | Auth |
|--------|------|---------|------|
| GET | `/` | redirect | — |
| GET/POST | `/register` | auth | — |
| GET/POST | `/login` | auth | — |
| POST | `/logout` | auth | — |
| GET/POST | `/invite/{token}` | family | — |
| GET | `/dashboard` | dashboard | required |
| GET | `/wallets` | wallet | required |
| GET | `/wallets/new` | wallet | required |
| POST | `/wallets` | wallet | required |
| GET | `/wallets/{id}/edit` | wallet | required |
| PUT | `/wallets/{id}` | wallet | required |
| DELETE | `/wallets/{id}` | wallet | required |
| GET | `/ledger` | ledger | required |
| GET | `/ledger/entries` | ledger | required |
| GET | `/ledger/new` | ledger | required |
| POST | `/ledger` | ledger | required |
| DELETE | `/ledger/{id}` | ledger | required |
| GET | `/family` | family | required |
| POST | `/family/invite` | family | admin |
| DELETE | `/family/members/{id}` | family | admin |

## Adding a New Feature

1. Add columns to `internal/db/migrations/001_initial.sql` using `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` (or create `002_*.sql`)
2. Update the matching struct in `internal/model/`
3. Add store methods in `internal/store/`
4. Add service logic in `internal/service/` if needed
5. Create or update handler methods in `internal/handler/`
6. Create or update templ templates in `web/templates/`
7. Wire new routes in `internal/router/router.go`
8. Run `make templ-gen && make build` to verify

## Testing

```bash
make test          # run all tests with race detector
make test-cover    # generate HTML coverage report
```

Unit tests live alongside the code they test (`_test.go` files). The stores can be tested against a real in-memory SQLite database by calling `db.Open(":memory:")`.
