# Data Model: Organization Management

## Core Entities

### Organization Entity
```go
type Organization struct {
    ID    uuid.UUID
    Title string  // Organization name
}
```

**Fields**:
- `ID`: Unique identifier (UUID)
- `Title`: Organization name (string, 3-100 characters)

**Validation Rules**:
- Title: required, min=3, max=100, alphanumeric with spaces/hyphens/underscores
- Trim whitespace before validation

**State Transitions**:
- Create: New organization with title
- Update: Modify existing organization title
- No delete operation (not in scope)

### User Entity (Extended)
```go
type User struct {
    ID             uuid.UUID
    Name           string
    Username       string
    EMail          string
    Password       string
    Origin         string
    OrganizationID uuid.UUID
}
```

**Authorization**:
- Users have roles within organizations
- ROLE_ADMIN can manage organization settings
- ROLE_USER cannot access organization management

### Principal Entity (Context)
```go
type Principal struct {
    Name           string
    Username       string
    OrganizationID uuid.UUID
    Roles          []string
}
```

**Authorization Methods**:
- `HasRole(role string) bool`: Check if user has specific role
- Used for admin-only access control

## Form Models

### Organization Form Model
```go
type organizationFormModel struct {
    CSRFToken string
    Name      string `validate:"required,min=3,max=100"`
}
```

**Fields**:
- `CSRFToken`: Security token for form submission
- `Name`: Organization name with validation

**Validation**:
- Name: required, 3-100 characters, alphanumeric with spaces/hyphens/underscores
- CSRF token validation on submission

## Repository Interfaces

### Organization Repository (Extended)
```go
type OrganizationRepository interface {
    InsertOrganization(ctx context.Context, organization *Organization) (*Organization, error)
    UpdateOrganization(ctx context.Context, organization *Organization) error
    FindOrganizationByID(ctx context.Context, organizationID uuid.UUID) (*Organization, error)
}
```

**Methods**:
- `UpdateOrganization`: Update existing organization
- `FindOrganizationByID`: Retrieve organization by ID
- Existing `InsertOrganization` method preserved

## Service Layer

### User Service (Extended)
```go
type UserService struct {
    userRepository         UserRepository
    organizationRepository OrganizationRepository
}

func (s *UserService) UpdateOrganization(ctx context.Context, organizationID uuid.UUID, name string) error
func (s *UserService) FindOrganizationByID(ctx context.Context, organizationID uuid.UUID) (*Organization, error)
```

**Business Logic**:
- Validate organization name
- Check user authorization (admin role)
- Update organization in repository
- Handle validation errors

## Web Handlers

### Organization Web Handlers
```go
type OrganizationWebHandlers struct {
    config             *shared.Config
    userService        *UserService
    organizationRepository OrganizationRepository
}

func (h *OrganizationWebHandlers) HandleOrganizationDialog() http.HandlerFunc
func (h *OrganizationWebHandlers) HandleOrganizationUpdate() http.HandlerFunc
```

**Endpoints**:
- `GET /profile/organization`: Show organization management dialog
- `POST /profile/organization`: Update organization name

**Authorization**:
- Check `Principal.HasRole("ROLE_ADMIN")`
- Return 403 for non-admin users
- Use existing Principal context

## Database Schema

### Organizations Table (Existing)
```sql
CREATE TABLE organizations (
    org_id       uuid not null,
    title        varchar(255),
    description  varchar(4000)
);
```

**Updates Required**:
- No schema changes needed
- Use existing `title` field for organization name
- Existing constraints and indexes preserved

## Validation Rules

### Organization Name Validation
- **Required**: Cannot be empty
- **Length**: 3-100 characters
- **Characters**: Letters, numbers, spaces, hyphens, underscores
- **Trim**: Remove leading/trailing whitespace
- **Uniqueness**: Not enforced (per specification clarification needed)

### Form Validation
- **CSRF Token**: Required for all POST requests
- **Input Sanitization**: HTML escape for XSS prevention
- **Error Messages**: User-friendly validation messages

## Error Handling

### Validation Errors
- Field-specific error messages
- Display in form with Bootstrap alert styling
- Preserve form state on validation failure

### Authorization Errors
- 403 Forbidden for non-admin users
- Redirect to profile page with error message
- Log authorization attempts

### Database Errors
- Log technical details
- Show user-friendly error message
- Preserve form state for retry
