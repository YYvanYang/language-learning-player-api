// ============================================
// FILE: cmd/api/main.go (REFINED)
// ============================================
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

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/time/rate"

	// --- Project Imports ---
	// Docs (Keep for Swagger generation)
	_ "github.com/yvanyang/language-learning-player-api/docs"

	// Adapters
	httpadapter "github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http"
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/middleware"
	repo "github.com/yvanyang/language-learning-player-api/internal/adapter/repository/postgres"
	googleauthadapter "github.com/yvanyang/language-learning-player-api/internal/adapter/service/google_auth"
	minioadapter "github.com/yvanyang/language-learning-player-api/internal/adapter/service/minio"

	// Core
	"github.com/yvanyang/language-learning-player-api/internal/config"
	uc "github.com/yvanyang/language-learning-player-api/internal/usecase"

	// Packages
	"github.com/yvanyang/language-learning-player-api/pkg/logger"
	"github.com/yvanyang/language-learning-player-api/pkg/security"
	"github.com/yvanyang/language-learning-player-api/pkg/validation"
)

// @title Language Learning Audio Player API
// @version 1.0.0
// @description API specification for the backend of the Language Learning Audio Player application. Provides endpoints for user authentication, audio content management, and user activity tracking.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support Team
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @tag.name Authentication
// @tag.description Operations related to user signup, login, and external authentication (e.g., Google).
// @tag.name Users
// @tag.description Operations related to user profiles and their specific resources.
// @tag.name Audio Tracks
// @tag.description Operations related to individual audio tracks, including retrieval and listing. Duration values in responses are in milliseconds.
// @tag.name Audio Collections
// @tag.description Operations related to managing audio collections (playlists, courses).
// @tag.name User Activity
// @tag.description Operations related to tracking user interactions like playback progress and bookmarks. Timestamp/Progress values in requests/responses are in milliseconds.
// @tag.name Uploads
// @tag.description Operations related to requesting upload URLs and finalizing uploads.
// @tag.name Health
// @tag.description API health checks.

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Configuration ---
	cfg, err := config.LoadConfig(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// --- Logger ---
	appLogger := logger.NewLogger(cfg.Log)
	slog.SetDefault(appLogger)
	appLogger.Info("Configuration loaded", "environment", os.Getenv("APP_ENV"))

	// --- Dependency Initialization ---
	appLogger.Info("Initializing dependencies...")

	// Database Pool
	dbPool, err := repo.NewPgxPool(ctx, cfg.Database, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize database connection pool", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close() // Ensure pool is closed on shutdown
	txManager := repo.NewTxManager(dbPool, appLogger)

	// Repositories
	userRepo := repo.NewUserRepository(dbPool, appLogger)
	trackRepo := repo.NewAudioTrackRepository(dbPool, appLogger)
	collectionRepo := repo.NewAudioCollectionRepository(dbPool, appLogger)
	progressRepo := repo.NewPlaybackProgressRepository(dbPool, appLogger)
	bookmarkRepo := repo.NewBookmarkRepository(dbPool, appLogger)
	refreshTokenRepo := repo.NewRefreshTokenRepository(dbPool, appLogger)

	// Services / Helpers
	secHelper, err := security.NewSecurity(cfg.JWT.SecretKey, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize security helper", "error", err)
		os.Exit(1)
	}
	storageService, err := minioadapter.NewMinioStorageService(cfg.Minio, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize MinIO storage service", "error", err)
		os.Exit(1)
	}
	googleAuthService, err := googleauthadapter.NewGoogleAuthService(cfg.Google.ClientID, appLogger)
	if err != nil {
		// Non-fatal: Log warning if Google Auth isn't critical
		appLogger.Warn("Failed to initialize Google Auth service (Google login disabled?)", "error", err)
	}
	validator := validation.New()

	// Use Cases (Injecting dependencies)
	authUseCase := uc.NewAuthUseCase(cfg.JWT, userRepo, refreshTokenRepo, secHelper, googleAuthService, appLogger)
	audioUseCase := uc.NewAudioContentUseCase(cfg, trackRepo, collectionRepo, storageService, txManager, progressRepo, bookmarkRepo, appLogger)
	activityUseCase := uc.NewUserActivityUseCase(progressRepo, bookmarkRepo, trackRepo, appLogger)
	uploadUseCase := uc.NewUploadUseCase(cfg.Minio, trackRepo, storageService, txManager, appLogger)
	userUseCase := uc.NewUserUseCase(userRepo, appLogger)

	// HTTP Handlers (Injecting use cases)
	authHandler := httpadapter.NewAuthHandler(authUseCase, validator)
	audioHandler := httpadapter.NewAudioHandler(audioUseCase, validator)
	activityHandler := httpadapter.NewUserActivityHandler(activityUseCase, validator)
	uploadHandler := httpadapter.NewUploadHandler(uploadUseCase, validator)
	userHandler := httpadapter.NewUserHandler(userUseCase)

	appLogger.Info("Dependencies initialized successfully")

	// --- HTTP Router Setup ---
	appLogger.Info("Setting up HTTP router...")
	router := chi.NewRouter()

	// --- Global Middleware Setup (Applied to ALL routes in order) ---
	router.Use(middleware.RequestID)                             // 1. Assign request ID
	router.Use(middleware.RequestLogger)                         // 2. Log requests (uses request ID)
	router.Use(middleware.Recoverer)                             // 3. Recover from panics
	router.Use(chimiddleware.RealIP)                             // 4. Determine real client IP
	ipLimiter := middleware.NewIPRateLimiter(rate.Limit(10), 20) // TODO: Make rate limit configurable
	router.Use(middleware.RateLimit(ipLimiter))                  // 5. Apply IP-based rate limiting
	router.Use(chimiddleware.StripSlashes)                       // 6. Remove trailing slashes
	router.Use(chimiddleware.Timeout(60 * time.Second))          // 7. Request timeout

	// --- CORS Middleware ---
	// Apply CORS globally before routing to specific handlers
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Cors.AllowedOrigins,
		AllowedMethods:   cfg.Cors.AllowedMethods,
		AllowedHeaders:   cfg.Cors.AllowedHeaders,
		ExposedHeaders:   []string{"Link", "X-Request-ID"}, // Expose necessary headers
		AllowCredentials: cfg.Cors.AllowCredentials,
		MaxAge:           cfg.Cors.MaxAge, // Cache preflight response
	}))

	// --- Non-API Routes (Health Check, Root Redirect to Docs) ---
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
	})

	// --- Swagger Documentation Route ---
	router.Group(func(r chi.Router) {
		// Apply relaxed security headers specifically for Swagger UI compatibility
		r.Use(middleware.SwaggerSecurityHeaders)
		// Serve Swagger UI files
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	})

	// --- API v1 Routes Group ---
	// All routes under /api/v1 get standard API security headers
	router.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.ApiSecurityHeaders) // Apply strict security headers

		// --- Public API Routes (No Authentication Required) ---
		// No specific middleware for this group, inherits from parent
		r.Group(func(public chi.Router) {
			// Authentication Endpoints (Public actions)
			public.Post("/auth/register", authHandler.Register)
			public.Post("/auth/login", authHandler.Login)
			public.Post("/auth/google/callback", authHandler.GoogleCallback)
			public.Post("/auth/refresh", authHandler.Refresh) // Refresh token endpoint is typically public

			// Public Audio Content Retrieval
			// Uses the `audioHandler.ListTracks` function
			public.Get("/audio/tracks", audioHandler.ListTracks)
			// Uses the `audioHandler.GetTrackDetails` function (handler checks auth internally for extra data)
			public.Get("/audio/tracks/{trackId}", audioHandler.GetTrackDetails)
			// Uses the `audioHandler.GetCollectionDetails` function (handler checks auth internally)
			public.Get("/audio/collections/{collectionId}", audioHandler.GetCollectionDetails)
		})

		// --- Protected API Routes (Authentication Required) ---
		// Apply the authentication middleware to all routes in this group
		r.Group(func(protected chi.Router) {
			protected.Use(middleware.Authenticator(secHelper))

			// --- Logout (Requires knowing *who* is logging out) ---
			protected.Post("/auth/logout", authHandler.Logout)

			// --- User Profile Routes ---
			protected.Route("/users/me", func(me chi.Router) {
				me.Get("/", userHandler.GetMyProfile)
				me.Get("/collections", audioHandler.ListMyCollections) // List user's own collections
				// User Activity (Progress)
				me.Route("/progress", func(progress chi.Router) {
					progress.Get("/", activityHandler.ListProgress)         // List user's progress
					progress.Post("/", activityHandler.RecordProgress)      // Record progress for a track
					progress.Get("/{trackId}", activityHandler.GetProgress) // Get specific track progress
				})
				// User Activity (Bookmarks)
				me.Route("/bookmarks", func(bookmarks chi.Router) {
					bookmarks.Get("/", activityHandler.ListBookmarks)                 // List user's bookmarks
					bookmarks.Post("/", activityHandler.CreateBookmark)               // Create a bookmark
					bookmarks.Delete("/{bookmarkId}", activityHandler.DeleteBookmark) // Delete a specific bookmark
				})
			})

			// --- Audio Collection Management Routes ---
			protected.Route("/audio/collections", func(collections chi.Router) {
				collections.Post("/", audioHandler.CreateCollection) // Create new collection
				// Routes for a specific collection (owned by user)
				collections.Route("/{collectionId}", func(collection chi.Router) {
					// GET /{collectionId} is public (defined above)
					collection.Put("/", audioHandler.UpdateCollectionMetadata)     // Update metadata
					collection.Delete("/", audioHandler.DeleteCollection)          // Delete collection
					collection.Put("/tracks", audioHandler.UpdateCollectionTracks) // Update tracks list
				})
			})

			// --- Upload Request Routes (Need auth to know who is uploading) ---
			protected.Route("/uploads/audio", func(upload chi.Router) {
				upload.Post("/request", uploadHandler.RequestUpload)            // Request URL for single file
				upload.Post("/batch/request", uploadHandler.RequestBatchUpload) // Request URLs for multiple files
			})

			// --- Upload Completion / Track Creation Routes (Need auth for ownership) ---
			// POST to create a new track resource *after* upload
			protected.Post("/audio/tracks", uploadHandler.CompleteUploadAndCreateTrack)
			// POST to create multiple track resources *after* batch upload
			protected.Post("/audio/tracks/batch/complete", uploadHandler.CompleteBatchUploadAndCreateTracks)
		})
	})

	appLogger.Info("HTTP router setup complete")

	// --- HTTP Server ---
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router, // Use the configured Chi router
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(appLogger.Handler(), slog.LevelError), // Use slog for server errors
	}

	// --- Start Server & Graceful Shutdown ---
	serverErrors := make(chan error, 1) // Channel to capture server errors

	go func() {
		appLogger.Info("Starting server", "address", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Block until a signal is received or server exits unexpectedly
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("Server failed to start or encountered an error", "error", err)
			// Optional: Attempt cleanup before forceful exit
			dbPool.Close()
			os.Exit(1)
		}
		// If ErrServerClosed, it means Shutdown was called, proceed normally
	case <-ctx.Done():
		// Shutdown signal received
		stop() // Prevent context cancellation from multiple signals
		appLogger.Info("Shutting down server gracefully, press Ctrl+C again to force")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Context for shutdown
		defer cancel()

		// Attempt graceful server shutdown
		if err := srv.Shutdown(shutdownCtx); err != nil {
			appLogger.Error("Server forced to shutdown", "error", err)
		}

		appLogger.Info("Server shutdown complete.")
	}

	// --- Final Cleanup (after shutdown or error exit) ---
	// This ensures the pool is closed even if the server goroutine errors out
	// before graceful shutdown starts.
	appLogger.Info("Closing database connection pool...")
	dbPool.Close() // Close the database pool
	appLogger.Info("Database connection pool closed.")
	appLogger.Info("Application finished.")
}
