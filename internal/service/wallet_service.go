package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/kasku/kasku/internal/model"
	"github.com/kasku/kasku/internal/store"
)

var ErrWalletNotFound = errors.New("wallet not found")

// WalletInput carries validated fields for create/update operations.
type WalletInput struct {
	Name        string
	Type        model.WalletType
	Currency    string
	Description string
}

// WalletService contains business logic for wallet management.
type WalletService struct {
	wallets *store.WalletStore
}

func NewWalletService(wallets *store.WalletStore) *WalletService {
	return &WalletService{wallets: wallets}
}

// Create validates input and inserts a new wallet for a family.
func (s *WalletService) Create(ctx context.Context, familyID int64, in WalletInput) (*model.Wallet, error) {
	if err := validateWalletInput(in); err != nil {
		return nil, err
	}

	w := &model.Wallet{
		FamilyID: familyID,
		Name:     in.Name,
		Type:     in.Type,
		Currency: in.Currency,
	}
	if in.Description != "" {
		w.Description.String = in.Description
		w.Description.Valid = true
	}

	if err := s.wallets.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return w, nil
}

// Update validates input and persists changes to a wallet owned by the family.
func (s *WalletService) Update(ctx context.Context, walletID, familyID int64, in WalletInput) error {
	if err := validateWalletInput(in); err != nil {
		return err
	}

	w, err := s.wallets.GetByID(ctx, walletID)
	if err != nil || w.FamilyID != familyID {
		return ErrWalletNotFound
	}

	w.Name = in.Name
	w.Type = in.Type
	w.Currency = in.Currency
	if in.Description != "" {
		w.Description.String = in.Description
		w.Description.Valid = true
	} else {
		w.Description.Valid = false
	}

	return s.wallets.Update(ctx, w)
}

func validateWalletInput(in WalletInput) error {
	if in.Name == "" {
		return &ValidationError{Fields: map[string]string{"name": "Name is required"}}
	}
	validTypes := map[model.WalletType]bool{
		model.WalletSalary: true, model.WalletSavings: true,
		model.WalletBusiness: true, model.WalletGeneral: true,
	}
	if !validTypes[in.Type] {
		return &ValidationError{Fields: map[string]string{"type": "Invalid wallet type"}}
	}
	if in.Currency == "" {
		in.Currency = "IDR"
	}
	return nil
}
