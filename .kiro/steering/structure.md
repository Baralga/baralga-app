# Project Structure & Architecture

## Domain-Driven Design
The project follows a domain-driven design with clear separation of concerns across business domains:

- **`auth/`** - Authentication and authorization logic
- **`tracking/`** - Core time tracking domain (activities, projects, tags)  
- **`user/`** - User and organization management
- **`shared/`** - Common utilities, domain types, and infrastructure

## Dependenies between Business Domains

The following dependencies between business domains are allowed:
- all may depend on shared, but shared may not depend on any other business domain
- auth may depend on user

## Layered Architecture Pattern
Each domain follows a consistent layered architecture:

```
domain/
├── *_domain.go      # Domain entities, value objects, interfaces or repositories, structs
├── *_service.go     # Business logic and use cases
├── *_repository_*.go # Data access implementations (db, mem)
├── *_rest.go        # REST API handlers
├── *_web.go         # Web UI handlers
└── *_test.go        # Unit tests
```

## Key Architectural Patterns

### Repository Pattern
- Interface definitions in domain files
- Database implementations: `*_repository_db.go`
- In-memory implementations: `*_repository_mem.go` (for testing)
- Transaction support via `RepositoryTxer` interface

### Service Layer
- Business logic encapsulated in service classes
- Services coordinate between repositories and handle transactions
- Context-based principal propagation for security

### Handler Separation in the Presentation Layer
- **REST handlers** (`*_rest.go`): JSON API endpoints under `/api`
- **Web handlers** (`*_web.go`): Server-side rendered HTML with HTMX

### Relaxed Layer Dependencies

The dependencies between the layers are:

* presentation layer -> service layer -> domain layer
* interface layer -> domain layer

So e.g. activity_rest my use the activity service from activity_service.go and ActivityRepository from activity_domain.go.

## File Organization

### Root Level
- `main.go` - Application entry point and dependency injection
- `Makefile` - Build automation and common tasks
- `go.mod/go.sum` - Go module dependencies
- `.env` - Local development configuration

### Shared Infrastructure
- `shared/assets/` - Static web assets (CSS, JS, images)
- `shared/migrations/` - Database schema migrations
- `shared/config.go` - Application configuration
- `shared/shared_domain.go` - Common domain types (Principal, interfaces)

### Testing Strategy
- Unit tests alongside source files (`*_test.go`)
- In-memory repository implementations for fast testing
- Integration tests using dockertest for database testing
- Test utilities: `matryer/is` for assertions, `testify` for mocking

## Naming Conventions
- **Files**: snake_case with domain prefix (e.g., `activity_service.go`)
- **Types**: PascalCase (e.g., `ActivityService`)
- **Interfaces**: PascalCase ending with interface purpose (e.g., `ActivityRepository`)
- **Database**: snake_case tables and columns
- **Constants**: SCREAMING_SNAKE_CASE or PascalCase for exported

## Multi-Tenancy
- Organization-based isolation using `OrganizationID` UUID
- All domain entities include `OrganizationID` for data segregation
- Principal context carries organization information for authorization