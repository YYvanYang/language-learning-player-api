// cmd/api/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5" // Import Chi
	chimiddleware "github.com/go-chi/chi/v5/middleware" // Chi's built-in middleware

	"your_project/internal/config" // Adjust import path
	"your_project/internal/adapter/handler/http/middleware" // Adjust import path for our custom middleware
	"your_project/pkg/logger"      // Adjust import path
)

func main() {
	// Create a context that listens for the interrupt signal from the OS.
	// This is used for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Configuration ---
	// Load configuration from file (e.g., ./config.yaml) or environment variables
	cfg, err := config.LoadConfig(".") // "." means look in the current directory
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// --- Logger ---
	appLogger := logger.NewLogger(cfg.Log)
	slog.SetDefault(appLogger) // Set as default logger
	appLogger.Info("Configuration loaded")

	// --- Placeholder for Dependency Initialization ---
	appLogger.Info("Initializing dependencies...")
	// TODO: Initialize Database connection (Phase 3)
	// TODO: Initialize MinIO client (Phase 5)
	// TODO: Initialize Google Auth service (Phase 4)
	// TODO: Initialize Repositories (Phase 3+)
	// TODO: Initialize Services (Phase 3+)
	// TODO: Initialize Usecases (Phase 3+)
	// TODO: Initialize Handlers (Phase 3+)
	appLogger.Info("Dependencies initialized (Placeholder)")

	// --- HTTP Router Setup ---
	appLogger.Info("Setting up HTTP router...")
	router := chi.NewRouter()

	// --- Middleware Setup ---
	// Order matters!
	router.Use(middleware.Recoverer)               // Recover from panics first
	router.Use(middleware.RequestID)               // Add request ID to context and header
	router.Use(middleware.RequestLogger)           // Log requests (uses request ID)
	router.Use(chimiddleware.RealIP)               // Use X-Forwarded-For or X-Real-IP
	router.Use(chimiddleware.StripSlashes)         // Remove trailing slashes
	// TODO: Add CORS middleware (using chi/cors or custom based on config)
	// TODO: Add Timeout middleware (chimiddleware.Timeout)
	// TODO: Add Auth middleware (for protected routes) - Phase 3

	// --- Routes ---
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// TODO: Mount API v1 routes under /api/v1 prefix (Phase 3+)
	// router.Mount("/api/v1", apiV1Routes())

	appLogger.Info("HTTP router setup complete")

	// --- HTTP Server ---
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router, // Use Chi router
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(appLogger.Handler(), slog.LevelError),
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		appLogger.Info("Starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// --- Graceful Shutdown ---
	// Wait for interrupt signal
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	appLogger.Info("Shutting down server gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the requests it is currently handling
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Timeout for graceful shutdown
	defer cancel()

	// Perform cleanup tasks here (e.g., close database connections)
	// TODO: Add cleanup for DB, MinIO client etc. when initialized

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err)
		os.Exit(1) // Exit immediately if shutdown fails
	}

	appLogger.Info("Server exiting")
}

// Placeholder for API routes function (to be implemented later)
// func apiV1Routes() http.Handler {
// 	r := chi.NewRouter()
//  // TODO: Add middleware specific to v1 (e.g., Auth)
//	// r.Use(middleware.Authenticator)
//
//  // TODO: Mount resource-specific routers (auth, audio, user, etc.)
// 	// r.Mount("/auth", authRoutes())
// 	// r.Mount("/audio", audioRoutes())
// 	return r
// }