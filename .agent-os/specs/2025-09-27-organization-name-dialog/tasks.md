# Spec Tasks

## Tasks

- [x] 1. **Add Organization Update Functionality**
  - [x] 1.1 Write tests for organization repository update functionality
  - [x] 1.2 Add UpdateOrganization method to OrganizationRepository interface
  - [x] 1.3 Implement UpdateOrganization in DbOrganizationRepository
  - [x] 1.4 Implement UpdateOrganization in InMemOrganizationRepository
  - [x] 1.5 Add UpdateOrganizationName method to UserService
  - [x] 1.6 Implement UpdateOrganizationName in UserService with role validation
  - [x] 1.7 Verify all tests pass

- [x] 2. **Create Organization Dialog Web Handlers**
  - [x] 2.1 Write tests for organization web handlers
  - [x] 2.2 Create organizationFormModel struct for form handling
  - [x] 2.3 Add HandleOrganizationDialog handler for displaying organization dialog
  - [x] 2.4 Add HandleOrganizationUpdate handler for processing organization name updates
  - [x] 2.5 Implement organization dialog HTML template with role-based input field
  - [x] 2.6 Add CSRF protection to organization update form
  - [x] 2.7 Verify all tests pass

- [x] 3. **Integrate Organization Dialog into Navbar**
  - [x] 3.1 Write tests for navbar organization dialog integration
  - [x] 3.2 Add organization name display to navbar with click handler
  - [x] 3.3 Add organization dialog trigger button/link in navbar
  - [x] 3.4 Implement HTMX integration for dialog opening
  - [x] 3.5 Add role-based styling for organization name display
  - [x] 3.6 Verify all tests pass
