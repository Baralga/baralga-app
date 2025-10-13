# Implementation Plan

- [x] 1. Add MCP Go SDK dependency and setup project structure
  - Add the official MCP Go SDK to go.mod dependencies
  - Create directory structure for MCP components
  - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

- [ ] 2. Implement shared MCP utilities and server setup
  - [x] 2.1 Create shared/shared_mcp.go with MCP server initialization
    - Implement MCP server setup using the official Go SDK
    - Add HTTP transport layer integration with existing Chi router
    - Create tool registration and capability negotiation functions
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [x] 2.2 Implement API key authentication middleware
    - Create middleware to extract API key from HTTP headers (X-API-Key or Authorization Bearer)
    - Validate email format and lookup user using existing UserRepository
    - Create shared.Principal context from authenticated email address
    - Handle authentication errors with proper MCP error responses
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [x] 2.3 Add MCP error handling utilities
    - Create functions to convert domain errors to MCP error responses
    - Implement structured error response formatting according to MCP specification
    - Add logging for MCP-specific errors
    - _Requirements: 1.4, 2.2, 3.3, 4.2_

- [ ] 3. Implement MCP tool handlers for activity operations
  - [x] 3.1 Create tracking/activity_mcp.go with core MCP tool handlers
    - Implement create_entry tool handler mapping to ActivityService.CreateActivity()
    - Implement get_entry tool handler mapping to ActivityRepository.FindActivityByID()
    - Implement update_entry tool handler mapping to ActivityService.UpdateActivity()
    - Implement delete_entry tool handler mapping to ActivityService.DeleteActivityByID()
    - _Requirements: 1.1-1.5, 2.1-2.3, 3.1-3.5, 4.1-4.4_

  - [x] 3.2 Implement list_entries tool handler
    - Create tool handler mapping to ActivityService.ReadActivitiesWithProjects()
    - Add parameter parsing for date range and project filtering
    - Implement response formatting using existing activityModel structures
    - _Requirements: 5.1-5.7_

  - [x] 3.3 Implement time summary and reporting tools
    - Create get_summary tool handler mapping to ActivityService.TimeReports()
    - Implement get_hours_by_project tool handler mapping to ActivityService.ProjectReports()
    - Add period type validation and date range processing
    - _Requirements: 6.1-6.8, 7.1-7.6_

  - [ ]* 3.4 Write unit tests for MCP tool handlers
    - Create unit tests using in-memory repository implementations
    - Test MCP request/response formatting and validation
    - Test parameter parsing and error handling scenarios
    - Verify principal context creation and authorization flows
    - _Requirements: 1.1-1.5, 2.1-2.3, 3.1-3.5, 4.1-4.4, 5.1-5.7, 6.1-6.8, 7.1-7.6_

- [ ] 4. Integrate MCP server with main application
  - [ ] 4.1 Update main.go to register MCP routes
    - Add MCP server initialization to newApp() function
    - Register MCP routes under /mcp/* path prefix
    - Wire up MCP handlers with existing dependency injection
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [ ] 4.2 Add MCP route registration to router setup
    - Create MCP route group with authentication middleware
    - Register all MCP tool handlers with proper routing
    - Add CORS headers for web compatibility
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [ ]* 4.3 Write integration tests for MCP endpoints
    - Create end-to-end tests using existing PostgreSQL database setup
    - Test complete MCP protocol communication flow
    - Verify tool discovery and capability negotiation
    - Test JSON-RPC 2.0 compliance with actual Baralga data
    - _Requirements: 1.1-1.5, 2.1-2.3, 3.1-3.5, 4.1-4.4, 5.1-5.7, 6.1-6.8, 7.1-7.6_

- [ ] 5. Implement MCP request/response models and validation
  - [ ] 5.1 Create MCP parameter structures for tool calls
    - Define parameter structures for create_entry, update_entry, delete_entry tools
    - Add parameter structures for list_entries with filtering options
    - Create parameter structures for get_summary and get_hours_by_project tools
    - _Requirements: 1.1-1.5, 2.1-2.3, 3.1-3.5, 4.1-4.4, 5.1-5.7, 6.1-6.8, 7.1-7.6_

  - [ ] 5.2 Add input validation for MCP tool parameters
    - Implement validation using go-playground/validator for all tool parameters
    - Add business rule validation (e.g., end time after start time)
    - Create validation error responses in MCP format
    - _Requirements: 1.2, 1.4, 3.2, 6.2-6.7_

  - [ ] 5.3 Implement response mapping functions
    - Create functions to convert Activity domain objects to MCP responses
    - Reuse existing activityModel and projectModel structures for consistency
    - Add response formatting for summary and report data
    - _Requirements: 1.5, 2.3, 3.4, 4.3, 5.1-5.7, 6.8, 7.6_

- [ ] 6. Add MCP tool discovery and capability negotiation
  - [ ] 6.1 Implement MCP server capabilities
    - Register all available tools with the MCP server
    - Add tool descriptions and parameter schemas
    - Implement capability negotiation according to MCP specification
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [ ] 6.2 Add MCP protocol compliance features
    - Implement JSON-RPC 2.0 message handling through SDK
    - Add proper error codes and message formatting
    - Ensure compatibility with MCP client applications
    - _Requirements: 1.4, 2.2, 3.3, 4.2, 5.7, 6.7, 7.5_