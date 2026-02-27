package handler

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/kasku/kasku/internal/store"
	"github.com/kasku/kasku/web/templates/pages"
)

// DashboardHandler renders the main overview page.
type DashboardHandler struct {
	session *scs.SessionManager
	wallets *store.WalletStore
	ledger  *store.LedgerStore
}

func NewDashboardHandler(session *scs.SessionManager, wallets *store.WalletStore, ledger *store.LedgerStore) *DashboardHandler {
	return &DashboardHandler{session: session, wallets: wallets, ledger: ledger}
}

func (h *DashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	family := GetFamily(r)
	flash := PopFlash(h.session, r)

	month := time.Now().Format("2006-01")

	walletList, _ := h.wallets.ListByFamily(r.Context(), family.ID, true)
	totalBalance, _ := h.wallets.TotalBalanceByFamily(r.Context(), family.ID)
	summary, _ := h.ledger.MonthlySummary(r.Context(), family.ID, month)
	recent, _ := h.ledger.RecentEntries(r.Context(), family.ID, 10)

	Render(w, r, http.StatusOK, pages.Dashboard(pages.DashboardData{
		User:          user,
		Family:        family,
		Flash:         flash,
		TotalBalance:  totalBalance,
		Currency:      "IDR",
		CurrentMonth:  month,
		Summary:       summary,
		Wallets:       walletList,
		RecentEntries: recent,
	}))
}
