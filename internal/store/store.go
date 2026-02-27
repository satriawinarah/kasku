package store

import "database/sql"

// Store aggregates all data-access repositories.
type Store struct {
	Family *FamilyStore
	User   *UserStore
	Wallet *WalletStore
	Ledger *LedgerStore
}

// New constructs a Store wiring all sub-stores to the same database connection.
func New(db *sql.DB) *Store {
	return &Store{
		Family: NewFamilyStore(db),
		User:   NewUserStore(db),
		Wallet: NewWalletStore(db),
		Ledger: NewLedgerStore(db),
	}
}
