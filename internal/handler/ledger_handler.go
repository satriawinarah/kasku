package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/kasku/kasku/internal/model"
	"github.com/kasku/kasku/internal/service"
	"github.com/kasku/kasku/internal/store"
	"github.com/kasku/kasku/web/templates/pages"
	"github.com/kasku/kasku/web/templates/partials"
)

// LedgerHandler manages ledger entry operations.
type LedgerHandler struct {
	session *scs.SessionManager
	ledger  *service.LedgerService
	wallets *store.WalletStore
	lstore  *store.LedgerStore
}

func NewLedgerHandler(
	session *scs.SessionManager,
	ledger *service.LedgerService,
	wallets *store.WalletStore,
	lstore *store.LedgerStore,
) *LedgerHandler {
	return &LedgerHandler{
		session: session,
		ledger:  ledger,
		wallets: wallets,
		lstore:  lstore,
	}
}

func (h *LedgerHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	family := GetFamily(r)
	flash := PopFlash(h.session, r)

	month, walletID := h.parseFilters(r)
	walletList, _ := h.wallets.ListByFamily(r.Context(), family.ID, true)
	entries, _ := h.lstore.ListByMonth(r.Context(), family.ID, walletID, month)
	summary, _ := h.lstore.MonthlySummary(r.Context(), family.ID, month)

	Render(w, r, http.StatusOK, pages.Ledger(pages.LedgerData{
		User:             user,
		Flash:            flash,
		Wallets:          walletList,
		SelectedWalletID: walletID,
		CurrentMonth:     month,
		Entries:          entries,
		Summary:          summary,
	}))
}

func (h *LedgerHandler) HandlePartial(w http.ResponseWriter, r *http.Request) {
	family := GetFamily(r)
	month, walletID := h.parseFilters(r)
	entries, _ := h.lstore.ListByMonth(r.Context(), family.ID, walletID, month)
	summary, _ := h.lstore.MonthlySummary(r.Context(), family.ID, month)
	Render(w, r, http.StatusOK, partials.LedgerTable(entries, summary))
}

func (h *LedgerHandler) HandleNewForm(w http.ResponseWriter, r *http.Request) {
	family := GetFamily(r)
	month, walletID := h.parseFilters(r)
	walletList, _ := h.wallets.ListByFamily(r.Context(), family.ID, true)
	Render(w, r, http.StatusOK, pages.LedgerForm(pages.LedgerFormData{
		Wallets:  walletList,
		Month:    month,
		WalletID: walletID,
	}))
}

func (h *LedgerHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	family := GetFamily(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	walletID, _ := strconv.ParseInt(r.FormValue("wallet_id"), 10, 64)
	date, err := time.Parse("2006-01-02", r.FormValue("date"))
	if err != nil {
		date = time.Now()
	}

	month := r.FormValue("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	in := service.LedgerInput{
		WalletID: walletID,
		UserID:   user.ID,
		Type:     model.EntryType(r.FormValue("type")),
		Amount:   amount,
		Category: model.Category(r.FormValue("category")),
		Note:     r.FormValue("note"),
		Date:     date,
	}

	_, err = h.ledger.AddEntry(r.Context(), family.ID, in)
	if err != nil {
		var ve *service.ValidationError
		walletList, _ := h.wallets.ListByFamily(r.Context(), family.ID, true)
		if errors.As(err, &ve) {
			Render(w, r, http.StatusUnprocessableEntity, pages.LedgerForm(pages.LedgerFormData{
				Wallets:     walletList,
				FieldErrors: ve.Fields,
				Error:       ve.Msg,
				Month:       month,
				WalletID:    walletID,
			}))
			return
		}
		Render(w, r, http.StatusInternalServerError, pages.LedgerForm(pages.LedgerFormData{
			Wallets: walletList,
			Error:   "Failed to save entry. Please try again.",
			Month:   month,
		}))
		return
	}

	// Reload fresh data for the primary swap target
	entries, _ := h.lstore.ListByMonth(r.Context(), family.ID, walletID, month)
	summary, _ := h.lstore.MonthlySummary(r.Context(), family.ID, month)

	// Close the modal and update the ledger table
	w.Header().Set("HX-Trigger", "closeModal")
	Render(w, r, http.StatusOK, partials.LedgerTable(entries, summary))
}

func (h *LedgerHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	family := GetFamily(r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := h.ledger.DeleteEntry(r.Context(), id, family.ID); err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	// Empty response; HTMX removes the row via outerHTML swap
	w.WriteHeader(http.StatusOK)
}

func (h *LedgerHandler) parseFilters(r *http.Request) (month string, walletID int64) {
	month = r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	walletID, _ = strconv.ParseInt(r.URL.Query().Get("wallet"), 10, 64)
	return
}
