package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kasku/kasku/internal/model"
	"github.com/kasku/kasku/internal/store"
)

var ErrEntryNotFound = errors.New("ledger entry not found")

// LedgerInput carries validated fields for creating a ledger entry.
type LedgerInput struct {
	WalletID int64
	UserID   int64
	Type     model.EntryType
	Amount   float64
	Category model.Category
	Note     string
	Date     time.Time
}

// LedgerService handles business logic around ledger entries.
type LedgerService struct {
	ledger  *store.LedgerStore
	wallets *store.WalletStore
}

func NewLedgerService(ledger *store.LedgerStore, wallets *store.WalletStore) *LedgerService {
	return &LedgerService{ledger: ledger, wallets: wallets}
}

// AddEntry validates and inserts a new ledger entry.
func (s *LedgerService) AddEntry(ctx context.Context, familyID int64, in LedgerInput) (*model.LedgerEntry, error) {
	if err := validateLedgerInput(in); err != nil {
		return nil, err
	}

	// Ensure the wallet belongs to the user's family
	wallet, err := s.wallets.GetByID(ctx, in.WalletID)
	if err != nil || wallet.FamilyID != familyID {
		return nil, ErrWalletNotFound
	}

	e := &model.LedgerEntry{
		WalletID: in.WalletID,
		UserID:   in.UserID,
		Type:     in.Type,
		Amount:   in.Amount,
		Category: in.Category,
		Date:     in.Date,
	}
	if in.Note != "" {
		e.Note.String = in.Note
		e.Note.Valid = true
	}

	if err := s.ledger.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("add entry: %w", err)
	}
	return e, nil
}

// DeleteEntry removes a ledger entry, verifying it belongs to the user's family.
func (s *LedgerService) DeleteEntry(ctx context.Context, entryID, familyID int64) error {
	entry, err := s.ledger.GetByID(ctx, entryID)
	if err != nil {
		return ErrEntryNotFound
	}
	wallet, err := s.wallets.GetByID(ctx, entry.WalletID)
	if err != nil || wallet.FamilyID != familyID {
		return ErrEntryNotFound
	}
	return s.ledger.Delete(ctx, entryID)
}

func validateLedgerInput(in LedgerInput) error {
	errs := map[string]string{}
	if in.Amount <= 0 {
		errs["amount"] = "Amount must be greater than zero"
	}
	if in.Type != model.EntryCredit && in.Type != model.EntryDebit {
		errs["type"] = "Type must be credit or debit"
	}
	if in.Date.IsZero() {
		errs["date"] = "Date is required"
	}
	if len(errs) > 0 {
		return &ValidationError{Fields: errs, Msg: "Please fix the form errors"}
	}
	return nil
}
