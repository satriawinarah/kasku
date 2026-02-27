package model

import (
	"database/sql"
	"time"
)

// EntryType distinguishes income from expenses.
type EntryType string

const (
	EntryCredit EntryType = "credit"
	EntryDebit  EntryType = "debit"
)

// Category classifies what a ledger entry is for.
type Category string

const (
	CatFood          Category = "food"
	CatTransport     Category = "transport"
	CatUtilities     Category = "utilities"
	CatEntertainment Category = "entertainment"
	CatHealthcare    Category = "healthcare"
	CatSalary        Category = "salary"
	CatSavings       Category = "savings"
	CatBusiness      Category = "business"
	CatOther         Category = "other"
)

// Categories returns all valid categories for UI selects.
func Categories() []Category {
	return []Category{
		CatFood, CatTransport, CatUtilities, CatEntertainment,
		CatHealthcare, CatSalary, CatSavings, CatBusiness, CatOther,
	}
}

// CategoryLabel returns a human-friendly label for a category.
func CategoryLabel(c Category) string {
	labels := map[Category]string{
		CatFood:          "Food & Dining",
		CatTransport:     "Transport",
		CatUtilities:     "Utilities & Bills",
		CatEntertainment: "Entertainment",
		CatHealthcare:    "Healthcare",
		CatSalary:        "Salary",
		CatSavings:       "Savings Transfer",
		CatBusiness:      "Business",
		CatOther:         "Other",
	}
	if l, ok := labels[c]; ok {
		return l
	}
	return string(c)
}

// LedgerEntry is a single credit or debit transaction on a wallet.
type LedgerEntry struct {
	ID        int64
	WalletID  int64
	UserID    int64
	Type      EntryType
	Amount    float64
	Category  Category
	Note      sql.NullString
	Date      time.Time
	CreatedAt time.Time

	// Joined display fields
	WalletName string
	UserName   string
}

// IsCredit returns true for income entries.
func (e *LedgerEntry) IsCredit() bool { return e.Type == EntryCredit }

// MonthlySummary holds aggregated income/expense totals for one month.
type MonthlySummary struct {
	Month       string  // "2006-01"
	TotalCredit float64 // total income
	TotalDebit  float64 // total expenses
	Net         float64 // TotalCredit - TotalDebit
}
