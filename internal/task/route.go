package task

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/raffy-io/taskifyv2/internal/db"
)

func TasksRoute(queries *db.Queries, sessionManager *scs.SessionManager) http.Handler {
    r := chi.NewRouter()

    h := &taskHandler{
        Queries: queries,
        Session: sessionManager,
    }

    // Since this sub-router is mounted at /task, 
    // these paths are relative to /task
    // r.Get("/", h.listTasks)       // GET /task/
    r.Post("/", h.createTask)     // POST /task/
    r.Patch("/{id}/toggle", h.toggleTask) // PATCH /task/{id}/toggle
    r.Delete("/{id}", h.deleteTask)       // DELETE /task/{id}

    return r
}