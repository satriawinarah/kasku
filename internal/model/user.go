package model

import (
	"database/sql"
	"time"
)

// UserRole defines the access level of a user within their family.
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleMember UserRole = "member"
)

// User represents a family member who can record transactions.
type User struct {
	ID           int64
	FamilyID     int64
	Name         string
	Email        string
	PasswordHash string
	Role         UserRole
	InviteToken  sql.NullString
	CreatedAt    time.Time
}

// IsAdmin returns true when the user has admin privileges.
func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }
