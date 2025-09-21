# Technology Stack

## Core Technologies
- **Language**: Go 1.24+ with toolchain go1.24.5
- **Database**: PostgreSQL with pgx/v5 driver
- **Web Framework**: Chi router (go-chi/chi/v5)
- **Frontend**: Server-side rendered HTML with HTMX 2.0.6, Bootstrap 5.3.2, Bootstrap Icons 1.10.5
- **Authentication**: JWT with go-chi/jwtauth/v5, CSRF protection with gorilla/csrf
- **Templating**: gomponents for type-safe HTML generation
- **Migrations**: golang-migrate/migrate for database schema management

## Key Dependencies
- **Validation**: go-playground/validator/v10
- **Security**: unrolled/secure middleware, golang.org/x/crypto for bcrypt
- **OAuth**: dghubble/gologin/v2, golang.org/x/oauth2
- **Testing**: testify/assert, matryer/is, ory/dockertest/v3 for integration tests
- **Utilities**: google/uuid, kelseyhightower/envconfig, pkg/errors

## Build System & Commands

### Development
```bash
# Run application (uses .env file for config)
go run .

# Start PostgreSQL via Docker
make docker.postgres
```

### Testing
```bash
# Run all tests with coverage
make test

# Run tests manually
go test -v -timeout 60s -coverprofile=cover.out -cover ./...
go tool cover -func=cover.out
```

### Linting
```bash
# Run linter before commits
make linter
```

### Database Management
```bash
# Run migrations up
make migrate.up

# Run migrations down  
make migrate.down

# Drop all tables
make migrate.drop

# Force migration version
make migrate.force version=<version>
```

### Build & Release
```bash
# Clean build directory
make clean

# Build binary
make build

# Test release build
make release.test
```

### Code Quality
```bash
# Run linter
make linter

# Check architecture rules
make arch-go.check
```

## Configuration
- Environment-based configuration using `envconfig`
- Default values defined in `shared.Config` struct
- `.env` file support for development
- Production vs development mode detection via `BARALGA_ENV`