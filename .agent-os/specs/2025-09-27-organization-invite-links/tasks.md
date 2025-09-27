# Spec Tasks

## Tasks

- [x] 1. Database Schema Implementation
  - [x] 1.1 Write tests for organization_invites table migration
  - [x] 1.2 Create migration files (000007_organization_invites.up.sql and 000007_organization_invites.down.sql)
  - [x] 1.3 Add domain model for OrganizationInvite in user domain
  - [x] 1.4 Update repository interfaces for invite management
  - [x] 1.5 Implement database repository for organization invites
  - [x] 1.6 Implement in-memory repository for organization invites (testing)
  - [x] 1.7 Verify all tests pass

- [x] 2. Invite Link Generation and Management
  - [x] 2.1 Write tests for invite generation service
  - [x] 2.2 Implement OrganizationInviteService with token generation
  - [x] 2.3 Add invite generation to UserService
  - [x] 2.4 Implement invite validation and expiration logic
  - [x] 2.5 Verify all tests pass

- [x] 3. Admin Interface for Invite Management
  - [x] 3.1 Write tests for admin web handlers
  - [x] 3.2 Implement admin web handlers for invite management
  - [x] 3.3 Create HTML templates for invite management interface
  - [x] 3.4 Add HTMX interactions for invite generation
  - [x] 3.5 Integrate invite creation into organization settings page
  - [x] 3.6 Add proper error handling and user feedback
  - [x] 3.7 Verify all tests pass

- [ ] 4. User Registration via Invite Links
  - [ ] 4.1 Write tests for invite validation in registration flow
  - [ ] 4.2 Implement invite token validation service
  - [ ] 4.3 Modify user registration to accept invite tokens
  - [ ] 4.4 Update UserService.SetUpNewUser to handle invite-based registration
  - [ ] 4.5 Implement automatic organization assignment for invited users
  - [ ] 4.6 Add ROLE_USER assignment for invited users
  - [ ] 4.7 Write tests for invite-based registration web handlers
  - [ ] 4.8 Implement web handlers for invite-based registration
  - [ ] 4.9 Create HTML templates for invite registration form
  - [ ] 4.10 Verify all tests pass

- [ ] 5. Integration and Security
  - [ ] 5.1 Write integration tests for complete invite flow
  - [ ] 5.2 Add security tests for invite token validation
  - [ ] 5.3 Implement proper error handling for expired/invalid invites
  - [ ] 5.4 Add cleanup job for expired invites (optional)
  - [ ] 5.5 Update main.go to wire up new services and handlers
  - [ ] 5.6 Add proper logging for invite operations
  - [ ] 5.7 Verify all tests pass and feature works end-to-end
