package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/kasku/kasku/internal/service"
	"github.com/kasku/kasku/internal/store"
	"github.com/kasku/kasku/web/templates/pages"
)

// FamilyHandler manages family members and invite flows.
type FamilyHandler struct {
	session *scs.SessionManager
	auth    *service.AuthService
	users   *store.UserStore
	appURL  string
}

func NewFamilyHandler(session *scs.SessionManager, auth *service.AuthService, users *store.UserStore, appURL string) *FamilyHandler {
	return &FamilyHandler{session: session, auth: auth, users: users, appURL: appURL}
}

func (h *FamilyHandler) HandlePage(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	family := GetFamily(r)
	flash := PopFlash(h.session, r)

	members, _ := h.users.ListByFamily(r.Context(), family.ID)
	Render(w, r, http.StatusOK, pages.Family(pages.FamilyData{
		User:    user,
		Family:  family,
		Flash:   flash,
		Members: members,
	}))
}

func (h *FamilyHandler) HandleInviteCreate(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	family := GetFamily(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")

	invited, err := h.auth.GenerateInvite(r.Context(), family.ID, name, email)
	if err != nil {
		members, _ := h.users.ListByFamily(r.Context(), family.ID)
		errMsg := "Failed to generate invite."
		if errors.Is(err, service.ErrEmailTaken) {
			errMsg = "That email is already registered."
		}
		Render(w, r, http.StatusUnprocessableEntity, pages.Family(pages.FamilyData{
			User:        user,
			Family:      family,
			Members:     members,
			InviteError: errMsg,
		}))
		return
	}

	inviteLink := h.appURL + "/invite/" + invited.InviteToken.String
	members, _ := h.users.ListByFamily(r.Context(), family.ID)
	Render(w, r, http.StatusOK, pages.Family(pages.FamilyData{
		User:       user,
		Family:     family,
		Members:    members,
		InviteLink: inviteLink,
	}))
}

func (h *FamilyHandler) HandleInvitePage(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	invitedUser, err := h.users.GetByInviteToken(r.Context(), token)
	if err != nil {
		http.Error(w, "Invite link not found or already used.", http.StatusNotFound)
		return
	}
	Render(w, r, http.StatusOK, pages.Invite(pages.InviteData{
		User:  invitedUser,
		Token: token,
	}))
}

func (h *FamilyHandler) HandleInviteAccept(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	confirm := r.FormValue("password_confirm")

	if len(password) < 8 {
		invitedUser, _ := h.users.GetByInviteToken(r.Context(), token)
		Render(w, r, http.StatusUnprocessableEntity, pages.Invite(pages.InviteData{
			User:  invitedUser,
			Token: token,
			Error: "Password must be at least 8 characters.",
		}))
		return
	}

	if password != confirm {
		invitedUser, _ := h.users.GetByInviteToken(r.Context(), token)
		Render(w, r, http.StatusUnprocessableEntity, pages.Invite(pages.InviteData{
			User:  invitedUser,
			Token: token,
			Error: "Passwords do not match.",
		}))
		return
	}

	u, err := h.auth.AcceptInvite(r.Context(), token, password)
	if err != nil {
		http.Error(w, "Invite expired or invalid.", http.StatusBadRequest)
		return
	}

	if err := h.session.RenewToken(r.Context()); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	h.session.Put(r.Context(), "userID", u.ID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *FamilyHandler) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)
	family := GetFamily(r)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if id == currentUser.ID {
		http.Error(w, "Cannot remove yourself", http.StatusBadRequest)
		return
	}

	target, err := h.users.GetByID(r.Context(), id)
	if err != nil || target.FamilyID != family.ID {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	h.users.Delete(r.Context(), id)
	// Return empty response; HTMX removes the list item via outerHTML swap
	w.WriteHeader(http.StatusOK)
}
