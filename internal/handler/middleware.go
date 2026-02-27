package handler

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/kasku/kasku/internal/store"
)

// RequireAuth validates the session and loads the user+family into the request context.
// Unauthenticated requests are redirected to /login.
func RequireAuth(session *scs.SessionManager, users *store.UserStore, families *store.FamilyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := session.GetInt64(r.Context(), "userID")
			if userID == 0 {
				if IsHTMX(r) {
					w.Header().Set("HX-Redirect", "/login")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			user, err := users.GetByID(r.Context(), userID)
			if err != nil {
				session.Destroy(r.Context())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			family, err := families.GetByID(r.Context(), user.FamilyID)
			if err != nil {
				http.Error(w, "Family not found", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUser, user)
			ctx = context.WithValue(ctx, ContextKeyFamily, family)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin returns 403 if the authenticated user is not an admin.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user == nil || !user.IsAdmin() {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
