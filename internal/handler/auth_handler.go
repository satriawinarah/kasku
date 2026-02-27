package handler

import (
	"errors"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/kasku/kasku/internal/service"
	"github.com/kasku/kasku/web/templates/pages"
)

// AuthHandler handles registration, login, logout, and invite flows.
type AuthHandler struct {
	session *scs.SessionManager
	auth    *service.AuthService
}

func NewAuthHandler(session *scs.SessionManager, auth *service.AuthService) *AuthHandler {
	return &AuthHandler{session: session, auth: auth}
}

func (h *AuthHandler) HandleRegisterPage(w http.ResponseWriter, r *http.Request) {
	Render(w, r, http.StatusOK, pages.Register(pages.RegisterData{}))
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	familyName := r.FormValue("family_name")
	userName := r.FormValue("user_name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if len(password) < 8 {
		Render(w, r, http.StatusUnprocessableEntity, pages.Register(pages.RegisterData{
			Error:      "Password must be at least 8 characters",
			FamilyName: familyName,
			UserName:   userName,
			Email:      email,
		}))
		return
	}

	user, err := h.auth.Register(r.Context(), familyName, userName, email, password)
	if err != nil {
		msg := "Registration failed. Please try again."
		if errors.Is(err, service.ErrEmailTaken) {
			msg = "That email is already registered."
		}
		Render(w, r, http.StatusUnprocessableEntity, pages.Register(pages.RegisterData{
			Error:      msg,
			FamilyName: familyName,
			UserName:   userName,
			Email:      email,
		}))
		return
	}

	if err := h.session.RenewToken(r.Context()); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	h.session.Put(r.Context(), "userID", user.ID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	Render(w, r, http.StatusOK, pages.Login(pages.LoginData{}))
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.auth.Login(r.Context(), email, password)
	if err != nil {
		Render(w, r, http.StatusUnprocessableEntity, pages.Login(pages.LoginData{
			Error: "Invalid email or password.",
			Email: email,
		}))
		return
	}

	if err := h.session.RenewToken(r.Context()); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	h.session.Put(r.Context(), "userID", user.ID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	h.session.Destroy(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
