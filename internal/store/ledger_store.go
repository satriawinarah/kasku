package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kasku/kasku/internal/model"
)

// LedgerStore handles database operations for the ledger table.
type LedgerStore struct{ db *sql.DB }

func NewLedgerStore(db *sql.DB) *LedgerStore { return &LedgerStore{db: db} }

func scanEntry(row interface{ Scan(...any) error }) (*model.LedgerEntry, error) {
	e := &model.LedgerEntry{}
	return e, row.Scan(
		&e.ID, &e.WalletID, &e.UserID, &e.Type, &e.Amount,
		&e.Category, &e.Note, &e.Date, &e.CreatedAt,
		&e.WalletName, &e.UserName,
	)
}

// Create inserts a new ledger entry. The wallet balance is updated by a DB trigger.
func (s *LedgerStore) Create(ctx context.Context, e *model.LedgerEntry) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO ledger (wallet_id, user_id, type, amount, category, note, date)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.WalletID, e.UserID, e.Type, e.Amount, e.Category, e.Note, e.Date.Format("2006-01-02"),
	)
	if err != nil {
		return fmt.Errorf("create ledger entry: %w", err)
	}
	e.ID, _ = res.LastInsertId()
	return nil
}

// GetByID fetches a single ledger entry with joined wallet and user names.
func (s *LedgerStore) GetByID(ctx context.Context, id int64) (*model.LedgerEntry, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT l.id, l.wallet_id, l.user_id, l.type, l.amount,
		       l.category, l.note, l.date, l.created_at,
		       w.name, u.name
		FROM ledger l
		JOIN wallets w ON w.id = l.wallet_id
		JOIN users   u ON u.id = l.user_id
		WHERE l.id = ?`, id)
	e, err := scanEntry(row)
	if err != nil {
		return nil, fmt.Errorf("get entry %d: %w", id, err)
	}
	return e, nil
}

// ListByMonth returns ledger entries for a family in a given month (format "2006-01").
// Pass walletID=0 to return entries from all wallets in the family.
func (s *LedgerStore) ListByMonth(ctx context.Context, familyID, walletID int64, month string) ([]*model.LedgerEntry, error) {
	args := []any{familyID, month + "-01", month + "-31"}
	walletFilter := ""
	if walletID > 0 {
		walletFilter = " AND l.wallet_id = ?"
		args = append(args, walletID)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT l.id, l.wallet_id, l.user_id, l.type, l.amount,
		       l.category, l.note, l.date, l.created_at,
		       w.name, u.name
		FROM ledger l
		JOIN wallets w ON w.id = l.wallet_id
		JOIN users   u ON u.id = l.user_id
		WHERE w.family_id = ?
		  AND l.date BETWEEN ? AND ?`+walletFilter+`
		ORDER BY l.date DESC, l.created_at DESC`, args...)
	if err != nil {
		return nil, fmt.Errorf("list ledger by month: %w", err)
	}
	defer rows.Close()

	var entries []*model.LedgerEntry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// MonthlySummary returns aggregated credit/debit totals for a family in a given month.
func (s *LedgerStore) MonthlySummary(ctx context.Context, familyID int64, month string) (*model.MonthlySummary, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT
		    COALESCE(SUM(CASE WHEN l.type='credit' THEN l.amount ELSE 0 END), 0),
		    COALESCE(SUM(CASE WHEN l.type='debit'  THEN l.amount ELSE 0 END), 0)
		FROM ledger l
		JOIN wallets w ON w.id = l.wallet_id
		WHERE w.family_id = ?
		  AND l.date BETWEEN ? AND ?`,
		familyID, month+"-01", month+"-31")

	s2 := &model.MonthlySummary{Month: month}
	if err := row.Scan(&s2.TotalCredit, &s2.TotalDebit); err != nil {
		return nil, fmt.Errorf("monthly summary: %w", err)
	}
	s2.Net = s2.TotalCredit - s2.TotalDebit
	return s2, nil
}

// RecentEntries returns the last N ledger entries across all wallets in a family.
func (s *LedgerStore) RecentEntries(ctx context.Context, familyID int64, limit int) ([]*model.LedgerEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT l.id, l.wallet_id, l.user_id, l.type, l.amount,
		       l.category, l.note, l.date, l.created_at,
		       w.name, u.name
		FROM ledger l
		JOIN wallets w ON w.id = l.wallet_id
		JOIN users   u ON u.id = l.user_id
		WHERE w.family_id = ?
		ORDER BY l.date DESC, l.created_at DESC
		LIMIT ?`, familyID, limit)
	if err != nil {
		return nil, fmt.Errorf("recent entries: %w", err)
	}
	defer rows.Close()

	var entries []*model.LedgerEntry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// Delete removes a ledger entry. The wallet balance is corrected by a DB trigger.
func (s *LedgerStore) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM ledger WHERE id = ?`, id)
	return err
}

// CategoryTotals returns debit totals grouped by category for a given month.
func (s *LedgerStore) CategoryTotals(ctx context.Context, familyID int64, month string) (map[model.Category]float64, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT l.category, COALESCE(SUM(l.amount), 0)
		FROM ledger l
		JOIN wallets w ON w.id = l.wallet_id
		WHERE w.family_id = ?
		  AND l.type = 'debit'
		  AND l.date BETWEEN ? AND ?
		GROUP BY l.category`,
		familyID, month+"-01", month+"-31")
	if err != nil {
		return nil, fmt.Errorf("category totals: %w", err)
	}
	defer rows.Close()

	totals := make(map[model.Category]float64)
	for rows.Next() {
		var cat model.Category
		var total float64
		if err := rows.Scan(&cat, &total); err != nil {
			return nil, err
		}
		totals[cat] = total
	}
	return totals, rows.Err()
}
