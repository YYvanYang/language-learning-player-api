// ============================================
// FILE: cmd/api/main.go (MODIFIED)
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

	_ "github.com/yvanyang/language-learning-player-backend/docs"                                    // Keep this - Import generated docs
	httpadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http" // Alias for http handler package
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"  // Adjust import path for our custom middleware
	repo "github.com/yvanyang/language-learning-player-backend/internal/adapter/repository/postgres" // Alias for postgres repo package
	googleauthadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/service/google_auth"
	minioadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/service/minio"
	"github.com/yvanyang/language-learning-player-backend/internal/config"     // Adjust import path
	uc "github.com/yvanyang/language-learning-player-backend/internal/usecase" // Alias usecase package if needed elsewhere
	"github.com/yvanyang/language-learning-player-backend/pkg/logger"          // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/security"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"
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
// @tag.description Operations related to user profiles.
// @tag.name Audio Tracks
// @tag.description Operations related to individual audio tracks, including retrieval and listing.
// @tag.name Audio Collections
// @tag.description Operations related to managing audio collections (playlists, courses).
// @tag.name User Activity
// @tag.description Operations related to tracking user interactions like playback progress and bookmarks.
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
	cfg, err := config.LoadConfig(".") // Pass full config struct
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// --- Logger ---
	appLogger := logger.NewLogger(cfg.Log)
	slog.SetDefault(appLogger)
	appLogger.Info("Configuration loaded")

	// --- Dependency Initialization ---
	appLogger.Info("Initializing dependencies...")

	// Database Pool
	dbPool, err := repo.NewPgxPool(ctx, cfg.Database, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize database connection pool", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close() // Ensure pool is closed on shutdown

	// Initialize Transaction Manager
	txManager := repo.NewTxManager(dbPool, appLogger)

	// Repositories
	userRepo := repo.NewUserRepository(dbPool, appLogger)
	trackRepo := repo.NewAudioTrackRepository(dbPool, appLogger)
	collectionRepo := repo.NewAudioCollectionRepository(dbPool, appLogger)
	progressRepo := repo.NewPlaybackProgressRepository(dbPool, appLogger)
	bookmarkRepo := repo.NewBookmarkRepository(dbPool, appLogger)

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
		// Optional: Log warning instead of exiting if Google Auth isn't critical for basic functionality
		appLogger.Warn("Failed to initialize Google Auth service (Google login disabled?)", "error", err)
		// os.Exit(1) // Only exit if Google Auth is mandatory
	}

	// Validator
	validator := validation.New()

	// Inject dependencies into Use Cases
	authUseCase := uc.NewAuthUseCase(cfg.JWT, userRepo, secHelper, googleAuthService, appLogger)
	// Pass the newly required progressRepo and bookmarkRepo to NewAudioContentUseCase
	audioUseCase := uc.NewAudioContentUseCase(
		cfg,            // config.Config
		trackRepo,      // port.AudioTrackRepository
		collectionRepo, // port.AudioCollectionRepository
		storageService, // port.FileStorageService
		txManager,      // port.TransactionManager
		progressRepo,   // port.PlaybackProgressRepository (ADDED)
		bookmarkRepo,   // port.BookmarkRepository (ADDED)
		appLogger,      // *slog.Logger
	)
	activityUseCase := uc.NewUserActivityUseCase(progressRepo, bookmarkRepo, trackRepo, appLogger)
	// Upload use case constructor might need full cfg if other parts are needed, or just MinioConfig
	uploadUseCase := uc.NewUploadUseCase(cfg.Minio, trackRepo, storageService, appLogger)
	userUseCase := uc.NewUserUseCase(userRepo, appLogger)

	// HTTP Handlers
	authHandler := httpadapter.NewAuthHandler(authUseCase, validator)
	audioHandler := httpadapter.NewAudioHandler(audioUseCase, validator)
	activityHandler := httpadapter.NewUserActivityHandler(activityUseCase, validator)
	uploadHandler := httpadapter.NewUploadHandler(uploadUseCase, validator)
	userHandler := httpadapter.NewUserHandler(userUseCase)

	appLogger.Info("Dependencies initialized successfully")

	// --- HTTP Router Setup ---
	appLogger.Info("Setting up HTTP router...")
	router := chi.NewRouter()

	// --- Middleware Setup (Order matters) ---
	router.Use(middleware.RequestID)                             // 1. Assign Request ID first
	router.Use(middleware.RequestLogger)                         // 2. Log incoming request (with ID)
	router.Use(middleware.Recoverer)                             // 3. Recover from panics (logs with ID)
	router.Use(chimiddleware.RealIP)                             // 4. Get real IP (needed for rate limiting)
	ipLimiter := middleware.NewIPRateLimiter(rate.Limit(10), 20) // Example: 10 req/sec, burst 20
	router.Use(middleware.RateLimit(ipLimiter))                  // 5. Apply rate limiting
	router.Use(cors.Handler(cors.Options{                        // 6. Handle CORS preflight/requests
		AllowedOrigins:   cfg.Cors.AllowedOrigins,
		AllowedMethods:   cfg.Cors.AllowedMethods,
		AllowedHeaders:   cfg.Cors.AllowedHeaders,
		ExposedHeaders:   []string{"Link", "X-Request-ID"}, // Expose Request ID if needed by client
		AllowCredentials: cfg.Cors.AllowCredentials,
		MaxAge:           cfg.Cors.MaxAge,
	}))
	router.Use(middleware.SecurityHeaders) // 7. Add Security Headers
	router.Use(chimiddleware.StripSlashes)
	router.Use(chimiddleware.Timeout(60 * time.Second)) // 8. Apply request timeout

	// --- Routes ---
	// Health Check
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add deeper health checks (e.g., DB ping) if needed
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Swagger Docs
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
	})
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	// API v1 Routes
	router.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/register", authHandler.Register)
			r.Post("/auth/login", authHandler.Login)
			r.Post("/auth/google/callback", authHandler.GoogleCallback)

			// Public Audio Routes
			r.Get("/audio/tracks", audioHandler.ListTracks)
			r.Get("/audio/tracks/{trackId}", audioHandler.GetTrackDetails) // Maybe requires auth if track is private? Handled in usecase/handler for now.
			// Public Collection Routes (If any? e.g., listing public courses)
			// r.Get("/audio/collections", audioHandler.ListPublicCollections) // Example
			r.Get("/audio/collections/{collectionId}", audioHandler.GetCollectionDetails) // Requires auth if collection is private
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			// Apply authentication middleware ONLY to this group
			r.Use(middleware.Authenticator(secHelper))

			// User Profile
			r.Get("/users/me", userHandler.GetMyProfile)
			// r.Put("/users/me", userHandler.UpdateMyProfile) // TODO

			// Authenticated Collection Actions
			r.Post("/audio/collections", audioHandler.CreateCollection)
			// r.Get("/users/me/collections", audioHandler.ListMyCollections) // Endpoint to list *only* user's collections
			r.Put("/audio/collections/{collectionId}", audioHandler.UpdateCollectionMetadata)
			r.Delete("/audio/collections/{collectionId}", audioHandler.DeleteCollection)
			r.Put("/audio/collections/{collectionId}/tracks", audioHandler.UpdateCollectionTracks)
			// Add/Remove single track endpoints?
			// r.Post("/audio/collections/{collectionId}/tracks", audioHandler.AddTrackToCollection)
			// r.Delete("/audio/collections/{collectionId}/tracks/{trackId}", audioHandler.RemoveTrackFromCollection)

			// User Activity Routes
			r.Post("/users/me/progress", activityHandler.RecordProgress)
			r.Get("/users/me/progress", activityHandler.ListProgress)
			r.Get("/users/me/progress/{trackId}", activityHandler.GetProgress)

			r.Post("/bookmarks", activityHandler.CreateBookmark)
			r.Get("/bookmarks", activityHandler.ListBookmarks)
			r.Delete("/bookmarks/{bookmarkId}", activityHandler.DeleteBookmark)

			// Upload Routes (Require Auth)
			r.Post("/uploads/audio/request", uploadHandler.RequestUpload)
			r.Post("/audio/tracks", uploadHandler.CompleteUploadAndCreateTrack) // Create track metadata after upload
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
		ErrorLog:     slog.NewLogLogger(appLogger.Handler(), slog.LevelError), // Route server errors to slog
	}

	// --- Start Server & Graceful Shutdown ---
	go func() {
		appLogger.Info("Starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("Server failed to start", "error", err)
			os.Exit(1) // Exit if server cannot start
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	appLogger.Info("Shutting down server gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 10 seconds to finish
	// the requests it is currently handling
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Close database pool first
	appLogger.Info("Closing database connection pool...")
	dbPool.Close()
	appLogger.Info("Database connection pool closed.")

	// Attempt graceful server shutdown
	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server shutdown complete.")
}
