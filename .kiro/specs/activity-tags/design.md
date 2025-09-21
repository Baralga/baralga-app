# Design Document

## Overview

The activity tags feature will extend the existing time tracking system to support tagging activities for better organization and analysis. The design follows the existing architectural patterns in the codebase, using a layered approach with domain models, repositories, services, and web handlers.

The implementation will add a new `Tag` domain entity and establish a many-to-many relationship between activities and tags. Tags will be stored in a normalized format (lowercase) but displayed consistently to users. The system will provide autocomplete functionality and filtering capabilities while maintaining the existing performance characteristics.

## Architecture

### Database Schema Changes

The design introduces two new tables to support the tagging functionality and requires the PostgreSQL `pg_trgm` extension for text search:

1. **tags table**: Stores unique tags per organization
   - `tag_id` (UUID, primary key)
   - `name` (VARCHAR, normalized lowercase name)
   - `org_id` (UUID, foreign key to organizations)
   - `created_at` (TIMESTAMP)
   - Unique constraint: (name, org_id) - ensures tag uniqueness within organization


2. **activity_tags table**: Junction table for many-to-many relationship
   - `activity_id` (UUID, foreign key to activities)
   - `tag_id` (UUID, foreign key to tags)
   - `org_id` (UUID, for query optimization and access control)
   - Primary key: (activity_id, tag_id)

### Domain Model Extensions

The existing `Activity` struct will be extended to include tags:

```go
type Activity struct {
    // existing fields...
    Tags   []string  // slice of tag names for easy handling
} 

type Tag struct {
    ID             uuid.UUID
    Name           string // normalized (lowercase)
    OrganizationID uuid.UUID
    CreatedAt      time.Time
}
```

### Repository Layer

New repository interfaces and implementations:

1. **TagRepository**: Manages tag CRUD operations
   - `FindTagsByOrganization()`: Get all tags for specific organization for autocomplete
   - `FindOrCreateTag()`: Get existing or create new tag for organization
   - `FindTagsByActivity()`: Get tags for specific activity
   - `SyncTagsForActivity()`: Create/update tags when activity is saved
   - `DeleteUnusedTags()`: Cleanup orphaned tags for organization

2. **ActivityRepository Extensions**: 
   - Extend existing methods to handle tags
   - Add tag-based filtering capabilities
   - Modify `FindActivities()` to support tag filters
   - Update CRUD operations to manage tag relationships

## Components and Interfaces

### Service Layer

**ActivityService Extensions**:
- Extend existing methods to handle tag operations
- Add tag normalization logic (case-insensitive handling)
- Implement tag autocomplete functionality
- Add tag-based filtering and reporting

**New TagService**:
- `GetTagsForAutocomplete()`: Returns matching tags for autocomplete within organization
- `NormalizeTagName()`: Converts tags to lowercase for storage
- `GetTagStatistics()`: Generate tag usage reports for organization
- `SyncActivityTags()`: Automatically create/update/delete tags when activities change
- `EnsureTagsExist()`: Create tags if they don't exist in organization

### Web Layer

**Form Model Extensions**:
```go
type activityFormModel struct {
    // existing fields...
    Tags string `validate:"max=1000"` // comma-separated tag string
}
```

**New API Endpoints**:
- `GET /api/tags/autocomplete?q={query}`: Tag autocomplete
- `GET /api/tags/statistics`: Tag usage statistics

**UI Components**:
- Tag input field with autocomplete dropdown
- Tag filter component for activity lists
- Tag display badges/chips in activity views
- Tag statistics display in reports

## Data Models

### Organization-Level Tag Management

Tags are managed at the organization level to promote consistency and collaboration:

- **Tag Creation**: When a user creates a tag, it becomes available to all users in the organization
- **Tag Reuse**: Users can see and use tags created by other users in their organization
- **Autocomplete Scope**: Autocomplete suggestions include all tags used within the organization
- **Tag Statistics**: Reports can show organization-wide tag usage patterns
- **Access Control**: Users can only see tags from their own organization

**Design Rationale**: Organization-level tags promote consistency in categorization across teams, reduce duplicate tags with slight variations (e.g., "meeting" vs "meetings"), and enable better organization-wide reporting and analytics.

### Tag Input Processing

Tags will be processed as follows:
1. User input: "Meeting, Development, bug-fix"
2. Split by comma/space and trim whitespace
3. Normalize to lowercase: ["meeting", "development", "bug-fix"]
4. Remove duplicates and validate length
5. Create new tags automatically if they don't exist in the organization

**Design Rationale**: Automatic tag creation reduces friction for users while maintaining organization-level consistency. The comma/space separation provides flexibility in tag input methods.

### Tag Storage Strategy

- **Normalization**: All tags stored in lowercase for consistent querying
- **Organization-Level Uniqueness**: Each tag name is unique within an organization, enforced at database level with unique constraint on (name, org_id)
- **Shared Tags**: Tags are shared across all users within the same organization, promoting consistency and reducing duplication
- **Relationships**: Many-to-many through junction table linking activities to organization-level tags
- **Automatic Lifecycle**: Tags are created when activities are saved, updated when activities change, and cleaned up when no longer used by any activity in the organization
- **Cleanup**: Automatic removal of unused tags via background process at organization level

### Filtering Logic

Tag filtering will support:
- **Single tag**: Show activities with specific tag
- **Multiple tags (OR)**: Show activities with any of the selected tags
- **Case-insensitive matching**: "meeting" matches "Meeting"

## Error Handling

### Validation Rules

- Tag names: 1-50 characters, alphanumeric plus hyphens/underscores
- Maximum 10 tags per activity
- Total tag string length: max 1000 characters
- Duplicate tag handling: silently deduplicate

### Error Scenarios

1. **Invalid tag format**: Return validation error with specific message
2. **Tag limit exceeded**: Return error indicating maximum tag count
3. **Database constraints**: Handle unique constraint violations gracefully
4. **Orphaned tags**: Background cleanup process handles automatically

### Graceful Degradation

- If tag service is unavailable, activities still function without tags
- Autocomplete falls back to simple text input if API fails
- Tag filters gracefully handle missing or deleted tags

## Testing Strategy

### Unit Tests

1. **Domain Models**:
   - Tag normalization logic
   - Activity-tag relationship handling
   - Validation rules

2. **Repository Layer**:
   - Tag CRUD operations
   - Activity-tag junction table operations
   - Query filtering with tags
   - Database constraint handling

3. **Service Layer**:
   - Tag autocomplete functionality
   - Case-insensitive tag matching
   - Tag management operations
   - Statistics generation

4. **Web Layer**:
   - Form validation with tags
   - API endpoint responses
   - Tag input parsing
   - Autocomplete UI behavior

### Integration Tests

1. **End-to-End Workflows**:
   - Create activity with tags
   - Edit activity tags
   - Filter activities by tags
   - Tag management operations

2. **Database Integration**:
   - Tag relationship integrity
   - Performance with large tag datasets
   - Concurrent tag operations

3. **API Integration**:
   - Tag autocomplete performance
   - Tag filtering accuracy
   - Error handling scenarios

### Performance Tests

1. **Autocomplete Response Time**: < 200ms for 1000+ tags
2. **Tag Filtering**: Maintain current activity list performance
3. **Tag Statistics**: Generate reports within acceptable time limits
4. **Database Queries**: Ensure proper indexing for tag-related queries

## Database Indexing Strategy

### Required Indexes

1. **tags table**:
   - `idx_tags_org_name` on (org_id, name) - for exact lookups within organization
   - `idx_tags_name_text` GIN index on (name gin_trgm_ops) - for autocomplete text search
   - `idx_tags_org_id` on (org_id) - for organization-wide tag queries


2. **activity_tags table**:
   - `idx_activity_tags_activity` on (activity_id) - for activity queries
   - `idx_activity_tags_tag` on (tag_id) - for tag-based filtering
   - `idx_activity_tags_org` on (org_id, tag_id) - for organization-specific filtering

3. **activities table** (existing, may need updates):
   - Ensure existing indexes support tag join queries efficiently

### Query Optimization

- Use EXISTS clauses for tag filtering to leverage indexes
- Utilize PostgreSQL trigram similarity (`%` operator) for fuzzy tag autocomplete
- Use GIN indexes with pg_trgm for fast text search on tag names
- Batch tag operations to minimize database round trips
- Enable pg_trgm extension: `CREATE EXTENSION IF NOT EXISTS pg_trgm;`