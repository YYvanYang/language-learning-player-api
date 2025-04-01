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

	"your_project/internal/config" // Adjust import path
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
	appLogger.Info("Configuration loaded")
	// Set as default logger for slog (optional, but convenient)
	slog.SetDefault(appLogger)

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

	// --- Placeholder for HTTP Router Setup ---
	appLogger.Info("Setting up HTTP router...")
	// TODO: Initialize Chi router (Phase 1.10)
	// TODO: Setup middleware (Phase 1.10 / Phase 3)
	// TODO: Register routes and handlers (Phase 1.10 / Phase 3+)
	// router := chi.NewRouter() // Example placeholder
	router := http.NewServeMux() // Temporary basic mux
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	}) // Basic health check
	appLogger.Info("HTTP router setup complete (Placeholder)")

	// --- HTTP Server ---
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router, // Use the actual router once initialized
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(appLogger.Handler(), slog.LevelError), // Redirect standard error log
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