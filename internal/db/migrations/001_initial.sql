PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- Sessions table (used by custom SCS SQLite store)
CREATE TABLE IF NOT EXISTS sessions (
    token  TEXT PRIMARY KEY,
    data   BLOB NOT NULL,
    expiry REAL NOT NULL
);
CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions (expiry);

-- Families: the top-level grouping for husband + wife
CREATE TABLE IF NOT EXISTS families (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

-- Users: each member of a family
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    family_id     INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
    name          TEXT    NOT NULL,
    email         TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL DEFAULT '',
    role          TEXT    NOT NULL CHECK(role IN ('admin','member')) DEFAULT 'member',
    invite_token  TEXT,
    created_at    DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS users_email_idx  ON users (email);
CREATE INDEX IF NOT EXISTS users_family_idx ON users (family_id);
CREATE INDEX IF NOT EXISTS users_invite_idx ON users (invite_token);

-- Wallets: salary, savings, business profit, etc.
CREATE TABLE IF NOT EXISTS wallets (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    family_id   INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    type        TEXT    NOT NULL CHECK(type IN ('salary','savings','business','general')),
    currency    TEXT    NOT NULL DEFAULT 'IDR',
    balance     REAL    NOT NULL DEFAULT 0,
    description TEXT,
    is_active   INTEGER NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS wallets_family_idx ON wallets (family_id);

-- Ledger: every credit (income) and debit (expense) entry
CREATE TABLE IF NOT EXISTS ledger (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    wallet_id  INTEGER NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id),
    type       TEXT    NOT NULL CHECK(type IN ('credit','debit')),
    amount     REAL    NOT NULL CHECK(amount > 0),
    category   TEXT    NOT NULL CHECK(category IN (
                   'food','transport','utilities','entertainment',
                   'healthcare','salary','savings','business','other'
               )),
    note       TEXT,
    date       DATE    NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS ledger_wallet_idx      ON ledger (wallet_id);
CREATE INDEX IF NOT EXISTS ledger_date_idx        ON ledger (date);
CREATE INDEX IF NOT EXISTS ledger_wallet_date_idx ON ledger (wallet_id, date);

-- Trigger: increment wallet balance on credit, decrement on debit
CREATE TRIGGER IF NOT EXISTS ledger_after_insert
AFTER INSERT ON ledger
BEGIN
    UPDATE wallets
    SET balance = balance + CASE
        WHEN NEW.type = 'credit' THEN  NEW.amount
        WHEN NEW.type = 'debit'  THEN -NEW.amount
    END
    WHERE id = NEW.wallet_id;
END;

-- Trigger: reverse the balance effect when a ledger entry is deleted
CREATE TRIGGER IF NOT EXISTS ledger_after_delete
AFTER DELETE ON ledger
BEGIN
    UPDATE wallets
    SET balance = balance - CASE
        WHEN OLD.type = 'credit' THEN  OLD.amount
        WHEN OLD.type = 'debit'  THEN -OLD.amount
    END
    WHERE id = OLD.wallet_id;
END;
