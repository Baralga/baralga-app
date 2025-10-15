package tracking

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/baralga/shared"
	"github.com/baralga/shared/paged"
	time_utils "github.com/baralga/tracking/time"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
)

// ActivityMCPHandlers handles MCP tool calls for activity operations
type ActivityMCPHandlers struct {
	activityService    *ActitivityService
	activityRepository ActivityRepository
	projectRepository  ProjectRepository
	projectService     *ProjectService
	validator          *validator.Validate
}

// NewActivityMCPHandlers creates a new ActivityMCPHandlers instance
func NewActivityMCPHandlers(
	activityService *ActitivityService,
	activityRepository ActivityRepository,
	projectRepository ProjectRepository,
	projectService *ProjectService,
) *ActivityMCPHandlers {
	v := validator.New()

	// Register custom validation functions
	v.RegisterValidation("datetime", validateDateTime)
	v.RegisterValidation("date", validateDate)

	return &ActivityMCPHandlers{
		activityService:    activityService,
		activityRepository: activityRepository,
		projectRepository:  projectRepository,
		projectService:     projectService,
		validator:          v,
	}
}

// RegisterMCPTools registers all activity-related MCP tools
func (h *ActivityMCPHandlers) RegisterMCPTools(registrar shared.ToolRegistrar) {
	// Register create_entry tool
	registrar.AddTool(&mcp.Tool{
		Name:        "create_entry",
		Description: "Create a new time tracking entry with start/end times, description, and project association",
	}, h.createEntryHandler)

	// Register get_entry tool
	registrar.AddTool(&mcp.Tool{
		Name:        "get_entry",
		Description: "Retrieve a specific time entry by its ID",
	}, h.getEntryHandler)

	// Register update_entry tool
	registrar.AddTool(&mcp.Tool{
		Name:        "update_entry",
		Description: "Update an existing time entry with new values",
	}, h.updateEntryHandler)

	// Register delete_entry tool
	registrar.AddTool(&mcp.Tool{
		Name:        "delete_entry",
		Description: "Delete a time entry by its ID",
	}, h.deleteEntryHandler)

	// Register list_entries tool
	registrar.AddTool(&mcp.Tool{
		Name:        "list_entries",
		Description: "List time entries with optional filtering by date range and project",
	}, h.listEntriesHandler)

	// Register get_summary tool
	registrar.AddTool(&mcp.Tool{
		Name:        "get_summary",
		Description: "Get time summaries for specified periods (day/week/month/quarter/year)",
	}, h.getSummaryHandler)

	// Register get_hours_by_project tool
	registrar.AddTool(&mcp.Tool{
		Name:        "get_hours_by_project",
		Description: "Get hours grouped by project for a date range",
	}, h.getHoursByProjectHandler)

	// Register list_projects tool
	registrar.AddTool(&mcp.Tool{
		Name:        "list_projects",
		Description: "Retrieve a list of all available projects with their unique identifiers",
	}, h.listProjectsHandler)
}

// Individual tool handler functions

// createEntryHandler handles create_entry tool calls
func (h *ActivityMCPHandlers) createEntryHandler(ctx context.Context, req *mcp.CallToolRequest, arguments shared.ToolArguments) (*mcp.CallToolResult, shared.ToolResponse, error) {
	var params CreateEntryParams
	err := h.parseArguments(arguments, &params)
	if err != nil {
		shared.LogMCPError("create_entry", err, map[string]any{"arguments": arguments})
		return nil, nil, err
	}
	result, response, err := h.handleCreateEntry(ctx, req, params)
	return result, shared.ToolResponse(response.(map[string]any)), err
}

// getEntryHandler handles get_entry tool calls
func (h *ActivityMCPHandlers) getEntryHandler(ctx context.Context, req *mcp.CallToolRequest, arguments shared.ToolArguments) (*mcp.CallToolResult, shared.ToolResponse, error) {
	var params GetEntryParams
	err := h.parseArguments(arguments, &params)
	if err != nil {
		shared.LogMCPError("get_entry", err, map[string]any{"arguments": arguments})
		return nil, nil, err
	}
	result, response, err := h.handleGetEntry(ctx, req, params)
	return result, shared.ToolResponse(response.(map[string]any)), err
}

// updateEntryHandler handles update_entry tool calls
func (h *ActivityMCPHandlers) updateEntryHandler(ctx context.Context, req *mcp.CallToolRequest, arguments shared.ToolArguments) (*mcp.CallToolResult, shared.ToolResponse, error) {
	var params UpdateEntryParams
	err := h.parseArguments(arguments, &params)
	if err != nil {
		shared.LogMCPError("update_entry", err, map[string]any{"arguments": arguments})
		return nil, nil, err
	}
	result, response, err := h.handleUpdateEntry(ctx, req, params)
	return result, shared.ToolResponse(response.(map[string]any)), err
}

// deleteEntryHandler handles delete_entry tool calls
func (h *ActivityMCPHandlers) deleteEntryHandler(ctx context.Context, req *mcp.CallToolRequest, arguments shared.ToolArguments) (*mcp.CallToolResult, shared.ToolResponse, error) {
	var params DeleteEntryParams
	err := h.parseArguments(arguments, &params)
	if err != nil {
		shared.LogMCPError("delete_entry", err, map[string]any{"arguments": arguments})
		return nil, nil, err
	}
	result, response, err := h.handleDeleteEntry(ctx, req, params)
	if response == nil {
		return result, nil, err
	}
	return result, shared.ToolResponse(response.(map[string]any)), err
}

// listEntriesHandler handles list_entries tool calls
func (h *ActivityMCPHandlers) listEntriesHandler(ctx context.Context, req *mcp.CallToolRequest, arguments shared.ToolArguments) (*mcp.CallToolResult, shared.ToolResponse, error) {
	var params ListEntriesParams
	err := h.parseArguments(arguments, &params)
	if err != nil {
		shared.LogMCPError("list_entries", err, map[string]any{"arguments": arguments})
		return nil, nil, err
	}
	result, response, err := h.handleListEntries(ctx, req, params)
	return result, shared.ToolResponse(response.(map[string]any)), err
}

// getSummaryHandler handles get_summary tool calls
func (h *ActivityMCPHandlers) getSummaryHandler(ctx context.Context, req *mcp.CallToolRequest, arguments shared.ToolArguments) (*mcp.CallToolResult, shared.ToolResponse, error) {
	var params GetSummaryParams
	err := h.parseArguments(arguments, &params)
	if err != nil {
		shared.LogMCPError("get_summary", err, map[string]any{"arguments": arguments})
		return nil, nil, err
	}
	result, response, err := h.handleGetSummary(ctx, req, params)
	return result, shared.ToolResponse(response.(map[string]any)), err
}

// getHoursByProjectHandler handles get_hours_by_project tool calls
func (h *ActivityMCPHandlers) getHoursByProjectHandler(ctx context.Context, req *mcp.CallToolRequest, arguments shared.ToolArguments) (*mcp.CallToolResult, shared.ToolResponse, error) {
	var params GetHoursByProjectParams
	err := h.parseArguments(arguments, &params)
	if err != nil {
		shared.LogMCPError("get_hours_by_project", err, map[string]any{"arguments": arguments})
		return nil, nil, err
	}
	result, response, err := h.handleGetHoursByProject(ctx, req, params)
	return result, shared.ToolResponse(response.(map[string]any)), err
}

// listProjectsHandler handles list_projects tool calls
func (h *ActivityMCPHandlers) listProjectsHandler(ctx context.Context, req *mcp.CallToolRequest, arguments shared.ToolArguments) (*mcp.CallToolResult, shared.ToolResponse, error) {
	var params ListProjectsParams
	err := h.parseArguments(arguments, &params)
	if err != nil {
		shared.LogMCPError("list_projects", err, map[string]any{"arguments": arguments})
		return nil, nil, err
	}
	result, response, err := h.handleListProjects(ctx, req, params)
	return result, shared.ToolResponse(response.(map[string]any)), err
}

// MCP parameter structures for tool calls

// CreateEntryParams represents parameters for create_entry tool
type CreateEntryParams struct {
	Start       string   `json:"start,omitempty" validate:"omitempty,datetime" jsonschema:"description:Start time in ISO 8601 format (optional, defaults to current time)"`
	End         string   `json:"end,omitempty" validate:"omitempty,datetime" jsonschema:"description:End time in ISO 8601 format (optional, defaults to current time)"`
	Description string   `json:"description" validate:"required,max=500" jsonschema:"description:Description of the work performed"`
	ProjectID   string   `json:"project_id" validate:"required,uuid" jsonschema:"description:UUID of the project"`
	Tags        []string `json:"tags,omitempty" validate:"dive,max=50" jsonschema:"description:Optional array of tag names"`
}

// GetEntryParams represents parameters for get_entry tool
type GetEntryParams struct {
	EntryID string `json:"entry_id" validate:"required,uuid" jsonschema:"description:UUID of the time entry to retrieve"`
}

// UpdateEntryParams represents parameters for update_entry tool
type UpdateEntryParams struct {
	EntryID     string   `json:"entry_id" validate:"required,uuid" jsonschema:"description:UUID of the time entry to update"`
	Start       string   `json:"start,omitempty" validate:"omitempty,datetime" jsonschema:"description:Start time in ISO 8601 format (optional)"`
	End         string   `json:"end,omitempty" validate:"omitempty,datetime" jsonschema:"description:End time in ISO 8601 format (optional)"`
	Description string   `json:"description,omitempty" validate:"omitempty,max=500" jsonschema:"description:Description of the work performed (optional)"`
	ProjectID   string   `json:"project_id,omitempty" validate:"omitempty,uuid" jsonschema:"description:UUID of the project (optional)"`
	Tags        []string `json:"tags,omitempty" validate:"dive,max=50" jsonschema:"description:Optional array of tag names"`
}

// DeleteEntryParams represents parameters for delete_entry tool
type DeleteEntryParams struct {
	EntryID string `json:"entry_id" validate:"required,uuid" jsonschema:"description:UUID of the time entry to delete"`
}

// ListEntriesParams represents parameters for list_entries tool
type ListEntriesParams struct {
	FromDate  string `json:"from_date,omitempty" validate:"omitempty,date" jsonschema:"description:Start date filter in YYYY-MM-DD format (optional)"`
	ToDate    string `json:"to_date,omitempty" validate:"omitempty,date" jsonschema:"description:End date filter in YYYY-MM-DD format (optional)"`
	ProjectID string `json:"project_id,omitempty" validate:"omitempty,uuid" jsonschema:"description:UUID of the project to filter by (optional)"`
}

// GetSummaryParams represents parameters for get_summary tool
type GetSummaryParams struct {
	PeriodType string `json:"period_type" validate:"required,oneof=day week month quarter year" jsonschema:"description:Type of period (day/week/month/quarter/year)"`
	Date       string `json:"date,omitempty" validate:"omitempty,date" jsonschema:"description:Date within the period in YYYY-MM-DD format (optional, defaults to current date)"`
}

// GetHoursByProjectParams represents parameters for get_hours_by_project tool
type GetHoursByProjectParams struct {
	FromDate string `json:"from_date,omitempty" validate:"omitempty,date" jsonschema:"description:Start date in YYYY-MM-DD format (optional, defaults to first day of current month)"`
	ToDate   string `json:"to_date,omitempty" validate:"omitempty,date" jsonschema:"description:End date in YYYY-MM-DD format (optional, defaults to current date)"`
}

// ListProjectsParams represents parameters for list_projects tool (no parameters needed)
type ListProjectsParams struct {
}

// MCP tool handlers

// handleCreateEntry handles the create_entry MCP tool call
func (h *ActivityMCPHandlers) handleCreateEntry(ctx context.Context, req *mcp.CallToolRequest, params CreateEntryParams) (*mcp.CallToolResult, any, error) {
	// Extract principal from context
	principal := shared.MustPrincipalFromContext(ctx)

	// Apply defaults and validate (in case called directly from tests)
	err := h.applyDefaultValues(&params)
	if err != nil {
		return nil, nil, err
	}

	err = h.validateBusinessRules(&params)
	if err != nil {
		return nil, nil, h.createBusinessLogicError(err.Error())
	}

	// Parse start and end times
	startTime, err := time_utils.ParseDateTime(params.Start)
	if err != nil {
		shared.LogMCPError("create_entry", err, map[string]any{"start": params.Start})
		return nil, nil, errors.Wrap(err, "invalid start time format")
	}

	endTime, err := time_utils.ParseDateTime(params.End)
	if err != nil {
		shared.LogMCPError("create_entry", err, map[string]any{"end": params.End})
		return nil, nil, errors.Wrap(err, "invalid end time format")
	}

	// Parse project ID
	projectID, err := uuid.Parse(params.ProjectID)
	if err != nil {
		shared.LogMCPError("create_entry", err, map[string]any{"project_id": params.ProjectID})
		return nil, nil, errors.Wrap(err, "invalid project ID format")
	}

	// Verify project exists and user has access
	_, err = h.projectRepository.FindProjectByID(ctx, principal.OrganizationID, projectID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			shared.LogMCPError("create_entry", err, map[string]any{"project_id": params.ProjectID})
			return nil, nil, errors.New("project not found")
		}
		shared.LogMCPError("create_entry", err, map[string]any{"project_id": params.ProjectID})
		return nil, nil, errors.Wrap(err, "failed to verify project")
	}

	// Prepare tags
	tags := make([]*Tag, len(params.Tags))
	for i, tagName := range params.Tags {
		tags[i] = &Tag{Name: tagName}
	}

	// Create activity
	activity := &Activity{
		Start:       *startTime,
		End:         *endTime,
		Description: params.Description,
		ProjectID:   projectID,
		Tags:        tags,
	}

	// Create the activity using the service
	createdActivity, err := h.activityService.CreateActivity(ctx, principal, activity)
	if err != nil {
		shared.LogMCPError("create_entry", err, map[string]any{
			"activity": activity,
		})
		return nil, nil, errors.Wrap(err, "failed to create activity")
	}

	// Convert to response format
	response := h.mapActivityToMCPResponse(createdActivity)

	shared.LogMCPToolCall("create_entry", map[string]any{"params": params}, true)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created time entry with ID: %s", createdActivity.ID.String()),
			},
			&mcp.TextContent{
				Text: fmt.Sprintf("Entry details: %s", h.formatActivityJSON(response)),
			},
		},
	}, response, nil
}

// handleGetEntry handles the get_entry MCP tool call
func (h *ActivityMCPHandlers) handleGetEntry(ctx context.Context, req *mcp.CallToolRequest, params GetEntryParams) (*mcp.CallToolResult, any, error) {
	// Extract principal from context
	principal := shared.MustPrincipalFromContext(ctx)

	// Parse entry ID
	entryID, err := uuid.Parse(params.EntryID)
	if err != nil {
		shared.LogMCPError("get_entry", err, map[string]any{"entry_id": params.EntryID})
		return nil, nil, errors.Wrap(err, "invalid entry ID format")
	}

	// Find the activity
	activity, err := h.activityRepository.FindActivityByID(ctx, entryID, principal.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			shared.LogMCPError("get_entry", err, map[string]any{"entry_id": params.EntryID})
			return nil, nil, errors.New("activity not found")
		}
		shared.LogMCPError("get_entry", err, map[string]any{"entry_id": params.EntryID})
		return nil, nil, errors.Wrap(err, "failed to retrieve activity")
	}

	// Check if user has access to this activity (non-admin users can only see their own)
	if !principal.HasRole("ROLE_ADMIN") && activity.Username != principal.Username {
		err := errors.New("access denied: you can only view your own activities")
		shared.LogMCPError("get_entry", err, map[string]any{
			"entry_id":       params.EntryID,
			"activity_user":  activity.Username,
			"principal_user": principal.Username,
		})
		return nil, nil, err
	}

	// Convert to response format
	response := h.mapActivityToMCPResponse(activity)

	shared.LogMCPToolCall("get_entry", map[string]any{"params": params}, true)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Retrieved time entry: %s", activity.ID.String()),
			},
			&mcp.TextContent{
				Text: fmt.Sprintf("Entry details: %s", h.formatActivityJSON(response)),
			},
		},
	}, response, nil
}

// handleUpdateEntry handles the update_entry MCP tool call
func (h *ActivityMCPHandlers) handleUpdateEntry(ctx context.Context, req *mcp.CallToolRequest, params UpdateEntryParams) (*mcp.CallToolResult, any, error) {
	// Extract principal from context
	principal := shared.MustPrincipalFromContext(ctx)

	// Parse entry ID
	entryID, err := uuid.Parse(params.EntryID)
	if err != nil {
		shared.LogMCPError("update_entry", err, map[string]any{"entry_id": params.EntryID})
		return nil, nil, errors.Wrap(err, "invalid entry ID format")
	}

	// Find the existing activity
	existingActivity, err := h.activityRepository.FindActivityByID(ctx, entryID, principal.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			shared.LogMCPError("update_entry", err, map[string]any{"entry_id": params.EntryID})
			return nil, nil, errors.New("activity not found")
		}
		shared.LogMCPError("update_entry", err, map[string]any{"entry_id": params.EntryID})
		return nil, nil, errors.Wrap(err, "failed to retrieve activity")
	}

	// Check if user has access to update this activity (non-admin users can only update their own)
	if !principal.HasRole("ROLE_ADMIN") && existingActivity.Username != principal.Username {
		err := errors.New("access denied: you can only update your own activities")
		shared.LogMCPError("update_entry", err, map[string]any{
			"entry_id":       params.EntryID,
			"activity_user":  existingActivity.Username,
			"principal_user": principal.Username,
		})
		return nil, nil, err
	}

	// Create updated activity starting with existing values
	updatedActivity := &Activity{
		ID:             existingActivity.ID,
		Start:          existingActivity.Start,
		End:            existingActivity.End,
		Description:    existingActivity.Description,
		ProjectID:      existingActivity.ProjectID,
		OrganizationID: existingActivity.OrganizationID,
		Username:       existingActivity.Username,
		Tags:           existingActivity.Tags,
	}

	// Update fields if provided
	if params.Start != "" {
		startTime, err := time_utils.ParseDateTime(params.Start)
		if err != nil {
			shared.LogMCPError("update_entry", err, map[string]any{"start": params.Start})
			return nil, nil, errors.Wrap(err, "invalid start time format")
		}
		updatedActivity.Start = *startTime
	}

	if params.End != "" {
		endTime, err := time_utils.ParseDateTime(params.End)
		if err != nil {
			shared.LogMCPError("update_entry", err, map[string]any{"end": params.End})
			return nil, nil, errors.Wrap(err, "invalid end time format")
		}
		updatedActivity.End = *endTime
	}

	// Validate business rules after updates
	if updatedActivity.End.Before(updatedActivity.Start) || updatedActivity.End.Equal(updatedActivity.Start) {
		err := errors.New("end time must be after start time")
		shared.LogMCPError("update_entry", err, map[string]any{
			"start": updatedActivity.Start,
			"end":   updatedActivity.End,
		})
		return nil, nil, err
	}

	if params.Description != "" {
		updatedActivity.Description = params.Description
	}

	if params.ProjectID != "" {
		projectID, err := uuid.Parse(params.ProjectID)
		if err != nil {
			shared.LogMCPError("update_entry", err, map[string]any{"project_id": params.ProjectID})
			return nil, nil, errors.Wrap(err, "invalid project ID format")
		}

		// Verify project exists and user has access
		_, err = h.projectRepository.FindProjectByID(ctx, principal.OrganizationID, projectID)
		if err != nil {
			if errors.Is(err, ErrProjectNotFound) {
				shared.LogMCPError("update_entry", err, map[string]any{"project_id": params.ProjectID})
				return nil, nil, errors.New("project not found")
			}
			shared.LogMCPError("update_entry", err, map[string]any{"project_id": params.ProjectID})
			return nil, nil, errors.Wrap(err, "failed to verify project")
		}

		updatedActivity.ProjectID = projectID
	}

	if params.Tags != nil {
		tags := make([]*Tag, len(params.Tags))
		for i, tagName := range params.Tags {
			tags[i] = &Tag{Name: tagName}
		}
		updatedActivity.Tags = tags
	}

	// Update the activity using the service
	finalActivity, err := h.activityService.UpdateActivity(ctx, principal, updatedActivity)
	if err != nil {
		shared.LogMCPError("update_entry", err, map[string]any{
			"activity": updatedActivity,
		})
		return nil, nil, errors.Wrap(err, "failed to update activity")
	}

	// Convert to response format
	response := h.mapActivityToMCPResponse(finalActivity)

	shared.LogMCPToolCall("update_entry", map[string]any{"params": params}, true)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully updated time entry: %s", finalActivity.ID.String()),
			},
			&mcp.TextContent{
				Text: fmt.Sprintf("Updated entry details: %s", h.formatActivityJSON(response)),
			},
		},
	}, response, nil
}

// handleDeleteEntry handles the delete_entry MCP tool call
func (h *ActivityMCPHandlers) handleDeleteEntry(ctx context.Context, req *mcp.CallToolRequest, params DeleteEntryParams) (*mcp.CallToolResult, any, error) {
	// Extract principal from context
	principal := shared.MustPrincipalFromContext(ctx)

	// Parse entry ID
	entryID, err := uuid.Parse(params.EntryID)
	if err != nil {
		shared.LogMCPError("delete_entry", err, map[string]any{"entry_id": params.EntryID})
		return nil, nil, errors.Wrap(err, "invalid entry ID format")
	}

	// Check if activity exists and user has access (for better error messages)
	activity, err := h.activityRepository.FindActivityByID(ctx, entryID, principal.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			shared.LogMCPError("delete_entry", err, map[string]any{"entry_id": params.EntryID})
			return nil, nil, errors.New("activity not found")
		}
		shared.LogMCPError("delete_entry", err, map[string]any{"entry_id": params.EntryID})
		return nil, nil, errors.Wrap(err, "failed to retrieve activity")
	}

	// Check if user has access to delete this activity (non-admin users can only delete their own)
	if !principal.HasRole("ROLE_ADMIN") && activity.Username != principal.Username {
		err := errors.New("access denied: you can only delete your own activities")
		shared.LogMCPError("delete_entry", err, map[string]any{
			"entry_id":       params.EntryID,
			"activity_user":  activity.Username,
			"principal_user": principal.Username,
		})
		return nil, nil, err
	}

	// Delete the activity using the service
	err = h.activityService.DeleteActivityByID(ctx, principal, entryID)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			shared.LogMCPError("delete_entry", err, map[string]any{"entry_id": params.EntryID})
			return nil, nil, errors.New("activity not found")
		}
		shared.LogMCPError("delete_entry", err, map[string]any{"entry_id": params.EntryID})
		return nil, nil, errors.Wrap(err, "failed to delete activity")
	}

	shared.LogMCPToolCall("delete_entry", map[string]any{"params": params}, true)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully deleted time entry: %s", entryID.String()),
			},
		},
	}, nil, nil
}

// handleListEntries handles the list_entries MCP tool call
func (h *ActivityMCPHandlers) handleListEntries(ctx context.Context, req *mcp.CallToolRequest, params ListEntriesParams) (*mcp.CallToolResult, any, error) {
	// Extract principal from context
	principal := shared.MustPrincipalFromContext(ctx)

	// Apply defaults and validate (in case called directly from tests)
	err := h.applyDefaultValues(&params)
	if err != nil {
		return nil, nil, err
	}

	err = h.validateBusinessRules(&params)
	if err != nil {
		return nil, nil, h.createBusinessLogicError(err.Error())
	}

	// Create activity filter
	filter := &ActivityFilter{
		Timespan: TimespanCustom,
	}

	// Parse and set date filters
	fromDate, err := time_utils.ParseDate(params.FromDate)
	if err != nil {
		shared.LogMCPError("list_entries", err, map[string]any{"from_date": params.FromDate})
		return nil, nil, errors.Wrap(err, "invalid from_date format, expected YYYY-MM-DD")
	}
	filter.start = *fromDate

	toDate, err := time_utils.ParseDate(params.ToDate)
	if err != nil {
		shared.LogMCPError("list_entries", err, map[string]any{"to_date": params.ToDate})
		return nil, nil, errors.Wrap(err, "invalid to_date format, expected YYYY-MM-DD")
	}
	filter.end = *toDate

	// Verify project exists if project filter is provided
	var projectFilter *Project
	if params.ProjectID != "" {
		projectID, err := uuid.Parse(params.ProjectID)
		if err != nil {
			shared.LogMCPError("list_entries", err, map[string]any{"project_id": params.ProjectID})
			return nil, nil, errors.Wrap(err, "invalid project_id format")
		}

		project, err := h.projectRepository.FindProjectByID(ctx, principal.OrganizationID, projectID)
		if err != nil {
			if errors.Is(err, ErrProjectNotFound) {
				shared.LogMCPError("list_entries", err, map[string]any{"project_id": params.ProjectID})
				return nil, nil, errors.New("project not found")
			}
			shared.LogMCPError("list_entries", err, map[string]any{"project_id": params.ProjectID})
			return nil, nil, errors.Wrap(err, "failed to verify project")
		}
		projectFilter = project
	}

	// Use a large page size to get all matching entries (or implement pagination later)
	pageParams := &paged.PageParams{
		Page: 0,
		Size: 1000, // Large enough to get most results
	}

	// Get activities with projects
	activitiesPage, projects, err := h.activityService.ReadActivitiesWithProjects(ctx, principal, filter, pageParams)
	if err != nil {
		shared.LogMCPError("list_entries", err, map[string]any{
			"filter":     filter,
			"pageParams": pageParams,
		})
		return nil, nil, errors.Wrap(err, "failed to retrieve activities")
	}

	// Filter by project if specified (since ActivityFilter doesn't have project filtering built-in)
	var filteredActivities []*Activity
	if projectFilter != nil {
		for _, activity := range activitiesPage.Activities {
			if activity.ProjectID == projectFilter.ID {
				filteredActivities = append(filteredActivities, activity)
			}
		}
	} else {
		filteredActivities = activitiesPage.Activities
	}

	// Convert activities to MCP response format
	var responseEntries []map[string]any
	for _, activity := range filteredActivities {
		// Find the project for this activity
		var projectName string
		for _, project := range projects {
			if project.ID == activity.ProjectID {
				projectName = project.Title
				break
			}
		}

		entry := h.mapActivityToMCPResponse(activity)
		// Add project name for better readability
		entry["project_name"] = projectName
		responseEntries = append(responseEntries, entry)
	}

	response := map[string]any{
		"entries": responseEntries,
		"total":   len(responseEntries),
		"filters": map[string]any{
			"from_date":  params.FromDate,
			"to_date":    params.ToDate,
			"project_id": params.ProjectID,
		},
	}

	var resultText string
	if len(responseEntries) == 0 {
		resultText = "No time entries found matching the specified criteria"
	} else {
		resultText = fmt.Sprintf("Found %d time entries", len(responseEntries))
		resultText += fmt.Sprintf(" between %s and %s", params.FromDate, params.ToDate)
		if projectFilter != nil {
			resultText += fmt.Sprintf(" for project '%s'", projectFilter.Title)
		}
	}

	shared.LogMCPToolCall("list_entries", map[string]any{"params": params}, true)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: resultText,
			},
			&mcp.TextContent{
				Text: fmt.Sprintf("Entries: %s", h.formatEntriesJSON(responseEntries)),
			},
		},
	}, response, nil
}

// handleGetSummary handles the get_summary MCP tool call
func (h *ActivityMCPHandlers) handleGetSummary(ctx context.Context, req *mcp.CallToolRequest, params GetSummaryParams) (*mcp.CallToolResult, any, error) {
	// Extract principal from context
	principal := shared.MustPrincipalFromContext(ctx)

	// Apply defaults and validate (in case called directly from tests)
	err := h.applyDefaultValues(&params)
	if err != nil {
		return nil, nil, err
	}

	// Parse the date
	date, err := time_utils.ParseDate(params.Date)
	if err != nil {
		shared.LogMCPError("get_summary", err, map[string]any{"date": params.Date})
		return nil, nil, errors.Wrap(err, "invalid date format, expected YYYY-MM-DD")
	}

	// Calculate period boundaries based on the period type and date
	var startDate, endDate time.Time
	switch params.PeriodType {
	case "day":
		startDate = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endDate = startDate.AddDate(0, 0, 1).Add(-time.Nanosecond)
	case "week":
		// Find Monday of the week containing the date
		weekday := int(date.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		startDate = date.AddDate(0, 0, -(weekday - 1))
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		endDate = startDate.AddDate(0, 0, 7).Add(-time.Nanosecond)
	case "month":
		startDate = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
	case "quarter":
		quarter := time_utils.Quarter(*date)
		startMonth := (quarter-1)*3 + 1
		startDate = time.Date(date.Year(), time.Month(startMonth), 1, 0, 0, 0, 0, date.Location())
		endDate = startDate.AddDate(0, 3, 0).Add(-time.Nanosecond)
	case "year":
		startDate = time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location())
		endDate = startDate.AddDate(1, 0, 0).Add(-time.Nanosecond)
	}

	// Create activity filter for the period
	filter := &ActivityFilter{
		Timespan: TimespanCustom,
		start:    startDate,
		end:      endDate,
	}

	// Get time reports using the service
	timeReports, err := h.activityService.TimeReports(ctx, principal, filter, params.PeriodType)
	if err != nil {
		shared.LogMCPError("get_summary", err, map[string]any{
			"filter":      filter,
			"period_type": params.PeriodType,
		})
		return nil, nil, errors.Wrap(err, "failed to retrieve time reports")
	}

	// Calculate total hours
	var totalMinutes int
	for _, report := range timeReports {
		totalMinutes += report.DurationInMinutesTotal
	}
	totalHours := float64(totalMinutes) / 60.0

	// Create response
	response := map[string]any{
		"period_type":   params.PeriodType,
		"date":          params.Date,
		"start_date":    time_utils.FormatDate(startDate),
		"end_date":      time_utils.FormatDate(endDate),
		"total_hours":   totalHours,
		"total_minutes": totalMinutes,
		"formatted":     time_utils.FormatMinutesAsDuration(float64(totalMinutes)),
	}

	var resultText string
	if totalMinutes == 0 {
		resultText = fmt.Sprintf("No time tracked for %s period containing %s", params.PeriodType, params.Date)
	} else {
		resultText = fmt.Sprintf("Total time for %s period containing %s: %s (%.2f hours)",
			params.PeriodType, params.Date, time_utils.FormatMinutesAsDuration(float64(totalMinutes)), totalHours)
	}

	shared.LogMCPToolCall("get_summary", map[string]any{"params": params}, true)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: resultText,
			},
			&mcp.TextContent{
				Text: fmt.Sprintf("Summary details: %s", h.formatSummaryJSON(response)),
			},
		},
	}, response, nil
}

// handleGetHoursByProject handles the get_hours_by_project MCP tool call
func (h *ActivityMCPHandlers) handleGetHoursByProject(ctx context.Context, req *mcp.CallToolRequest, params GetHoursByProjectParams) (*mcp.CallToolResult, any, error) {
	// Extract principal from context
	principal := shared.MustPrincipalFromContext(ctx)

	// Apply defaults and validate (in case called directly from tests)
	err := h.applyDefaultValues(&params)
	if err != nil {
		return nil, nil, err
	}

	err = h.validateBusinessRules(&params)
	if err != nil {
		return nil, nil, h.createBusinessLogicError(err.Error())
	}

	// Parse dates
	fromDate, err := time_utils.ParseDate(params.FromDate)
	if err != nil {
		shared.LogMCPError("get_hours_by_project", err, map[string]any{"from_date": params.FromDate})
		return nil, nil, errors.Wrap(err, "invalid from_date format, expected YYYY-MM-DD")
	}

	toDate, err := time_utils.ParseDate(params.ToDate)
	if err != nil {
		shared.LogMCPError("get_hours_by_project", err, map[string]any{"to_date": params.ToDate})
		return nil, nil, errors.Wrap(err, "invalid to_date format, expected YYYY-MM-DD")
	}

	// Create activity filter for the date range
	filter := &ActivityFilter{
		Timespan: TimespanCustom,
		start:    *fromDate,
		end:      toDate.AddDate(0, 0, 1).Add(-time.Nanosecond), // Include the entire end date
	}

	// Get project reports using the service
	projectReports, err := h.activityService.ProjectReports(ctx, principal, filter)
	if err != nil {
		shared.LogMCPError("get_hours_by_project", err, map[string]any{
			"filter": filter,
		})
		return nil, nil, errors.Wrap(err, "failed to retrieve project reports")
	}

	// Convert to response format
	var projects []map[string]any
	var totalMinutes int
	for _, report := range projectReports {
		totalMinutes += report.DurationInMinutesTotal
		projectData := map[string]any{
			"project_id":    report.ProjectID.String(),
			"project_title": report.ProjectTitle,
			"total_hours":   float64(report.DurationInMinutesTotal) / 60.0,
			"total_minutes": report.DurationInMinutesTotal,
			"formatted":     report.DurationFormatted(),
		}
		projects = append(projects, projectData)
	}

	// Create response
	response := map[string]any{
		"from_date":     params.FromDate,
		"to_date":       params.ToDate,
		"projects":      projects,
		"total_hours":   float64(totalMinutes) / 60.0,
		"total_minutes": totalMinutes,
		"formatted":     time_utils.FormatMinutesAsDuration(float64(totalMinutes)),
	}

	var resultText string
	if len(projects) == 0 {
		resultText = fmt.Sprintf("No time tracked for any projects between %s and %s", params.FromDate, params.ToDate)
	} else {
		resultText = fmt.Sprintf("Found time tracked across %d projects between %s and %s. Total: %s (%.2f hours)",
			len(projects), params.FromDate, params.ToDate,
			time_utils.FormatMinutesAsDuration(float64(totalMinutes)), float64(totalMinutes)/60.0)
	}

	shared.LogMCPToolCall("get_hours_by_project", map[string]any{"params": params}, true)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: resultText,
			},
			&mcp.TextContent{
				Text: fmt.Sprintf("Project breakdown: %s", h.formatProjectReportJSON(response)),
			},
		},
	}, response, nil
}

// handleListProjects handles the list_projects MCP tool call
func (h *ActivityMCPHandlers) handleListProjects(ctx context.Context, req *mcp.CallToolRequest, params ListProjectsParams) (*mcp.CallToolResult, any, error) {
	// Extract principal from context
	principal := shared.MustPrincipalFromContext(ctx)

	// Use a large page size to get all projects (or implement pagination later)
	pageParams := &paged.PageParams{
		Page: 0,
		Size: 1000, // Large enough to get all projects
	}

	// Get projects using the service
	projectsPaged, err := h.projectService.ReadProjects(ctx, principal, pageParams)
	if err != nil {
		shared.LogMCPError("list_projects", err, map[string]any{
			"pageParams": pageParams,
		})
		return nil, nil, errors.Wrap(err, "failed to retrieve projects")
	}

	// Convert projects to MCP response format with consistent ordering (alphabetical by title)
	var responseProjects []map[string]any
	for _, project := range projectsPaged.Projects {
		projectData := map[string]any{
			"id":          project.ID.String(),
			"title":       project.Title,
			"description": project.Description,
			"active":      project.Active,
		}
		responseProjects = append(responseProjects, projectData)
	}

	// Sort projects alphabetically by title for consistent ordering
	// Using a simple bubble sort for simplicity since project count is typically small
	for i := 0; i < len(responseProjects); i++ {
		for j := i + 1; j < len(responseProjects); j++ {
			if responseProjects[i]["title"].(string) > responseProjects[j]["title"].(string) {
				responseProjects[i], responseProjects[j] = responseProjects[j], responseProjects[i]
			}
		}
	}

	response := map[string]any{
		"projects": responseProjects,
		"total":    len(responseProjects),
	}

	var resultText string
	if len(responseProjects) == 0 {
		resultText = "No projects found in the system"
	} else {
		resultText = fmt.Sprintf("Found %d projects", len(responseProjects))
	}

	shared.LogMCPToolCall("list_projects", map[string]any{"params": params}, true)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: resultText,
			},
			&mcp.TextContent{
				Text: fmt.Sprintf("Projects: %s", h.formatProjectsJSON(responseProjects)),
			},
		},
	}, response, nil
}

// Helper methods

// mapActivityToMCPResponse converts an Activity to MCP response format
func (h *ActivityMCPHandlers) mapActivityToMCPResponse(activity *Activity) map[string]any {
	// Prepare tags
	tagNames := make([]string, len(activity.Tags))
	for i, tag := range activity.Tags {
		tagNames[i] = tag.Name
	}

	return map[string]any{
		"id":          activity.ID.String(),
		"start":       time_utils.FormatDateTime(activity.Start),
		"end":         time_utils.FormatDateTime(activity.End),
		"description": activity.Description,
		"project_id":  activity.ProjectID.String(),
		"username":    activity.Username,
		"tags":        tagNames,
		"duration": map[string]any{
			"hours":     activity.DurationHours(),
			"minutes":   activity.DurationMinutes(),
			"decimal":   activity.DurationDecimal(),
			"formatted": activity.DurationFormatted(),
		},
	}
}

// formatActivityJSON formats activity response as JSON string for display
func (h *ActivityMCPHandlers) formatActivityJSON(response map[string]any) string {
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("Failed to format activity JSON: %v", err)
		return fmt.Sprintf("%+v", response)
	}
	return string(jsonBytes)
}

// formatEntriesJSON formats entries list as JSON string for display
func (h *ActivityMCPHandlers) formatEntriesJSON(entries []map[string]any) string {
	jsonBytes, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		log.Printf("Failed to format entries JSON: %v", err)
		return fmt.Sprintf("%+v", entries)
	}
	return string(jsonBytes)
}

// formatSummaryJSON formats summary response as JSON string for display
func (h *ActivityMCPHandlers) formatSummaryJSON(response map[string]any) string {
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("Failed to format summary JSON: %v", err)
		return fmt.Sprintf("%+v", response)
	}
	return string(jsonBytes)
}

// formatProjectReportJSON formats project report response as JSON string for display
func (h *ActivityMCPHandlers) formatProjectReportJSON(response map[string]any) string {
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("Failed to format project report JSON: %v", err)
		return fmt.Sprintf("%+v", response)
	}
	return string(jsonBytes)
}

// formatProjectsJSON formats projects list as JSON string for display
func (h *ActivityMCPHandlers) formatProjectsJSON(projects []map[string]any) string {
	jsonBytes, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		log.Printf("Failed to format projects JSON: %v", err)
		return fmt.Sprintf("%+v", projects)
	}
	return string(jsonBytes)
}

// parseArguments converts ToolArguments to typed parameters with validation and default value handling
func (h *ActivityMCPHandlers) parseArguments(arguments shared.ToolArguments, target any) error {
	// Convert arguments map to JSON and then unmarshal to target struct
	jsonBytes, err := json.Marshal(arguments)
	if err != nil {
		return h.createValidationError("failed to marshal arguments", err)
	}

	err = json.Unmarshal(jsonBytes, target)
	if err != nil {
		return h.createValidationError("failed to unmarshal arguments to target type", err)
	}

	// Apply default values before validation
	err = h.applyDefaultValues(target)
	if err != nil {
		return h.createValidationError("failed to apply default values", err)
	}

	// Validate the parsed parameters with defaults applied
	err = h.validator.Struct(target)
	if err != nil {
		return h.createValidationError("parameter validation failed", err)
	}

	// Apply business rule validation after basic validation
	err = h.validateBusinessRules(target)
	if err != nil {
		return h.createBusinessLogicError(err.Error())
	}

	return nil
}

// applyDefaultValues applies default values for missing time parameters
func (h *ActivityMCPHandlers) applyDefaultValues(target any) error {
	switch params := target.(type) {
	case *CreateEntryParams:
		now := time.Now()
		if params.Start == "" {
			params.Start = time_utils.FormatDateTime(now)
		}
		if params.End == "" {
			// Default end time to 1 hour after start time to avoid business rule violation
			endTime := now.Add(time.Hour)
			if params.Start != "" {
				if startTime, err := time_utils.ParseDateTime(params.Start); err == nil {
					endTime = startTime.Add(time.Hour)
				}
			}
			params.End = time_utils.FormatDateTime(endTime)
		}

	case *ListEntriesParams:
		if params.FromDate == "" || params.ToDate == "" {
			now := time.Now()
			if params.FromDate == "" {
				firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				params.FromDate = time_utils.FormatDate(firstDayOfMonth)
			}
			if params.ToDate == "" {
				params.ToDate = time_utils.FormatDate(now)
			}
		}

	case *GetSummaryParams:
		if params.Date == "" {
			params.Date = time_utils.FormatDate(time.Now())
		}

	case *GetHoursByProjectParams:
		if params.FromDate == "" || params.ToDate == "" {
			now := time.Now()
			if params.FromDate == "" {
				firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				params.FromDate = time_utils.FormatDate(firstDayOfMonth)
			}
			if params.ToDate == "" {
				params.ToDate = time_utils.FormatDate(now)
			}
		}
	}

	return nil
}

// validateBusinessRules validates business logic rules after applying defaults
func (h *ActivityMCPHandlers) validateBusinessRules(target any) error {
	switch params := target.(type) {
	case *CreateEntryParams:
		return h.validateTimeRange(params.Start, params.End)

	case *UpdateEntryParams:
		// For updates, only validate if both times are provided
		if params.Start != "" && params.End != "" {
			return h.validateTimeRange(params.Start, params.End)
		}

	case *ListEntriesParams:
		return h.validateDateRange(params.FromDate, params.ToDate)

	case *GetHoursByProjectParams:
		return h.validateDateRange(params.FromDate, params.ToDate)
	}

	return nil
}

// validateTimeRange validates that end time is after start time
func (h *ActivityMCPHandlers) validateTimeRange(startStr, endStr string) error {
	if startStr == "" || endStr == "" {
		return nil // Skip validation if either is empty
	}

	startTime, err := time_utils.ParseDateTime(startStr)
	if err != nil {
		return fmt.Errorf("invalid start time format: %v", err)
	}

	endTime, err := time_utils.ParseDateTime(endStr)
	if err != nil {
		return fmt.Errorf("invalid end time format: %v", err)
	}

	if endTime.Before(*startTime) || endTime.Equal(*startTime) {
		return errors.New("end time must be after start time")
	}

	return nil
}

// validateDateRange validates that to_date is on or after from_date
func (h *ActivityMCPHandlers) validateDateRange(fromDateStr, toDateStr string) error {
	if fromDateStr == "" || toDateStr == "" {
		return nil // Skip validation if either is empty
	}

	fromDate, err := time_utils.ParseDate(fromDateStr)
	if err != nil {
		return fmt.Errorf("invalid from_date format: %v", err)
	}

	toDate, err := time_utils.ParseDate(toDateStr)
	if err != nil {
		return fmt.Errorf("invalid to_date format: %v", err)
	}

	if toDate.Before(*fromDate) {
		return errors.New("to_date must be on or after from_date")
	}

	return nil
}

// createValidationError creates a validation error with proper MCP formatting
func (h *ActivityMCPHandlers) createValidationError(message string, err error) error {
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		details := h.formatValidationErrors(validationErr)
		return fmt.Errorf("validation error: %s", details)
	}
	return fmt.Errorf("validation error: %s - %v", message, err)
}

// createBusinessLogicError creates a business logic error
func (h *ActivityMCPHandlers) createBusinessLogicError(message string) error {
	return fmt.Errorf("business rule violation: %s", message)
}

// formatValidationErrors formats validator.ValidationErrors into a readable string
func (h *ActivityMCPHandlers) formatValidationErrors(validationErr validator.ValidationErrors) string {
	var errorMessages []string
	for _, err := range validationErr {
		switch err.Tag() {
		case "required":
			errorMessages = append(errorMessages, fmt.Sprintf("field '%s' is required", err.Field()))
		case "uuid":
			errorMessages = append(errorMessages, fmt.Sprintf("field '%s' must be a valid UUID", err.Field()))
		case "max":
			errorMessages = append(errorMessages, fmt.Sprintf("field '%s' must be at most %s characters", err.Field(), err.Param()))
		case "min":
			errorMessages = append(errorMessages, fmt.Sprintf("field '%s' must be at least %s characters", err.Field(), err.Param()))
		case "oneof":
			errorMessages = append(errorMessages, fmt.Sprintf("field '%s' must be one of: %s", err.Field(), err.Param()))
		case "datetime":
			errorMessages = append(errorMessages, fmt.Sprintf("field '%s' must be a valid ISO 8601 datetime (YYYY-MM-DDTHH:MM:SS)", err.Field()))
		case "date":
			errorMessages = append(errorMessages, fmt.Sprintf("field '%s' must be a valid date (YYYY-MM-DD)", err.Field()))
		case "dive":
			errorMessages = append(errorMessages, fmt.Sprintf("field '%s' contains invalid items", err.Field()))
		default:
			errorMessages = append(errorMessages, fmt.Sprintf("field '%s' validation failed: %s", err.Field(), err.Tag()))
		}
	}
	return strings.Join(errorMessages, "; ")
}

// Custom validation functions

// validateDateTime validates ISO 8601 datetime format
func validateDateTime(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}

	_, err := time_utils.ParseDateTime(value)
	return err == nil
}

// validateDate validates YYYY-MM-DD date format
func validateDate(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}

	_, err := time_utils.ParseDate(value)
	return err == nil
}
