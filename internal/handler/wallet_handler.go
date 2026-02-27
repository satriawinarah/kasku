package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/kasku/kasku/internal/model"
	"github.com/kasku/kasku/internal/service"
	"github.com/kasku/kasku/internal/store"
	"github.com/kasku/kasku/web/templates/pages"
)

// WalletHandler manages wallet CRUD operations.
type WalletHandler struct {
	session *scs.SessionManager
	wallets *service.WalletService
	store   *store.WalletStore
}

func NewWalletHandler(session *scs.SessionManager, wallets *service.WalletService, store *store.WalletStore) *WalletHandler {
	return &WalletHandler{session: session, wallets: wallets, store: store}
}

func (h *WalletHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	family := GetFamily(r)
	flash := PopFlash(h.session, r)

	walletList, _ := h.store.ListByFamily(r.Context(), family.ID, false)
	Render(w, r, http.StatusOK, pages.Wallets(pages.WalletsData{
		User:    user,
		Flash:   flash,
		Wallets: walletList,
	}))
}

func (h *WalletHandler) HandleNewForm(w http.ResponseWriter, r *http.Request) {
	Render(w, r, http.StatusOK, pages.WalletForm(pages.WalletFormData{}))
}

func (h *WalletHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	family := GetFamily(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	in := service.WalletInput{
		Name:        r.FormValue("name"),
		Type:        model.WalletType(r.FormValue("type")),
		Currency:    r.FormValue("currency"),
		Description: r.FormValue("description"),
	}
	if in.Currency == "" {
		in.Currency = "IDR"
	}

	_, err := h.wallets.Create(r.Context(), family.ID, in)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			Render(w, r, http.StatusUnprocessableEntity, pages.WalletForm(pages.WalletFormData{
				FieldErrors: ve.Fields,
				Error:       ve.Msg,
			}))
			return
		}
		http.Error(w, "Failed to create wallet", http.StatusInternalServerError)
		return
	}

	SetFlashOK(h.session, r, "Wallet created successfully.")
	HTMXRedirect(w, r, "/wallets")
}

func (h *WalletHandler) HandleEditForm(w http.ResponseWriter, r *http.Request) {
	wallet, ok := h.loadWalletForFamily(w, r)
	if !ok {
		return
	}
	Render(w, r, http.StatusOK, pages.WalletForm(pages.WalletFormData{Wallet: wallet}))
}

func (h *WalletHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	wallet, ok := h.loadWalletForFamily(w, r)
	if !ok {
		return
	}
	family := GetFamily(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	in := service.WalletInput{
		Name:        r.FormValue("name"),
		Type:        model.WalletType(r.FormValue("type")),
		Currency:    r.FormValue("currency"),
		Description: r.FormValue("description"),
	}
	if in.Currency == "" {
		in.Currency = "IDR"
	}

	err := h.wallets.Update(r.Context(), wallet.ID, family.ID, in)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			Render(w, r, http.StatusUnprocessableEntity, pages.WalletForm(pages.WalletFormData{
				Wallet:      wallet,
				FieldErrors: ve.Fields,
				Error:       ve.Msg,
			}))
			return
		}
		http.Error(w, "Failed to update wallet", http.StatusInternalServerError)
		return
	}

	SetFlashOK(h.session, r, "Wallet updated.")
	HTMXRedirect(w, r, "/wallets")
}

func (h *WalletHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	wallet, ok := h.loadWalletForFamily(w, r)
	if !ok {
		return
	}
	h.store.SetActive(r.Context(), wallet.ID, false)
	// Return empty string so HTMX outerHTML swap removes the card
	w.WriteHeader(http.StatusOK)
}

func (h *WalletHandler) loadWalletForFamily(w http.ResponseWriter, r *http.Request) (*model.Wallet, bool) {
	family := GetFamily(r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return nil, false
	}
	wallet, err := h.store.GetByID(r.Context(), id)
	if err != nil || wallet.FamilyID != family.ID {
		http.Error(w, "Not found", http.StatusNotFound)
		return nil, false
	}
	return wallet, true
}
