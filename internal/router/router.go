package router

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kasku/kasku/internal/handler"
	"github.com/kasku/kasku/internal/store"
)

// Handlers is a convenience struct grouping all HTTP handler instances.
type Handlers struct {
	Auth      *handler.AuthHandler
	Dashboard *handler.DashboardHandler
	Wallet    *handler.WalletHandler
	Ledger    *handler.LedgerHandler
	Family    *handler.FamilyHandler
}

// New constructs the Chi router and wires all routes.
func New(session *scs.SessionManager, stores *store.Store, h *Handlers) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(session.LoadAndSave)

	// Public routes
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
	})
	r.Get("/register", h.Auth.HandleRegisterPage)
	r.Post("/register", h.Auth.HandleRegister)
	r.Get("/login", h.Auth.HandleLoginPage)
	r.Post("/login", h.Auth.HandleLogin)
	r.Post("/logout", h.Auth.HandleLogout)

	// Invite routes (no auth required — invitee has no account yet)
	r.Get("/invite/{token}", h.Family.HandleInvitePage)
	r.Post("/invite/{token}", h.Family.HandleInviteAccept)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(handler.RequireAuth(session, stores.User, stores.Family))

		r.Get("/dashboard", h.Dashboard.HandleDashboard)

		// Wallets
		r.Get("/wallets", h.Wallet.HandleList)
		r.Get("/wallets/new", h.Wallet.HandleNewForm)
		r.Post("/wallets", h.Wallet.HandleCreate)
		r.Get("/wallets/{id}/edit", h.Wallet.HandleEditForm)
		r.Put("/wallets/{id}", h.Wallet.HandleUpdate)
		r.Delete("/wallets/{id}", h.Wallet.HandleDelete)

		// Ledger
		r.Get("/ledger", h.Ledger.HandleList)
		r.Get("/ledger/entries", h.Ledger.HandlePartial)
		r.Get("/ledger/new", h.Ledger.HandleNewForm)
		r.Post("/ledger", h.Ledger.HandleCreate)
		r.Delete("/ledger/{id}", h.Ledger.HandleDelete)

		// Family (all authenticated users can view; admin-only sub-routes use With)
		r.Get("/family", h.Family.HandlePage)
		r.With(handler.RequireAdmin).Post("/family/invite", h.Family.HandleInviteCreate)
		r.With(handler.RequireAdmin).Delete("/family/members/{id}", h.Family.HandleRemoveMember)
	})

	return r
}
