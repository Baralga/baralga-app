# Baralga App Constitution

## Core Principles

### I. Domain-Driven Design (DDD) Architecture
Every module must follow a strict layered DDD architecture with clear separation of concerns. The architecture enforces dependency rules that prevent circular dependencies and maintain clean boundaries between layers.

### II. Module-Based Organization
The application is organized into domain modules (user, auth, tracking, shared) where each module represents a distinct business capability. Dependencies between modules may only occur at the domain layer, with the shared module being completely independent.

### III. Layered Dependency Rules (NON-NEGOTIABLE)
The dependency flow is strictly enforced:
1. Presentation layer may depend on domain layer only
2. Domain layer may not have any dependencies to other layers
3. Infrastructure layer may depend on domain layer only

### IV. Clean Architecture Boundaries
Each layer has a specific responsibility:
- **Presentation Layer**: REST and Web interfaces
- **Domain Layer**: Services for use cases and business logic, repository interfaces, and simple domain structs
- **Infrastructure Layer**: Repository implementations and external system integrations

### V. Test-First Development
Domain logic must be thoroughly tested with unit tests. Infrastructure layer repositories (especially in-memory implementations) may be exempt from unit testing when they serve as simple data access patterns.

## Architecture Standards

### Module Structure
- Each domain module must contain all three layers (presentation, domain, infrastructure)
- Cross-module dependencies are only allowed at the domain layer
- The shared module must remain completely independent and not depend on any other module

### Layer Responsibilities
- **Presentation**: Handle HTTP requests, input validation, response formatting
- **Domain**: Business logic, use cases, domain models, repository contracts
- **Infrastructure**: Data persistence, external service integration, repository implementations

### Dependency Management
- No circular dependencies between modules
- Shared module serves as common utilities without domain-specific logic
- Domain interfaces define contracts that infrastructure implements

## Development Workflow

### Code Organization
- Follow the established module structure (user/, auth/, tracking/, shared/)
- Maintain clear separation between layers within each module
- Use dependency injection to maintain loose coupling

### Testing Strategy
- Unit tests for domain services and business logic
- Integration tests for cross-layer interactions
- Repository implementations may use in-memory alternatives for testing

### Code Review Requirements
- Verify layer dependencies follow the established rules
- Ensure domain logic is properly isolated
- Check that shared module remains independent
