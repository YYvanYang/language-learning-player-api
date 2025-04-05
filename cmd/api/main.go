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
	httpSwagger "github.com/swaggo/http-swagger" // 添加Swagger UI处理器
	"golang.org/x/time/rate"

	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware" // Adjust import path for our custom middleware
	httpadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http" // Alias for http handler package
	minioadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/service/minio" // Alias for minio adapter
	googleauthadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/service/google_auth" // Restored import
	repo "github.com/yvanyang/language-learning-player-backend/internal/adapter/repository/postgres"
	uc "github.com/yvanyang/language-learning-player-backend/internal/usecase" // Alias usecase package if needed elsewhere
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
	// Restore Google Auth Service initialization
	googleAuthService, err := googleauthadapter.NewGoogleAuthService(cfg.Google.ClientID, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize Google Auth service", "error", err)
		// Decide if this is fatal. If Google login is optional, maybe just log a warning?
		// If mandatory or a core feature, exit.
		os.Exit(1) // Assuming it's important
	}

	// Validator
	validator := validation.New()

	// Use Cases
	// Pass googleAuthService to AuthUseCase
	authUseCase := uc.NewAuthUseCase(cfg.JWT, userRepo, secHelper, googleAuthService, appLogger)
	audioUseCase := uc.NewAudioContentUseCase(cfg.Minio, trackRepo, collectionRepo, storageService, appLogger)
	activityUseCase := uc.NewUserActivityUseCase(progressRepo, bookmarkRepo, trackRepo, appLogger)
	uploadUseCase := uc.NewUploadUseCase(cfg.Minio, trackRepo, storageService, appLogger) // New

	// HTTP Handlers
	authHandler := httpadapter.NewAuthHandler(authUseCase, validator)
	audioHandler := httpadapter.NewAudioHandler(audioUseCase, validator)
	activityHandler := httpadapter.NewUserActivityHandler(activityUseCase, validator)
	uploadHandler := httpadapter.NewUploadHandler(uploadUseCase, validator) // New

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

	// --- Routes ---
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
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
			r.Get("/audio/tracks", audioHandler.ListTracks)
			r.Get("/audio/tracks/{trackId}", audioHandler.GetTrackDetails)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticator(secHelper))

			// User Profile (Placeholder)
			r.Get("/users/me", func(w http.ResponseWriter, r *http.Request) { /* ... */ })

			// Collections (Authenticated Actions)
			r.Post("/audio/collections", audioHandler.CreateCollection)
			r.Get("/audio/collections/{collectionId}", audioHandler.GetCollectionDetails)
			r.Put("/audio/collections/{collectionId}", audioHandler.UpdateCollectionMetadata)
			r.Delete("/audio/collections/{collectionId}", audioHandler.DeleteCollection)
			r.Put("/audio/collections/{collectionId}/tracks", audioHandler.UpdateCollectionTracks)

			// User Activity Routes
			r.Post("/users/me/progress", activityHandler.RecordProgress)
			r.Get("/users/me/progress", activityHandler.ListProgress)
			r.Get("/users/me/progress/{trackId}", activityHandler.GetProgress)

			r.Post("/bookmarks", activityHandler.CreateBookmark)
			r.Get("/bookmarks", activityHandler.ListBookmarks)
			r.Delete("/bookmarks/{bookmarkId}", activityHandler.DeleteBookmark)

			// Upload Routes (New)
			r.Post("/uploads/audio/request", uploadHandler.RequestUpload)
			// Reuse POST /audio/tracks for completing the upload and creating the record
			r.Post("/audio/tracks", uploadHandler.CompleteUploadAndCreateTrack)
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

	router.Get("/swagger/*", httpSwagger.WrapHandler)

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
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()
	appLogger.Info("Shutting down server gracefully, press Ctrl+C again to force")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Close database pool
	appLogger.Info("Closing database connection pool...")
	dbPool.Close()
	appLogger.Info("Database connection pool closed.")

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err)
		// Consider os.Exit(1) here as well if shutdown fails critically
	}
	appLogger.Info("Server shutdown complete.")
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