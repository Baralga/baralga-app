# Implementation Plan

- [x] 1. Create database migration for tags tables
  - Create migration file 000005_add_tags.up.sql for tags and activity_tags tables with proper indexes and unique constraints
  - Enable pg_trgm extension for text search
  - Add unique constraint on (name, org_id) in tags table for organization-level uniqueness
  - Create indexes for efficient tag queries and autocomplete functionality
  - _Requirements: 1.1, 1.3, 2.1_

- [x] 2. Implement Tag domain model and repository interfaces
  - Create Tag struct with ID, Name, OrganizationID, CreatedAt fields in tracking/activity_domain.go
  - Define TagRepository interface with FindTagsByOrganization, FindOrCreateTag, SyncTagsForActivity methods
  - Add Tags field ([]string) to existing Activity struct in tracking/activity_domain.go
  - Update ActivitiesFilter struct to include tag filtering fields
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 3. Implement database tag repository
  - Create tracking/tag_repository_db.go with DbTagRepository struct implementing TagRepository interface
  - Implement FindTagsByOrganization method with trigram text search for autocomplete
  - Implement FindOrCreateTag method with case-insensitive tag creation at organization level
  - Implement SyncTagsForActivity method to manage activity-tag relationships
  - Implement DeleteUnusedTags cleanup method for organization-level cleanup
  - _Requirements: 1.1, 1.3, 2.1, 3.1, 3.2_

- [x] 4. Extend activity repository to handle tags
  - Modify FindActivities method in tracking/activity_repository_db.go to load tags for each activity
  - Modify InsertActivity method to sync tags after creation using TagRepository
  - Modify UpdateActivity methods to sync tags after updates using TagRepository
  - Add tag filtering support to FindActivities method with JOIN queries
  - _Requirements: 1.1, 1.4, 4.1, 4.2_

- [x] 5. Create TagService for business logic
  - Create tracking/tag_service.go with TagService struct
  - Implement GetTagsForAutocomplete method with fuzzy matching within organization
  - Implement NormalizeTagName method for case-insensitive handling
  - Implement ParseTagsFromString method to split comma/space separated tags
  - Implement ValidateTags method for tag validation (max 10 tags, length limits)
  - _Requirements: 2.1, 2.2, 3.1, 3.3, 3.5_

- [x] 6. Extend ActivityService to handle tags
  - Modify existing activity CRUD methods in tracking/activity_service.go to handle tag synchronization
  - Integrate TagService for tag normalization and validation
  - Update activity filtering to support tag-based queries
  - Add TagService dependency to ActivityService constructor
  - _Requirements: 1.1, 1.4, 1.5, 4.1, 4.2_

- [x] 7. Create tag autocomplete API endpoint
  - Add GET /api/tags/autocomplete endpoint to tracking/activity_rest.go
  - Implement query parameter handling for tag search within organization
  - Return JSON response with matching tag suggestions from organization
  - Add proper error handling and validation
  - _Requirements: 3.1, 3.2, 3.3, 3.5_

- [x] 8. Update activity form to include tag input
  - Add Tags field to activityFormModel struct in tracking/activity_web.go with validation
  - Modify ActivityForm function to include tag input field
  - Use a simple text input without any magic where tags are entered with space or comma as separator
  - Update mapFormToActivity and mapActivityToForm functions to handle tags
  - _Requirements: 1.1, 1.2, 1.5, 3.1_

- [ ] 9. Implement tag filtering in activity list UI
  - Add tag filter input field to activity list page in tracking/activity_web.go
  - Update HandleTrackingPage to parse tag filter parameters from URL
  - Modify activity list queries to pass tag filters to the service layer
  - Add tag filter state management in URL parameters for pagination
  - Update activity list templates to show active tag filters with clear options
  - _Requirements: 4.1, 4.2, 4.4, 4.5_

- [x] 10. Add tag display to activity views
  - Update ActivitiesSumByDayView template to display tags for each activity
  - Add CSS styling for tag display (badges/chips) in activity list
  - Ensure tags are displayed consistently across all activity views
  - _Requirements: 1.4, 4.1_

- [ ] 11. Implement tag color generation system
  - Add GetTagColor method to TagService for consistent color generation based on tag name hash
  - Create a predefined set of accessible colors with sufficient contrast ratios
  - Update activity web templates to use generated colors instead of generic bg-light class
  - Ensure the same tag always has the same color across all views and users within organization
  - Add fallback to default neutral color if color generation fails
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 12. Write database repository tests for tag functionality
  - Create tracking/tag_repository_db_test.go with comprehensive tests for DbTagRepository
  - Test FindTagsByOrganization with trigram search functionality
  - Test FindOrCreateTag with case-insensitive handling and organization isolation
  - Test SyncTagsForActivity with proper transaction handling
  - Test DeleteUnusedTags cleanup functionality
  - Add integration tests using dockertest for database operations
  - _Requirements: All requirements_

- [ ] 13. Create tag statistics reporting
  - Add GET /api/tags/statistics endpoint for organization-wide tag usage data
  - Implement tag-based grouping in time reports respecting organization boundaries
  - Add tag statistics to report views and templates
  - Create database queries for organization-level tag usage analytics
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 14. Write comprehensive tests for tag functionality
  - Create unit tests for Tag domain model and validation in tracking/tag_service_test.go
  - Write repository tests for all tag CRUD operations in tracking/tag_repository_mem_test.go
  - Add service layer tests for tag business logic
  - Create integration tests for tag API endpoints in tracking/activity_rest_test.go
  - Write web handler tests for tag form processing in tracking/activity_web_test.go
  - Add end-to-end tests for complete tag workflows
  - _Requirements: All requirements_