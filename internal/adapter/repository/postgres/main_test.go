// internal/adapter/repository/postgres/main_test.go
package postgres_test // Use _test package

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres" // DB driver for migrate
    _ "github.com/golang-migrate/migrate/v4/source/file"      // File source driver for migrate
    "log/slog"
    "io"
)

var testDBPool *pgxpool.Pool
var testLogger *slog.Logger

// TestMain sets up the test database container and runs migrations.
func TestMain(m *testing.M) {
    testLogger = slog.New(slog.NewTextHandler(io.Discard, nil)) // Discard logs during tests
    pool, err := dockertest.NewPool("") // Connect to Docker daemon
    if err != nil {
        log.Fatalf("Could not construct docker pool: %s", err)
    }

    err = pool.Client.Ping() // Check Docker connection
    if err != nil {
        log.Fatalf("Could not connect to Docker: %s", err)
    }

    // Pull image, run container, and wait for it to be ready
    resource, err := pool.RunWithOptions(&dockertest.RunOptions{
        Repository: "postgres",
        Tag:        "15-alpine", // Use a specific version
        Env: []string{
            "POSTGRES_PASSWORD=secret",
            "POSTGRES_USER=testuser",
            "POSTGRES_DB=testdb",
            "listen_addresses = '*'",
        },
    }, func(config *docker.HostConfig) {
        // Set AutoRemove to true so that stopped container is automatically removed
        config.AutoRemove = true
        config.RestartPolicy = docker.RestartPolicy{Name: "no"}
    })
    if err != nil {
        log.Fatalf("Could not start resource: %s", err)
    }

    // Set expiration timer for the container (optional)
    // resource.Expire(120) // Tell docker to kill the container after 120 seconds

    hostAndPort := resource.GetHostPort("5432/tcp")
    databaseUrl := fmt.Sprintf("postgresql://testuser:secret@%s/testdb?sslmode=disable", hostAndPort)

    log.Println("Connecting to database on url: ", databaseUrl)

    // Wait for the database container to be ready
    if err := pool.Retry(func() error {
        var err error
        // Use context with timeout for connection attempt
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        testDBPool, err = pgxpool.New(ctx, databaseUrl)
        if err != nil {
            return err
        }
        return testDBPool.Ping(ctx)
    }); err != nil {
        log.Fatalf("Could not connect to docker database: %s", err)
    }

    log.Println("Database connection successful.")

    // --- Run Migrations ---
    // Construct the migration source URL relative to the test file's location
    // This assumes your migrations folder is ../../../migrations relative to this file
    // Adjust the path accordingly if your structure is different.
    migrationURL := "file://../../../migrations"
    log.Println("Running migrations from:", migrationURL)

    mig, err := migrate.New(migrationURL, databaseUrl)
    if err != nil {
        log.Fatalf("Could not create migrate instance: %s", err)
    }
    if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
         log.Fatalf("Could not run database migrations: %s", err)
    }
    log.Println("Database migrations applied successfully.")
    // ----------------------

    // Run the tests
    code := m.Run()

    // --- Teardown ---
    // Close pool before purging container
    if testDBPool != nil {
        testDBPool.Close()
    }

    // You can't defer this because os.Exit doesn't care for defer
    if err := pool.Purge(resource); err != nil {
        log.Fatalf("Could not purge resource: %s", err)
    }
    log.Println("Database container purged.")

    os.Exit(code)
}