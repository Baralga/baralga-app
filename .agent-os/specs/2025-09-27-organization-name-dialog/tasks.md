# Spec Tasks

## Tasks

- [ ] 1. **Add Organization Update Functionality**
  - [ ] 1.1 Write tests for organization repository update functionality
  - [ ] 1.2 Add UpdateOrganization method to OrganizationRepository interface
  - [ ] 1.3 Implement UpdateOrganization in DbOrganizationRepository
  - [ ] 1.4 Implement UpdateOrganization in InMemOrganizationRepository
  - [ ] 1.5 Add UpdateOrganizationName method to UserService
  - [ ] 1.6 Implement UpdateOrganizationName in UserService with role validation
  - [ ] 1.7 Verify all tests pass

- [ ] 2. **Create Organization Dialog Web Handlers**
  - [ ] 2.1 Write tests for organization web handlers
  - [ ] 2.2 Create organizationFormModel struct for form handling
  - [ ] 2.3 Add HandleOrganizationDialog handler for displaying organization dialog
  - [ ] 2.4 Add HandleOrganizationUpdate handler for processing organization name updates
  - [ ] 2.5 Implement organization dialog HTML template with role-based input field
  - [ ] 2.6 Add CSRF protection to organization update form
  - [ ] 2.7 Verify all tests pass

- [ ] 3. **Integrate Organization Dialog into Navbar**
  - [ ] 3.1 Write tests for navbar organization dialog integration
  - [ ] 3.2 Add organization name display to navbar with click handler
  - [ ] 3.3 Add organization dialog trigger button/link in navbar
  - [ ] 3.4 Implement HTMX integration for dialog opening
  - [ ] 3.5 Add role-based styling for organization name display
  - [ ] 3.6 Verify all tests pass

- [ ] 4. **Add Organization Routes and Integration**
  - [ ] 4.1 Write tests for organization routes integration
  - [ ] 4.2 Add organization routes to main router configuration
  - [ ] 4.3 Register organization web handlers in main.go
  - [ ] 4.4 Test end-to-end organization name update workflow
  - [ ] 4.5 Verify all tests pass
