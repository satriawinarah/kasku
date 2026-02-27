package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/kasku/kasku/internal/model"
	"github.com/kasku/kasku/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors returned by AuthService methods.
var (
	ErrEmailTaken        = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken      = errors.New("invite token not found or already used")
)

// AuthService handles registration, login, and invite flows.
type AuthService struct {
	families *store.FamilyStore
	users    *store.UserStore
}

func NewAuthService(families *store.FamilyStore, users *store.UserStore) *AuthService {
	return &AuthService{families: families, users: users}
}

// Register creates a new family with an admin user in one logical operation.
func (s *AuthService) Register(ctx context.Context, familyName, userName, email, password string) (*model.User, error) {
	// Check for duplicate email first
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil, ErrEmailTaken
	}

	family, err := s.families.Create(ctx, familyName)
	if err != nil {
		return nil, fmt.Errorf("create family: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &model.User{
		FamilyID:     family.ID,
		Name:         userName,
		Email:        email,
		PasswordHash: string(hash),
		Role:         model.RoleAdmin,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// Login verifies credentials and returns the authenticated user.
func (s *AuthService) Login(ctx context.Context, email, password string) (*model.User, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Constant-time comparison even on missing user to prevent timing attacks
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$placeholder"), []byte(password))
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

// GenerateInvite creates an invited (pending) user record with a one-time token.
// The admin copies the invite link to share with the invitee.
func (s *AuthService) GenerateInvite(ctx context.Context, familyID int64, name, email string) (*model.User, error) {
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil, ErrEmailTaken
	}

	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(raw)

	u := &model.User{
		FamilyID:    familyID,
		Name:        name,
		Email:       email,
		Role:        model.RoleMember,
		InviteToken: sql.NullString{String: token, Valid: true},
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create invited user: %w", err)
	}
	return u, nil
}

// AcceptInvite sets the password for an invited user and clears their invite token.
func (s *AuthService) AcceptInvite(ctx context.Context, token, password string) (*model.User, error) {
	u, err := s.users.GetByInviteToken(ctx, token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u.PasswordHash = string(hash)
	u.InviteToken = sql.NullString{Valid: false}
	if err := s.users.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return u, nil
}
