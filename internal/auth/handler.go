package auth

import (
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/raffy-io/taskifyv2/internal/db"
	"github.com/raffy-io/taskifyv2/internal/utils"
)

type authHandler struct {
	Queries *db.Queries
	Session *scs.SessionManager
}

func (h *authHandler) processRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	form := &authForm{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}

	if !form.validate() {
		slog.Error("Failed to register", "err", form.Errors)
		http.Error(w, "Validation failed", http.StatusBadRequest)
		return
	}

	// Hash the password easily using your internal utility!
	hashedPassword, err := utils.HashPassword(form.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.Queries.CreateUser(ctx, db.CreateUserParams{
		Email:    form.Email,
		PasswordHash: hashedPassword,
	})

	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	slog.Info("User %v registered successfully!", "email", user.Email)
}

func (h *authHandler) processLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	form := &authForm{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}

	user, err := h.Queries.GetUserByEmail(ctx, form.Email)
	if err != nil {
		// User not found or DB error
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	match := utils.CheckPassword(form.Password, user.PasswordHash)
	if !match {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	h.Session.Put(r.Context(), "user_id", user.ID)

	slog.Info("Welcome %v ", "email", user.Email)

}

func (h *authHandler) processLogout(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Destroy the session completely (removes it from DB and clears memory)
    err := h.Session.Destroy(ctx)
    if err != nil {
        http.Error(w, "Failed to logout", http.StatusInternalServerError)
        return
    }

    // 2. Renew the token just to be extra secure against session fixation attacks
    err = h.Session.RenewToken(ctx)
    if err != nil {
        http.Error(w, "Failed to renew session token", http.StatusInternalServerError)
        return
    }

    // 3. Redirect the user back to the login page with a 303 See Other
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}
