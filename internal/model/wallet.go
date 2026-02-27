package model

import (
	"database/sql"
	"time"
)

// WalletType categorises what a wallet represents.
type WalletType string

const (
	WalletSalary   WalletType = "salary"
	WalletSavings  WalletType = "savings"
	WalletBusiness WalletType = "business"
	WalletGeneral  WalletType = "general"
)

// WalletTypes returns all valid wallet types for use in UI selects.
func WalletTypes() []WalletType {
	return []WalletType{WalletSalary, WalletSavings, WalletBusiness, WalletGeneral}
}

// Wallet belongs to a family and tracks a running balance.
type Wallet struct {
	ID          int64
	FamilyID    int64
	Name        string
	Type        WalletType
	Currency    string
	Balance     float64
	Description sql.NullString
	IsActive    bool
	CreatedAt   time.Time
}

// TypeLabel returns a human-friendly label for the wallet type.
func (w *Wallet) TypeLabel() string {
	switch w.Type {
	case WalletSalary:
		return "Salary"
	case WalletSavings:
		return "Savings"
	case WalletBusiness:
		return "Business"
	default:
		return "General"
	}
}
