package task

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	gosqlc "github.com/raffy-io/taskifyv2/internal/db"
)

type taskHandler struct {
	Queries *gosqlc.Queries
	Session *scs.SessionManager
}


func (h *taskHandler) createTask(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()

    if err := r.ParseForm(); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

	userID := h.Session.GetInt64(ctx, "user_id")
    if userID == 0 {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	form := &taskForm{
		UserID:   userID,
		Title: r.PostFormValue("title"),
		Completed: false,
	}

	task , err := h.Queries.CreateTask(ctx, gosqlc.CreateTaskParams{
        UserID:    form.UserID,
        Title:     form.Title,
    })
    if err != nil {
        http.Error(w, "Failed to create task", http.StatusInternalServerError)
        return
    }

	slog.Info("Successfully added task!","task",task.Title)
}


func (h *taskHandler) toggleTask(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    userID := h.Session.GetInt64(ctx, "user_id")
    if userID == 0 {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }

    // Securely toggle checking both id and user_id to prevent IDOR attacks
    _, err = h.Queries.ToggleTask(ctx, gosqlc.ToggleTaskParams{
        ID:     id,
        UserID: userID,
    })
    if err != nil {
        http.Error(w, "Failed to toggle task", http.StatusInternalServerError)
        return
    }

    slog.Info("Successfully toggled task!", "task_id", id, "user_id", userID)
    w.WriteHeader(http.StatusOK)
}

func (h *taskHandler) deleteTask(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    userID := h.Session.GetInt64(ctx, "user_id")
    if userID == 0 {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }

    // Securely delete checking both id and user_id
    err = h.Queries.DeleteTask(ctx, gosqlc.DeleteTaskParams{
        ID:     id,
        UserID: userID,
    })
    if err != nil {
        http.Error(w, "Failed to delete task", http.StatusInternalServerError)
        return
    }

    slog.Info("Successfully deleted task!", "task_id", id, "user_id", userID)
    w.WriteHeader(http.StatusOK)
}