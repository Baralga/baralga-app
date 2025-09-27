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


- [x] 8a. Store tag color in db
  - create a migration script to add tag color to tags table
  - create a struct for the tag with properties name and color
  - before calling SyncTagsForActivity assign a color to the tags
  - _Requirements: 4.1, 4.2, 4.4, 4.5_

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

- [x] 15. Extend TagService for report generation
  - Add GenerateTagReports method to TagService for comprehensive tag-based reporting
  - Implement GetTagReportData method to retrieve filtered tag report data for specific date ranges
  - Add logic to handle tag selection filtering (include/exclude specific tags)
  - Implement time aggregation by tag with daily/weekly/monthly breakdown support
  - Handle activities with multiple tags (count activity time for each associated tag)
  - _Requirements: 6.2, 6.3, 6.4, 6.5_

- [x] 16. Create Tag report category UI
  - Add "Tag" as a fourth report category in the existing reporting navigation interface
  - Create tag report page template following existing report category patterns
  - Implement tag selection interface for choosing a single tag
  - Maintain consistent UI patterns with existing general, time, and project report categories
  - _Requirements: 6.1, 6.2, 6.4_

- [ ] 17. Implement Tag report views and data display
  - Create Summary View showing total time and activity count per tag with visual indicators
  - Implement Detailed View with expandable sections showing individual activities per tag
  - Add Timeline View displaying tag usage over time with trend analysis
  - Create Comparison View for side-by-side comparison of selected tags
  - Display tag colors consistently throughout all report views for visual identification
  - _Requirements: 6.2, 6.3, 6.5_

- [ ] 19. Handle edge cases for Tag reporting
  - Implement appropriate messaging when no tagged activities are available for selected criteria
  - Add graceful handling of empty tag data in all report views and export functions
  - Ensure proper error handling when tag report generation fails or times out
  - Add validation for tag report parameters (date ranges, tag selections)
  - Test and handle scenarios with large datasets and multiple tag combinations
  - _Requirements: 6.7_

