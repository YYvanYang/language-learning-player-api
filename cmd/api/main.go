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
	"github.com/go-chi/cors" // Import chi cors

	"your_project/internal/config" // Adjust import path
	"your_project/internal/adapter/handler/http/middleware" // Adjust import path for our custom middleware
	httpadapter "your_project/internal/adapter/handler/http" // Alias for http handler package
	repo "your_project/internal/adapter/repository/postgres"
	// service "your_project/internal/adapter/service" // Alias if needed for google/minio later
	"your_project/internal/usecase"
	"your_project/pkg/logger"      // Adjust import path
	"your_project/pkg/security"
	"your_project/pkg/validation"
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
	
	// Database
	dbPool, err := repo.NewPgxPool(ctx, cfg.Database, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize database connection pool", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close() // Ensure pool is closed on shutdown

	
	// Repositories
	userRepo := repo.NewUserRepository(dbPool, appLogger)
	// TODO: Initialize other repositories (AudioTrack, Collection, etc.)

	// Services / Helpers
	secHelper, err := security.NewSecurity(cfg.JWT.SecretKey, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize security helper", "error", err)
		os.Exit(1)
	}
	// TODO: Initialize ExternalAuthService (Google) - Phase 4
	// TODO: Initialize FileStorageService (MinIO) - Phase 5

	// Validator
	validator := validation.New()

	// Use Cases
	authUseCase := usecase.NewAuthUseCase(cfg.JWT, userRepo, secHelper, nil, appLogger) // Pass nil for extAuth for now
	// TODO: Initialize other use cases

	// HTTP Handlers
	authHandler := httpadapter.NewAuthHandler(authUseCase, validator)
	// TODO: Initialize other handlers

	appLogger.Info("Dependencies initialized successfully")

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
	
	router.Use(chimiddleware.Timeout(60 * time.Second)) // Example: 60s request timeout

	// CORS Middleware Setup from Config
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Cors.AllowedOrigins,
		AllowedMethods:   cfg.Cors.AllowedMethods,
		AllowedHeaders:   cfg.Cors.AllowedHeaders,
		ExposedHeaders:   []string{"Link"}, // Add any other headers you expose
		AllowCredentials: cfg.Cors.AllowCredentials,
		MaxAge:           cfg.Cors.MaxAge,
	}))

	// TODO: Add Auth middleware (for protected routes) - Phase 3

	// --- Routes ---
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		// TODO: Add DB ping check here for more comprehensive health check
		fmt.Fprintln(w, "OK")
	})

	// API v1 Routes
	router.Route("/api/v1", func(r chi.Router) {
		// Public routes (Auth)
		r.Group(func(r chi.Router) {
			// No auth middleware here
			r.Post("/auth/register", authHandler.Register)
			r.Post("/auth/login", authHandler.Login)
			r.Post("/auth/google/callback", authHandler.GoogleCallback) // Keep it here, handler deals with logic
		})

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			// Apply JWT authentication middleware
			r.Use(middleware.Authenticator(secHelper))

			// TODO: Add protected endpoints here
			// Example: Get current user profile
			r.Get("/users/me", func(w http.ResponseWriter, r *http.Request) {
				userID, ok := middleware.GetUserIDFromContext(r.Context())
				if !ok {
					// Should not happen if Authenticator middleware worked correctly
					httputil.RespondError(w, r, domain.ErrUnauthenticated)
					return
				}
				// TODO: Implement handler to fetch user details using userID
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Authenticated User ID: %s", userID)
			})
			// TODO: Mount other protected resource handlers (audio, bookmarks, etc.)
		})
	})

	appLogger.Info("HTTP router setup complete")

	// --- HTTP Server ---
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(appLogger.Handler(), slog.LevelError),
	}

	// --- Start Server & Graceful Shutdown ---
	go func() {
		appLogger.Info("Starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("Server failed to start", "error", err)
			os.Exit(1) // Use os.Exit here as logger might not be available in main goroutine anymore
		}
	}()

	<-ctx.Done()
	stop()
	appLogger.Info("Shutting down server gracefully, press Ctrl+C again to force")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout
	defer cancel()

	// Close database pool
	appLogger.Info("Closing database connection pool...")
	dbPool.Close() // pgxpool Close doesn't take a context
	appLogger.Info("Database connection pool closed.")

	// TODO: Add cleanup for MinIO client etc. when initialized

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
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