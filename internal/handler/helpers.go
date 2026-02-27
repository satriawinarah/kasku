package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/alexedwards/scs/v2"
	"github.com/kasku/kasku/internal/model"
	"github.com/kasku/kasku/web/templates/components"
)

type contextKey string

const (
	ContextKeyUser   contextKey = "user"
	ContextKeyFamily contextKey = "family"
)

// GetUser retrieves the authenticated user from the request context.
func GetUser(r *http.Request) *model.User {
	u, _ := r.Context().Value(ContextKeyUser).(*model.User)
	return u
}

// GetFamily retrieves the family from the request context.
func GetFamily(r *http.Request) *model.Family {
	f, _ := r.Context().Value(ContextKeyFamily).(*model.Family)
	return f
}

// IsHTMX returns true when the request carries the HX-Request header.
func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// Render writes a templ component to the response.
func Render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	return c.Render(r.Context(), w)
}

// PopFlash reads and clears a flash message from the session.
func PopFlash(session *scs.SessionManager, r *http.Request) components.FlashData {
	if msg := session.PopString(r.Context(), "flash_ok"); msg != "" {
		return components.FlashData{Type: "success", Message: msg}
	}
	if msg := session.PopString(r.Context(), "flash_err"); msg != "" {
		return components.FlashData{Type: "error", Message: msg}
	}
	return components.FlashData{}
}

// SetFlashOK stores a success flash message in the session.
func SetFlashOK(session *scs.SessionManager, r *http.Request, msg string) {
	session.Put(r.Context(), "flash_ok", msg)
}

// SetFlashErr stores an error flash message in the session.
func SetFlashErr(session *scs.SessionManager, r *http.Request, msg string) {
	session.Put(r.Context(), "flash_err", msg)
}

// HTMXRedirect sends HX-Redirect for HTMX or falls back to a 303 redirect.
func HTMXRedirect(w http.ResponseWriter, r *http.Request, url string) {
	if IsHTMX(r) {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}
