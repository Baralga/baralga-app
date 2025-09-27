# Data Model: Organization Management

## Entities

### Organization (EXISTING)
**Purpose**: Represents the organizational entity that users belong to
**Attributes**:
- `ID`: Unique identifier (UUID) - org_id in database
- `Title`: Organization name (string, required, max 255 characters) - title in database

**Validation Rules**:
- Title must not be empty
- Title must be between 1 and 255 characters
- Title must be unique within the system
- Title cannot contain only whitespace

**State Transitions**:
- Organization can be created (initial state)
- Organization title can be updated (active state)
- Organization cannot be deleted (business rule)

**Relationships**:
- One-to-many with User (users belong to organization)
- Users have roles within organization (admin, regular user)

### User (EXISTING - Extended)
**Purpose**: Represents individual users with organization membership and roles
**Existing Attributes**: 
- `ID`: User unique identifier
- `Name`: User display name
- `Username`: Login username
- `EMail`: User email address
- `Password`: Encrypted password
- `Origin`: User origin
- `OrganizationID`: Foreign key to Organization (EXISTING)

**New Attributes**:
- `Role`: User role within organization (admin, user) - NEEDS TO BE ADDED

**Role Definitions**:
- `admin`: Can manage organization settings
- `user`: Regular user, cannot manage organization

## Database Schema Changes

### Organizations Table (EXISTING)
```sql
-- Table already exists
CREATE TABLE organizations (
     org_id       uuid not null,
     title        varchar(255),
     description  varchar(4000)
);
```

### Users Table (EXISTING - Needs Extension)
```sql
-- Users table already exists with organization_id
-- Need to add role column
ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user';

-- Create index for role lookups
CREATE INDEX idx_users_role ON users(role);

-- Add check constraint for valid roles
ALTER TABLE users ADD CONSTRAINT check_valid_role CHECK (role IN ('admin', 'user'));
```

## Domain Model

### Organization Domain Object (EXISTING)
```go
type Organization struct {
    ID    uuid.UUID
    Title string
}
```

### Organization Service Interface (NEW)
```go
type OrganizationService interface {
    GetOrganization(ctx context.Context, orgID uuid.UUID) (*Organization, error)
    UpdateOrganizationTitle(ctx context.Context, orgID uuid.UUID, title string) error
    IsUserAdmin(ctx context.Context, userID uuid.UUID, orgID uuid.UUID) (bool, error)
}
```

### Organization Repository Interface (EXISTING - Needs Extension)
```go
type OrganizationRepository interface {
    InsertOrganization(ctx context.Context, organization *Organization) (*Organization, error)
    // NEW METHODS NEEDED:
    FindByID(ctx context.Context, orgID uuid.UUID) (*Organization, error)
    Update(ctx context.Context, organization *Organization) error
    Exists(ctx context.Context, orgID uuid.UUID) (bool, error)
    FindByTitle(ctx context.Context, title string) (*Organization, error)
}
```

## Validation Rules

### Organization Title Validation
- **Required**: Title cannot be empty or nil
- **Length**: Must be between 1 and 255 characters
- **Format**: Must not be only whitespace
- **Uniqueness**: Must be unique across all organizations
- **Content**: Should not contain special characters that could cause issues

### Authorization Validation
- **Admin Check**: Only users with admin role can modify organization
- **Organization Membership**: User must belong to the organization being modified
- **Active User**: User must be active and authenticated

## Data Integrity

### Constraints
- Organization title must be unique
- User must belong to exactly one organization
- Only admin users can modify organization settings
- Organization cannot be deleted if it has users

### Referential Integrity
- Users.organization_id must reference existing organization
- Role must be one of: 'admin', 'user'
- Organization updates must maintain referential integrity

## Migration Strategy

### Database Migration
1. Add role column to users table
2. Create indexes for performance
3. Add foreign key constraints
4. Migrate existing data (if applicable)

### Application Migration
1. Extend existing organization repository with new methods
2. Add organization service for business logic
3. Add organization web interface for management
4. Update user service to handle organization relationships
5. Add web interface for organization management

## Implementation Approach

### Leverage Existing Infrastructure
- **Database**: Use existing organizations table
- **Domain Model**: Use existing Organization struct
- **Repository**: Extend existing OrganizationRepository interface
- **Repository Implementations**: Extend existing database and in-memory repositories

### Add New Components
- **Service Layer**: New OrganizationService for business logic
- **Web Interface**: New organization web handlers
- **Authorization**: Admin role checking
- **Validation**: Organization title validation

### Minimal Changes Required
- Extend existing repository interface with new methods
- Add organization service for business logic
- Add web interface for user interaction
- Add role-based authorization