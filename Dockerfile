# Dockerfile

# ---- Build Stage ----
    FROM golang:1.23-alpine AS builder

    # Set working directory inside the container
    WORKDIR /app
    
    # Install build tools if needed (e.g., git for private repos, gcc for cgo if CGO_ENABLED=1)
    # RUN apk add --no-cache git build-base
    
    # Download Go modules to leverage Docker layer caching
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the entire application source code
    COPY . .
    
    # Build the Go application
    # - Static linking (optional but recommended for alpine/distroless)
    # - Linker flags to reduce binary size (optional)
    # - Trimpath to remove absolute file paths
    # - Set output path
    ARG TARGETOS=linux
    ARG TARGETARCH=amd64
    RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
        -ldflags='-w -s' \
        -trimpath \
        -o /app/language-player-api \
        ./cmd/api
    
    # ---- Final Stage ----
    # Use a minimal base image
    # Option 1: Alpine (small, includes shell)
    FROM alpine:3.18
    
    # Option 2: Distroless static (even smaller, no shell, requires fully static binary from build stage)
    # FROM gcr.io/distroless/static-debian11 AS final
    
    # Set working directory
    WORKDIR /app
    
    # Copy necessary files from the builder stage
    # Copy only the compiled binary
    COPY --from=builder /app/language-player-api /app/language-player-api
    # Copy configuration templates or default configs (if needed at runtime)
    COPY config.example.yaml /app/config.example.yaml
    # Copy migration files (needed if running migrations from container, or for reference)
    COPY migrations /app/migrations
    
    # Set permissions (Optional, good practice if not running as root)
    RUN addgroup -S appgroup && adduser -S appuser -G appgroup
    RUN chown -R appuser:appgroup /app
    USER appuser
    
    # Expose the port the application listens on (from config, default 8080)
    EXPOSE 8080
    
    # Define the entry point for the container
    # This command will run when the container starts
    ENTRYPOINT ["/app/language-player-api"]
    
    # Optional: Define a default command (can be overridden)
    # CMD [""]