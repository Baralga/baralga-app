# Implementation Plan

- [x] 1. Refactor existing MCP implementation to align with relaxed layered DDD architecture
  - Review and refactor tracking/activity_mcp.go to use either Domain Layer services OR repository interfaces from domain layer
  - Allow MCP handlers to access ActivityRepository and ProjectRepository interfaces defined in domain layer
  - Ensure MCP handlers do not access Infrastructure Layer implementations (ActivityRepositoryDB, etc.)
  - Update shared/shared_mcp.go to provide proper infrastructure support for Presentation Layer
  - Verify layer boundaries are respected (domain interfaces allowed, infrastructure implementations forbidden)
  - Test refactored implementation to ensure functionality is preserved
  - _Requirements: 1.1-1.5, 2.1-2.3, 3.1-3.5, 4.1-4.4, 5.1-5.7, 6.1-6.8, 7.1-7.6, 8.1-8.5_

- [x] 2. Add MCP Go SDK dependency and setup project structure
  - Add the official MCP Go SDK to go.mod dependencies
  - Create directory structure for MCP components
  - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

- [ ] 3. Implement shared MCP utilities and server setup (Infrastructure support for Presentation Layer)
  - [x] 3.1 Create shared/shared_mcp.go with MCP server initialization
    - Implement MCP server setup using the official Go SDK
    - Add HTTP transport layer integration with existing Chi router
    - Create tool registration and capability negotiation functions
    - Provide infrastructure support for Presentation Layer MCP handlers
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [x] 3.2 Implement API key authentication middleware (respecting layer boundaries)
    - Create middleware to extract API key from HTTP headers (X-API-Key or Authorization Bearer)
    - Validate email format and create shared.Principal context
    - Use existing authentication patterns without violating layer dependencies
    - Handle authentication errors with proper MCP error responses
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [x] 3.3 Add MCP error handling utilities
    - Create functions to convert domain errors to MCP error responses
    - Implement structured error response formatting according to MCP specification
    - Add logging for MCP-specific errors
    - _Requirements: 1.4, 2.2, 3.3, 4.2_

- [ ] 4. Implement MCP tool handlers for activity operations (Presentation Layer)
  - [x] 4.1 Create tracking/activity_mcp.go with core MCP tool handlers
    - Implement create_entry tool handler using ActivityService.CreateActivity() OR ActivityRepository.InsertActivity()
    - Implement get_entry tool handler using ActivityService.ReadActivitiesWithProjects() OR ActivityRepository.FindActivityByID()
    - Implement update_entry tool handler using ActivityService.UpdateActivity() OR ActivityRepository.UpdateActivity()
    - Implement delete_entry tool handler using ActivityService.DeleteActivityByID() OR ActivityRepository.DeleteActivityByID()
    - Ensure handlers access only Domain Layer services or repository interfaces, not Infrastructure Layer implementations
    - _Requirements: 1.1-1.5, 2.1-2.3, 3.1-3.5, 4.1-4.4_

  - [x] 4.2 Implement list_entries tool handler
    - Create tool handler using ActivityService.ReadActivitiesWithProjects() OR ActivityRepository.FindActivities()
    - Add parameter parsing for date range and project filtering
    - Implement response formatting using existing activityModel structures
    - Apply organization-based filtering when using repository interfaces directly
    - _Requirements: 5.1-5.7_

  - [x] 4.3 Implement time summary and reporting tools
    - Create get_summary tool handler using ActivityService.TimeReports() OR ActivityRepository.TimeReportByDay/Week/Month/Quarter()
    - Implement get_hours_by_project tool handler using ActivityService.ProjectReports() OR ActivityRepository.ProjectReport()
    - Add period type validation and date range processing
    - Apply organization-based filtering when using repository interfaces directly
    - _Requirements: 6.1-6.8, 7.1-7.6_

  - [x] 4.4 Implement project listing tool
    - Create list_projects tool handler using ProjectService.ReadProjects() OR ProjectRepository.FindProjects()
    - Implement response formatting to include project names and UUIDs
    - Add proper error handling for empty project lists
    - Ensure consistent ordering of projects in response
    - Apply organization-based filtering when using repository interfaces directly
    - _Requirements: 8.1-8.5_

  - [ ]* 4.5 Write unit tests for MCP tool handlers (layer-compliant testing)
    - Create unit tests for Presentation Layer handlers using mocked Domain Layer services OR in-memory repository implementations
    - Test MCP request/response formatting and validation at Presentation Layer
    - Test parameter parsing and error handling scenarios
    - Verify principal context creation and authentication flows
    - Ensure tests maintain layer boundaries (domain interfaces allowed, infrastructure implementations forbidden)
    - Include tests for both service-based and repository-based handler implementations
    - Include tests for list_projects tool handler
    - _Requirements: 1.1-1.5, 2.1-2.3, 3.1-3.5, 4.1-4.4, 5.1-5.7, 6.1-6.8, 7.1-7.6, 8.1-8.5_

- [ ] 5. Integrate MCP server with main application
  - [x] 5.1 Update main.go to register MCP routes
    - Add MCP server initialization to newApp() function
    - Register MCP routes under /mcp/* path prefix
    - Wire up MCP handlers with existing dependency injection
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [ ] 5.2 Add MCP route registration to router setup
    - Create MCP route group with authentication middleware
    - Register all MCP tool handlers with proper routing
    - Add CORS headers for web compatibility
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [ ]* 5.3 Write integration tests for MCP endpoints (layer-aware testing)
    - Create end-to-end tests using existing PostgreSQL database setup
    - Test complete MCP protocol communication flow through all layers
    - Verify tool discovery and capability negotiation at Presentation Layer
    - Test business logic through Domain Layer services OR repository interfaces (not infrastructure implementations)
    - Ensure layer boundaries are maintained throughout testing flow
    - Test both service-based and repository-based handler implementations
    - Test JSON-RPC 2.0 compliance with actual Baralga data
    - _Requirements: 1.1-1.5, 2.1-2.3, 3.1-3.5, 4.1-4.4, 5.1-5.7, 6.1-6.8, 7.1-7.6_

- [ ] 6. Implement MCP request/response models and validation
  - [ ] 6.1 Create MCP parameter structures for tool calls
    - Define parameter structures for create_entry, update_entry, delete_entry tools
    - Add parameter structures for list_entries with filtering options
    - Create parameter structures for get_summary and get_hours_by_project tools
    - Add parameter structure for list_projects tool (minimal parameters needed)
    - _Requirements: 1.1-1.5, 2.1-2.3, 3.1-3.5, 4.1-4.4, 5.1-5.7, 6.1-6.8, 7.1-7.6, 8.1-8.5_

  - [ ] 6.2 Add input validation for MCP tool parameters
    - Implement validation using go-playground/validator for all tool parameters
    - Add business rule validation (e.g., end time after start time)
    - Create validation error responses in MCP format
    - _Requirements: 1.2, 1.4, 3.2, 6.2-6.7_

  - [ ] 6.3 Implement response mapping functions
    - Create functions to convert Activity domain objects to MCP responses
    - Reuse existing activityModel and projectModel structures for consistency
    - Add response formatting for summary and report data
    - Add response mapping for project list with UUIDs and names
    - _Requirements: 1.5, 2.3, 3.4, 4.3, 5.1-5.7, 6.8, 7.6, 8.1-8.5_

- [ ] 7. Add MCP tool discovery and capability negotiation
  - [ ] 7.1 Implement MCP server capabilities
    - Register all available tools with the MCP server including list_projects
    - Add tool descriptions and parameter schemas for all tools
    - Implement capability negotiation according to MCP specification
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1, 8.1_

  - [ ] 7.2 Add MCP protocol compliance features
    - Implement JSON-RPC 2.0 message handling through SDK
    - Add proper error codes and message formatting
    - Ensure compatibility with MCP client applications
    - _Requirements: 1.4, 2.2, 3.3, 4.2, 5.7, 6.7, 7.5_