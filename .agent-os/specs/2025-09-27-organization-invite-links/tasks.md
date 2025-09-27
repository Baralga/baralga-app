# Spec Tasks

## Tasks

- [ ] 1. Database Schema Implementation
  - [ ] 1.1 Write tests for organization_invites table migration
  - [ ] 1.2 Create migration files (000007_organization_invites.up.sql and 000007_organization_invites.down.sql)
  - [ ] 1.3 Add domain model for OrganizationInvite in user domain
  - [ ] 1.4 Update repository interfaces for invite management
  - [ ] 1.5 Implement database repository for organization invites
  - [ ] 1.6 Implement in-memory repository for organization invites (testing)
  - [ ] 1.7 Verify all tests pass

- [ ] 2. Invite Link Generation and Management
  - [ ] 2.1 Write tests for invite generation service
  - [ ] 2.2 Implement OrganizationInviteService with token generation
  - [ ] 2.3 Add invite generation to UserService
  - [ ] 2.4 Implement invite validation and expiration logic
  - [ ] 2.5 Add invite revocation functionality
  - [ ] 2.6 Write tests for invite management REST endpoints
  - [ ] 2.7 Implement REST API endpoints for invite management
  - [ ] 2.8 Verify all tests pass

- [ ] 3. Admin Interface for Invite Management
  - [ ] 3.1 Write tests for admin web handlers
  - [ ] 3.2 Implement admin web handlers for invite management
  - [ ] 3.3 Create HTML templates for invite management interface
  - [ ] 3.4 Add HTMX interactions for invite generation and revocation
  - [ ] 3.5 Integrate invite management into organization settings page
  - [ ] 3.6 Add proper error handling and user feedback
  - [ ] 3.7 Verify all tests pass

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
