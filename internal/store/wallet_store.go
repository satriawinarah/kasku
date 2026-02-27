package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kasku/kasku/internal/model"
)

// WalletStore handles database operations for the wallets table.
type WalletStore struct{ db *sql.DB }

func NewWalletStore(db *sql.DB) *WalletStore { return &WalletStore{db: db} }

const walletColumns = `id, family_id, name, type, currency, balance, description, is_active, created_at`

func scanWallet(row interface{ Scan(...any) error }) (*model.Wallet, error) {
	w := &model.Wallet{}
	return w, row.Scan(
		&w.ID, &w.FamilyID, &w.Name, &w.Type, &w.Currency,
		&w.Balance, &w.Description, &w.IsActive, &w.CreatedAt,
	)
}

// Create inserts a new wallet.
func (s *WalletStore) Create(ctx context.Context, w *model.Wallet) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO wallets (family_id, name, type, currency, description)
		 VALUES (?, ?, ?, ?, ?)`,
		w.FamilyID, w.Name, w.Type, w.Currency, w.Description,
	)
	if err != nil {
		return fmt.Errorf("create wallet: %w", err)
	}
	w.ID, _ = res.LastInsertId()
	return nil
}

// GetByID fetches a wallet by primary key.
func (s *WalletStore) GetByID(ctx context.Context, id int64) (*model.Wallet, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+walletColumns+` FROM wallets WHERE id = ?`, id)
	w, err := scanWallet(row)
	if err != nil {
		return nil, fmt.Errorf("get wallet %d: %w", id, err)
	}
	return w, nil
}

// ListByFamily returns wallets for a family. Pass activeOnly=true to exclude deactivated wallets.
func (s *WalletStore) ListByFamily(ctx context.Context, familyID int64, activeOnly bool) ([]*model.Wallet, error) {
	q := `SELECT ` + walletColumns + ` FROM wallets WHERE family_id = ?`
	args := []any{familyID}
	if activeOnly {
		q += ` AND is_active = 1`
	}
	q += ` ORDER BY created_at`

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list wallets: %w", err)
	}
	defer rows.Close()

	var wallets []*model.Wallet
	for rows.Next() {
		w, err := scanWallet(rows)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}
	return wallets, rows.Err()
}

// Update persists changes to a wallet's name, type, currency, and description.
func (s *WalletStore) Update(ctx context.Context, w *model.Wallet) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE wallets SET name=?, type=?, currency=?, description=? WHERE id=?`,
		w.Name, w.Type, w.Currency, w.Description, w.ID,
	)
	return err
}

// SetActive toggles a wallet's active status (soft delete / restore).
func (s *WalletStore) SetActive(ctx context.Context, id int64, active bool) error {
	v := 0
	if active {
		v = 1
	}
	_, err := s.db.ExecContext(ctx, `UPDATE wallets SET is_active=? WHERE id=?`, v, id)
	return err
}

// TotalBalanceByFamily sums the balances of all active wallets in a family.
func (s *WalletStore) TotalBalanceByFamily(ctx context.Context, familyID int64) (float64, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(balance), 0) FROM wallets WHERE family_id = ? AND is_active = 1`,
		familyID)
	var total float64
	return total, row.Scan(&total)
}
