package main

import (
	"context"
	"database/sql"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/raffy-io/taskifyv2"
	"github.com/raffy-io/taskifyv2/internal/auth"
	gosqlc "github.com/raffy-io/taskifyv2/internal/db" // Aliased to avoid package name conflict
	"github.com/raffy-io/taskifyv2/internal/middleware"
	"github.com/raffy-io/taskifyv2/internal/task"
)

var sessionManager *scs.SessionManager

func main() {
    ctx := context.Background()
    //--- Load ENV ---//
    _ = godotenv.Load()

    dbURL := os.Getenv("DB_URL")
    if dbURL == "" {
        log.Fatalf("DB_URL environment variable is required")
    }

    //--- Database Connection for Migrations & Sessions ---//
    // Goose and SCS both use standard database/sql connections (*sql.DB)
    sessionDB, err := sql.Open("pgx", dbURL)
    if err != nil {
        log.Fatalf("failed to open database connection: %v", err)
    }
    defer sessionDB.Close()

    // Run auto-migrations on startup
    if err := taskifyv2.RunMigrations(sessionDB); err != nil {
        log.Fatalf("migration failed: %v", err)
    }
    log.Println("Migrations applied successfully!")

    //--- Create Connection Pool for SQLC ---//
    pool, err := pgxpool.New(ctx, dbURL)
    if err != nil {
        slog.Error("Failed to create connection pool", "error", err)
        os.Exit(1)
    }
    defer pool.Close()

    //--- Queries per domain ---//
    queries := gosqlc.New(pool)

    //--- Initialize SCS Session Manager ---//
    sessionManager = scs.New()
    sessionManager.Store = postgresstore.New(sessionDB)
    sessionManager.Lifetime = 24 * time.Hour
    sessionManager.Cookie.HttpOnly = true
    sessionManager.Cookie.SameSite = http.SameSiteStrictMode
    // sessionManager.Cookie.Secure = true // Enable this when using HTTPS in production

    //--- Init CHI ---//
    r := chi.NewRouter()

    //--- Static assets CSS JS IMG ---//
    staticFS, err := fs.Sub(taskifyv2.EmbeddedAssets, "static")
    if err != nil {
        slog.Error("Cant find static assets", "error", err)
        os.Exit(1)
    }
    r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

    // --- Public Routes (Sessions loaded, NO auth required) ---
    r.Group(func(r chi.Router) {
        r.Use(sessionManager.LoadAndSave)
        // r.Use(middleware.NoSurf)

        r.Mount("/auth", auth.AuthRoutes(queries, sessionManager))
    })

    // --- Protected Routes (Require Authentication) ---
    r.Group(func(r chi.Router) {
        r.Use(sessionManager.LoadAndSave)
        // r.Use(middleware.NoSurf)
        r.Use(middleware.RequireAuth(sessionManager))

        r.Mount("/task", task.TasksRoute(queries, sessionManager))
    })

    //--- Server ---//
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    srv := &http.Server{
        Addr:              ":" + port,
        Handler:           r,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       15 * time.Second,
        WriteTimeout:      60 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    slog.Info("Server is running", "addr", srv.Addr)
    if err := srv.ListenAndServe(); err != nil {
        slog.Error("Server failed to start", "error", err)
        os.Exit(1)
    }
}