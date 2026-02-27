package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kasku/kasku/internal/model"
)

// UserStore handles database operations for the users table.
type UserStore struct{ db *sql.DB }

func NewUserStore(db *sql.DB) *UserStore { return &UserStore{db: db} }

const userColumns = `id, family_id, name, email, password_hash, role, invite_token, created_at`

func scanUser(row interface{ Scan(...any) error }) (*model.User, error) {
	u := &model.User{}
	return u, row.Scan(
		&u.ID, &u.FamilyID, &u.Name, &u.Email,
		&u.PasswordHash, &u.Role, &u.InviteToken, &u.CreatedAt,
	)
}

// Create inserts a new user record.
func (s *UserStore) Create(ctx context.Context, u *model.User) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (family_id, name, email, password_hash, role, invite_token)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		u.FamilyID, u.Name, u.Email, u.PasswordHash, u.Role, u.InviteToken,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	u.ID, _ = res.LastInsertId()
	return nil
}

// GetByEmail fetches a user by email address.
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = ?`, email)
	u, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// GetByID fetches a user by primary key.
func (s *UserStore) GetByID(ctx context.Context, id int64) (*model.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = ?`, id)
	u, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}
	return u, nil
}

// GetByInviteToken fetches a pending (uninvited) user by their one-time invite token.
func (s *UserStore) GetByInviteToken(ctx context.Context, token string) (*model.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE invite_token = ?`, token)
	u, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("get user by invite token: %w", err)
	}
	return u, nil
}

// ListByFamily returns all users belonging to a family.
func (s *UserStore) ListByFamily(ctx context.Context, familyID int64) ([]*model.User, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE family_id = ? ORDER BY created_at`, familyID)
	if err != nil {
		return nil, fmt.Errorf("list users by family: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Update persists changes to name, email, password_hash, role, and invite_token.
func (s *UserStore) Update(ctx context.Context, u *model.User) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET name=?, email=?, password_hash=?, role=?, invite_token=? WHERE id=?`,
		u.Name, u.Email, u.PasswordHash, u.Role, u.InviteToken, u.ID,
	)
	return err
}

// Delete removes a user record permanently.
func (s *UserStore) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}
