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

	"path/filepath" // Need this

	"github.com/go-chi/chi/v5" // Import Chi
	chimiddleware "github.com/go-chi/chi/v5/middleware" // Chi's built-in middleware
	"github.com/go-chi/cors" // Import chi cors
	"golang.org/x/time/rate"

	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware" // Adjust import path for our custom middleware
	httpadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http" // Alias for http handler package
	minioadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/service/minio" // Alias for minio adapter
	googleauthadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/service/google_auth" // New import
	repo "github.com/yvanyang/language-learning-player-backend/internal/adapter/repository/postgres"
	// service "github.com/yvanyang/language-learning-player-backend/internal/adapter/service" // Alias if needed for google/minio later
	"github.com/yvanyang/language-learning-player-backend/internal/usecase"
	"github.com/yvanyang/language-learning-player-backend/pkg/logger"      // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/security"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"
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
	trackRepo := repo.NewAudioTrackRepository(dbPool, appLogger) // New
	collectionRepo := repo.NewAudioCollectionRepository(dbPool, appLogger) // New
	progressRepo := repo.NewPlaybackProgressRepository(dbPool, appLogger) // New
	bookmarkRepo := repo.NewBookmarkRepository(dbPool, appLogger)         // New

	// Services / Helpers
	secHelper, err := security.NewSecurity(cfg.JWT.SecretKey, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize security helper", "error", err)
		os.Exit(1)
	}
	storageService, err := minioadapter.NewMinioStorageService(cfg.Minio, appLogger) // New
	if err != nil {
		appLogger.Error("Failed to initialize MinIO storage service", "error", err)
		os.Exit(1)
	}
	googleAuthService, err := googleauthadapter.NewGoogleAuthService(cfg.Google.ClientID, appLogger) // New
	if err != nil {
		appLogger.Error("Failed to initialize Google Auth service", "error", err)
		// Decide if this is fatal. If Google login is optional, maybe just log a warning?
		// If mandatory or a core feature, exit.
		os.Exit(1) // Assuming it's important
	}

	// Validator
	validator := validation.New()

	// Use Cases
	authUseCase := usecase.NewAuthUseCase(cfg.JWT, userRepo, secHelper, nil, appLogger) // Pass nil for extAuth for now
	audioUseCase := usecase.NewAudioContentUseCase(cfg.Minio, trackRepo, collectionRepo, storageService, appLogger) // New
	activityUseCase := usecase.NewUserActivityUseCase(progressRepo, bookmarkRepo, trackRepo, appLogger) // New

	// HTTP Handlers
	authHandler := httpadapter.NewAuthHandler(authUseCase, validator)
	audioHandler := httpadapter.NewAudioHandler(audioUseCase, validator) // New
	activityHandler := httpadapter.NewUserActivityHandler(activityUseCase, validator) // New

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

	// Rate Limiter (Example: 10 requests/sec, burst of 20 per IP)
    ipLimiter := middleware.NewIPRateLimiter(rate.Limit(10), 20)
    // Optional: Run cleanup in background (requires better cleanup logic in limiter)
    // go ipLimiter.CleanUpOldLimiters(10*time.Minute, 1*time.Hour)
    router.Use(middleware.RateLimit(ipLimiter)) // Add rate limiting middleware

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
		// Public routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/register", authHandler.Register)
			r.Post("/auth/login", authHandler.Login)
			r.Post("/auth/google/callback", authHandler.GoogleCallback)

			// Public Audio Routes
			r.Get("/audio/tracks", audioHandler.ListTracks)           // List public tracks
			r.Get("/audio/tracks/{trackId}", audioHandler.GetTrackDetails) // Get details (auth check maybe inside usecase/handler for private)
			// Maybe list public collections?
			// r.Get("/audio/collections", audioHandler.ListPublicCollections)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticator(secHelper))

			// User Profile
			r.Get("/users/me", func(w http.ResponseWriter, r *http.Request) { /* ... Placeholder ... */ })
			// TODO: Add PUT /users/me

			// Collections (Authenticated Actions)
			r.Post("/audio/collections", audioHandler.CreateCollection)
			r.Get("/audio/collections/{collectionId}", audioHandler.GetCollectionDetails) // Re-add here for authenticated access/ownership check
			r.Put("/audio/collections/{collectionId}", audioHandler.UpdateCollectionMetadata)
			r.Delete("/audio/collections/{collectionId}", audioHandler.DeleteCollection)
			r.Put("/audio/collections/{collectionId}/tracks", audioHandler.UpdateCollectionTracks)
			// TODO: Add DELETE /audio/collections/{collectionId}/tracks/{trackId} ? (or handle via UpdateCollectionTracks)

			// User Activity Routes (New)
			r.Post("/users/me/progress", activityHandler.RecordProgress)
			r.Get("/users/me/progress", activityHandler.ListProgress)       // Get list of all progress
            r.Get("/users/me/progress/{trackId}", activityHandler.GetProgress) // Get progress for specific track

			r.Post("/bookmarks", activityHandler.CreateBookmark)
			r.Get("/bookmarks", activityHandler.ListBookmarks) // List bookmarks (can filter by trackId query param)
			r.Delete("/bookmarks/{bookmarkId}", activityHandler.DeleteBookmark)

			// TODO: Maybe add protected routes for uploading tracks, managing own tracks?
		})
	})

	// Serve OpenAPI spec and Swagger UI
	docsDir := "./docs" // Path to your docs folder
	specFile := "/openapi.yaml"
	uiPath := "/docs/swagger-ui/"

	// Route for the spec file
	router.Get(specFile, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(docsDir, specFile))
	})

	// Route for Swagger UI static files
	fs := http.FileServer(http.Dir(filepath.Join(docsDir, "swagger-ui")))
	router.Handle(uiPath+"*", http.StripPrefix(uiPath, fs))

	// Optional: Redirect /docs to /docs/swagger-ui/
	router.Get("/docs", func(w http.ResponseWriter, r *http.Request){
	   http.Redirect(w, r, uiPath, http.StatusMovedPermanently)
	})

	router.Get("/swagger/*", httpSwagger.WrapHandler) // Use httpSwagger

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