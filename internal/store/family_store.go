package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kasku/kasku/internal/model"
)

// FamilyStore handles database operations for the families table.
type FamilyStore struct{ db *sql.DB }

func NewFamilyStore(db *sql.DB) *FamilyStore { return &FamilyStore{db: db} }

// Create inserts a new family and returns it with its generated ID.
func (s *FamilyStore) Create(ctx context.Context, name string) (*model.Family, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO families (name) VALUES (?)`, name)
	if err != nil {
		return nil, fmt.Errorf("create family: %w", err)
	}
	id, _ := res.LastInsertId()
	return s.GetByID(ctx, id)
}

// GetByID fetches a family by its primary key.
func (s *FamilyStore) GetByID(ctx context.Context, id int64) (*model.Family, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at FROM families WHERE id = ?`, id)
	f := &model.Family{}
	if err := row.Scan(&f.ID, &f.Name, &f.CreatedAt); err != nil {
		return nil, fmt.Errorf("get family %d: %w", id, err)
	}
	return f, nil
}
