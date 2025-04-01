# Makefile (Example)

# Replace with your actual DSN or load from .env file
DB_URL = postgresql://user:password@localhost:5432/language_learner_db?sslmode=disable

.PHONY: migrateup migratedown gotool

migrateup: gotool
	@echo "Applying database migrations..."
	@migrate -database "$(DB_URL)" -path migrations up

migratedown: gotool
	@echo "Reverting last database migration..."
	@migrate -database "$(DB_URL)" -path migrations down 1

# Ensures migrate CLI is installed
gotool:
	@command -v migrate >/dev/null 2>&1 || { echo >&2 "migrate CLI not found. Installing..."; \
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	if ! command -v migrate >/dev/null 2>&1; then echo >&2 "migrate installation failed or PATH not configured."; exit 1; fi; }