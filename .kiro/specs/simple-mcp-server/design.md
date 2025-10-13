# Design Document

## Overview

The Simple Time Tracking MCP Server will be integrated into the existing Baralga web application under the `/mcp` root path. The server will follow the Model Context Protocol (MCP) specification and expose time tracking functionality through MCP tools via HTTP endpoints, allowing external applications to manage time entries and generate reports. The design leverages the existing Baralga domain components, specifically reusing the tracking domain's Activity entities, services, and repositories.

The MCP server will be accessible via HTTP requests to `/mcp/*` endpoints within the same web application, providing seamless integration while sharing the same codebase, configuration, database connections, and authentication system as the main application. This approach ensures consistency with the main application's data model and business rules while simplifying deployment and maintenance.

## Architecture

### High-Level Architecture

```mermaid
graph TB
    MCPClient[MCP Client Application] --> MCPEndpoints[/mcp/* HTTP Endpoints]
    WebClient[Web Browser] --> WebEndpoints[/api/* & / HTTP Endpoints]
    
    MCPEndpoints --> MCPHandlers[MCP Tool Handlers]
    WebEndpoints --> RestHandlers[REST API Handlers]
    
    MCPHandlers --> ActivityService[Existing ActivityService]
    RestHandlers --> ActivityService
    
    ActivityService --> ActivityRepository[Existing ActivityRepository]
    ActivityRepository --> Storage[(PostgreSQL Database)]
    
    subgraph "Baralga Web Application"
        MCPEndpoints
        WebEndpoints
        MCPHandlers
        RestHandlers
        ActivityService
        ActivityRepository
        Storage
    end
```

### MCP Protocol Integration

The server will implement the MCP specification by:
- Accepting JSON-RPC 2.0 messages over HTTP at `/mcp` endpoints
- Exposing tools through the MCP tools interface via HTTP POST requests
- Handling tool calls and returning structured JSON responses
- Supporting MCP initialization and capability negotiation via HTTP endpoints
- Using standard HTTP authentication and CORS headers for web compatibility

### Domain Architecture

The MCP server will be integrated into the existing Baralga layered architecture by adding MCP handlers alongside existing REST handlers:

```
baralga/
├── main.go                    # Existing entry point (updated to register MCP routes)
├── tracking/                  # Existing domain (extended)
│   ├── activity_domain.go     # Existing domain entities
│   ├── activity_service.go    # Existing business logic
│   ├── activity_repository_*.go # Existing data access
│   ├── activity_rest.go       # Existing REST API handlers
│   ├── activity_mcp.go        # NEW: MCP tool handlers
│   ├── project_rest.go        # Existing project REST handlers
│   └── project_mcp.go         # NEW: Project MCP tool handlers (if needed)
├── shared/                    # Existing shared components (extended)
│   ├── config.go
│   ├── shared_domain.go
│   ├── shared_rest.go         # Existing REST utilities
│   └── shared_mcp.go          # NEW: MCP protocol utilities
└── ...
```

The MCP integration follows existing patterns:
- `activity_mcp.go` - MCP tool handlers alongside `activity_rest.go`
- `shared_mcp.go` - MCP protocol utilities alongside `shared_rest.go`
- Reuses all existing domain entities, services, and repositories
- Follows the same layered architecture and naming conventions
- Integrates with existing Chi router and middleware stack

## Components and Interfaces

### MCP Server Core

**ActivityMCPHandlers** (in `tracking/activity_mcp.go`) - MCP tool handlers for activity operations
- Handles MCP tool calls for activity CRUD operations and reporting
- Manages JSON-RPC 2.0 message parsing and tool execution
- Integrates with existing `ActivityService` for business logic
- Uses existing `activityModel` structures for consistent responses
- Follows the same patterns as `ActivityRestHandlers`

**SharedMCPUtilities** (in `shared/shared_mcp.go`) - Common MCP protocol utilities
- MCP protocol message parsing and response formatting
- Tool registration and capability negotiation
- Error handling and MCP error response formatting
- Integration with existing Chi router for `/mcp/*` routes

### Reused Domain Components

**Activity** - Existing domain entity from `tracking.Activity`
```go
type Activity struct {
    ID             uuid.UUID
    Start          time.Time
    End            time.Time
    Description    string
    ProjectID      uuid.UUID
    OrganizationID uuid.UUID
    Username       string
    Tags           []*Tag
}
```

**ActivityRepository** - Existing data access interface from `tracking.ActivityRepository`
- `FindActivityByID(ctx, activityID, organizationID)` - Get single activity
- `InsertActivity(ctx, activity)` - Create new activity
- `UpdateActivity(ctx, organizationID, activity)` - Update existing activity
- `DeleteActivityByID(ctx, organizationID, activityID)` - Delete activity
- `FindActivities(ctx, filter, pageParams)` - List activities with filtering
- `TimeReportByDay/Week/Month/Quarter(ctx, filter)` - Time aggregation reports
- `ProjectReport(ctx, filter)` - Project-based reports

**ActivityService** - Existing business logic from `tracking.ActivityService`
- `CreateActivity(ctx, principal, activity)` - Create with validation and tags
- `UpdateActivity(ctx, principal, activity)` - Update with validation
- `DeleteActivityByID(ctx, principal, activityID)` - Delete with authorization
- `ReadActivitiesWithProjects(ctx, principal, filter, pageParams)` - List with projects
- `TimeReports(ctx, principal, filter, aggregateBy)` - Time summaries
- `ProjectReports(ctx, principal, filter)` - Project summaries

### MCP Tool Handlers

**ActivityTools** - MCP tool implementations that bridge MCP calls to existing services
- `create_entry` - Maps to `ActivityService.CreateActivity()` with proper principal context
- `get_entry` - Maps to `ActivityRepository.FindActivityByID()` with organization filtering
- `update_entry` - Maps to `ActivityService.UpdateActivity()` with validation
- `delete_entry` - Maps to `ActivityService.DeleteActivityByID()` with authorization
- `list_entries` - Maps to `ActivityService.ReadActivitiesWithProjects()` with filtering
- `get_summary` - Maps to `ActivityService.TimeReports()` with period aggregation
- `get_hours_by_project` - Maps to `ActivityService.ProjectReports()` with date filtering

Each tool handler will:
- Parse and validate MCP tool call parameters
- Create appropriate `shared.Principal` context for authorization
- Convert MCP requests to existing service method calls
- Transform existing domain objects to MCP response format
- Handle existing domain errors and convert to MCP error responses

## Data Models

### Core Entities (Reused from Baralga)

**Activity** (from `tracking.Activity`)
- `ID`: UUID identifier
- `Start`: Entry start timestamp
- `End`: Entry end timestamp  
- `Description`: Text description of work performed
- `ProjectID`: UUID reference to project
- `OrganizationID`: UUID for multi-tenancy
- `Username`: User who created the entry
- `Tags`: Associated tags with colors

**ActivitiesFilter** (from `tracking.ActivitiesFilter`)
- `Start`: Start date filter
- `End`: End date filter
- `OrganizationID`: Organization context
- `Username`: User filter (for non-admin access)
- `SortBy`: Sort field
- `SortOrder`: Sort direction

**ActivityTimeReportItem** (from `tracking.ActivityTimeReportItem`)
- `Year/Quarter/Month/Week/Day`: Time period identifiers
- `DurationInMinutesTotal`: Total minutes for the period

### MCP Request/Response Models (Reusing Existing API Structures)

**activityModel** (from `tracking/activity_rest.go`)
```go
type activityModel struct {
    ID          string         `json:"id"`
    Start       string         `json:"start" validate:"required"`
    End         string         `json:"end" validate:"required"`
    Description string         `json:"description" validate:"max=500"`
    Duration    *durationModel `json:"duration"`
    Links       *hal.Links     `json:"_links"`
}

type durationModel struct {
    Hours     int     `json:"hours"`
    Minutes   int     `json:"minutes"`
    Decimal   float64 `json:"decimal"`
    Formatted string  `json:"formatted"`
}
```

**projectModel** (from `tracking/project_rest.go`)
```go
type projectModel struct {
    ID          string     `json:"id"`
    Title       string     `json:"title" validate:"required,min=3,max=100"`
    Description string     `json:"description" validate:"max=500"`
    Active      bool       `json:"active"`
    Links       *hal.Links `json:"_links"`
}
```

**MCP Tool Parameters** (Simple parameter structures for MCP tools)
- MCP tools will accept simple parameter objects with basic validation
- Responses will use the existing `activityModel` and `projectModel` structures
- Existing mapping functions (`mapToActivity`, `mapToActivityModel`, etc.) will be reused
- This ensures consistency between REST API and MCP server responses

## Error Handling

### Error Categories

1. **Validation Errors** - Invalid input parameters, constraint violations
2. **Not Found Errors** - Requested resources don't exist
3. **Business Logic Errors** - Domain rule violations (e.g., end time before start time)
4. **System Errors** - Database connection issues, internal server errors

### Error Response Format

All errors will be returned as MCP error responses with structured error information:
```json
{
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": {
      "type": "validation_error",
      "details": "End time must be after start time"
    }
  }
}
```

### Error Handling Strategy

- Input validation at the tool handler level using go-playground/validator
- Business rule validation in the service layer
- Repository errors wrapped and propagated with context
- Consistent error response formatting across all tools
- Logging of errors for debugging and monitoring

## Testing Strategy

### Unit Testing Approach

**MCP Tool Handler Testing**
- Use existing in-memory repository implementations (`tracking.*RepositoryMem`) for fast, isolated testing
- Wire up real ActivityService with in-memory repositories for authentic business logic testing
- Test MCP request/response formatting and validation
- Verify parameter parsing and error handling
- Test tool call routing and response marshaling
- Validate principal context creation and authorization

**Service Integration Testing**
- Use existing in-memory repository implementations from Baralga
- Test complete flow from MCP tool call through real services to in-memory storage
- Verify compatibility with existing domain validation rules
- Test error propagation from domain layer to MCP responses
- Validate business logic like tag handling, authorization, and data transformations

**MCP Protocol Testing**
- Test JSON-RPC 2.0 compliance with MCP specification
- Verify tool discovery and capability negotiation
- Test error response formatting according to MCP standards

### Integration Testing

**End-to-End MCP Testing**
- Test complete MCP protocol communication flow using existing PostgreSQL database
- Verify tool discovery and capability negotiation
- Test real database operations through existing repositories
- Validate JSON-RPC 2.0 compliance with actual Baralga data

**Database Integration**
- Reuse existing PostgreSQL database and schema from Baralga
- Test with existing migration system and data structures
- Leverage existing transaction handling through `shared.RepositoryTxer`
- Use existing test data fixtures and organization setup

### Test Data Management

- Reuse existing Baralga test fixtures and factory functions
- Leverage existing in-memory repository implementations for fast tests
- Use existing organization and user setup from Baralga test utilities
- Maintain compatibility with existing database cleanup and isolation patterns