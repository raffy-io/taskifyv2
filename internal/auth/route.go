package auth

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	gosqlc "github.com/raffy-io/taskifyv2/internal/db"
)

func AuthRoutes(queries *gosqlc.Queries, sessionManager *scs.SessionManager) http.Handler {
    r := chi.NewRouter()

    h := &authHandler{
        Queries: queries,
        Session: sessionManager,
    }

    r.Post("/register", h.processRegister)
    r.Post("/login", h.processLogin)
    r.Delete("/logout", h.processLogout)

    return r
}