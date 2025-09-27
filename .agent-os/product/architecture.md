# Architecture

## System Architecture

Baralga follows a clean, domain-driven architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Web Interface Layer                      │
├─────────────────────────────────────────────────────────────┤
│  Web Handlers (HTML/HTMX)  │  REST Handlers (JSON API)    │
├─────────────────────────────────────────────────────────────┤
│                    Service Layer                           │
│  Auth Service  │  User Service  │  Activity Service        │
├─────────────────────────────────────────────────────────────┤
│                   Repository Layer                         │
│  User Repo     │  Activity Repo │  Project Repo            │
├─────────────────────────────────────────────────────────────┤
│                    Database Layer                          │
│                    PostgreSQL                              │
└─────────────────────────────────────────────────────────────┘
```

## Domain Structure

### Authentication Domain (`auth/`)
- **Purpose**: User authentication and authorization
- **Key Components**:
  - JWT token management
  - OAuth integration (GitHub, Google)
  - Password authentication
  - CSRF protection
- **Files**: `auth_service.go`, `auth_rest.go`, `auth_web.go`

### User Management Domain (`user/`)
- **Purpose**: User and organization management
- **Key Components**:
  - User CRUD operations
  - Organization management
  - Role-based access control
  - User registration and activation
- **Files**: `user_service.go`, `user_web.go`, `organization_repository_*.go`

### Time Tracking Domain (`tracking/`)
- **Purpose**: Core time tracking functionality
- **Key Components**:
  - Activity management
  - Project management
  - Tag system
  - Reporting and analytics
- **Files**: `activity_*.go`, `project_*.go`, `tag_*.go`, `report_web.go`

### Shared Infrastructure (`shared/`)
- **Purpose**: Common utilities and infrastructure
- **Key Components**:
  - Configuration management
  - Database connections
  - HTML templating
  - Security middleware
  - Domain types and interfaces
- **Files**: `config.go`, `shared_domain.go`, `shared_rest.go`, `shared_web.go`

## Data Model

### Core Entities

#### Organizations
- Multi-tenant isolation
- Each organization has its own projects, users, and activities
- UUID-based identification

#### Users
- Belong to a single organization
- Role-based permissions (USER, ADMIN)
- OAuth and password authentication support

#### Projects
- Organization-scoped
- Active/inactive status
- Associated with activities

#### Activities
- Time tracking entries
- Associated with projects and users
- Support for multiple tags
- Start/end time tracking

#### Tags
- Organization-scoped
- Color-coded categorization
- Many-to-many relationship with activities

## Security Architecture

### Authentication Flow
1. User provides credentials (username/password or OAuth)
2. System validates credentials
3. JWT token generated with user context
4. Token stored in secure HTTP-only cookie
5. Subsequent requests validated via JWT middleware

### Authorization Model
- **Principal-based**: All operations require authenticated principal
- **Organization-scoped**: Data access limited to user's organization
- **Role-based**: Different permissions for USER vs ADMIN roles
- **Resource-level**: Fine-grained access control per resource type

### Security Measures
- **CSRF Protection**: All state-changing operations protected
- **Security Headers**: Comprehensive security headers via middleware
- **Password Hashing**: BCrypt with configurable strength
- **JWT Security**: Secure token generation and validation
- **Input Validation**: Request validation and sanitization

## API Design

### REST API Structure
```
/api/
├── /auth/login          # Authentication endpoints
├── /activities          # Activity management
├── /projects           # Project management
└── /reports            # Reporting endpoints
```

### Web Interface Structure
```
/
├── /login              # Authentication pages
├── /activities         # Activity management UI
├── /projects          # Project management UI
├── /reports           # Reporting interface
└── /assets/           # Static assets
```

## Database Design

### Key Tables
- `organizations` - Multi-tenant root entities
- `users` - User accounts with organization association
- `projects` - Project definitions within organizations
- `activities` - Time tracking entries
- `tags` - Categorization system
- `activity_tags` - Many-to-many relationship
- `roles` - User role assignments

### Indexing Strategy
- Organization-based indexes for multi-tenancy
- User-based indexes for performance
- Time-based indexes for reporting queries
- Composite indexes for complex queries

## Deployment Architecture

### Development Environment
- Local PostgreSQL database
- Docker Compose for services
- Environment-based configuration
- Hot reloading for development

### Production Considerations
- PostgreSQL with connection pooling
- Reverse proxy (nginx) for static assets
- SSL/TLS termination
- Health check endpoints
- Graceful shutdown handling

## Testing Strategy

### Unit Testing
- Service layer business logic
- Repository layer data access
- Domain model validation
- Utility function testing

### Integration Testing
- Database integration tests
- API endpoint testing
- Authentication flow testing
- End-to-end user workflows

### Test Data Management
- In-memory repositories for unit tests
- Docker-based database for integration tests
- Test data fixtures and factories
- Isolated test environments
