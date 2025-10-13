package tracking

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
	validator          *validator.Validate
}

// NewActivityMCPHandlers creates a new ActivityMCPHandlers instance
func NewActivityMCPHandlers(
	activityService *ActitivityService,
	activityRepository ActivityRepository,
	projectRepository ProjectRepository,
) *ActivityMCPHandlers {
	return &ActivityMCPHandlers{
		activityService:    activityService,
		activityRepository: activityRepository,
		projectRepository:  projectRepository,
		validator:          validator.New(),
	}
}

// RegisterMCPTools registers all activity-related MCP tools
func (h *ActivityMCPHandlers) RegisterMCPTools(server *mcp.Server) {
	// Register create_entry tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_entry",
		Description: "Create a new time tracking entry with start/end times, description, and project association",
	}, h.handleCreateEntry)

	// Register get_entry tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_entry",
		Description: "Retrieve a specific time entry by its ID",
	}, h.handleGetEntry)

	// Register update_entry tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_entry",
		Description: "Update an existing time entry with new values",
	}, h.handleUpdateEntry)

	// Register delete_entry tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_entry",
		Description: "Delete a time entry by its ID",
	}, h.handleDeleteEntry)

	// Register list_entries tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_entries",
		Description: "List time entries with optional filtering by date range and project",
	}, h.handleListEntries)
}

// MCP parameter structures for tool calls

// CreateEntryParams represents parameters for create_entry tool
type CreateEntryParams struct {
	Start       string   `json:"start" validate:"required" jsonschema:"description=Start time in ISO 8601 format"`
	End         string   `json:"end" validate:"required" jsonschema:"description=End time in ISO 8601 format"`
	Description string   `json:"description" validate:"max=500" jsonschema:"description=Description of the work performed"`
	ProjectID   string   `json:"project_id" validate:"required,uuid" jsonschema:"description=UUID of the project"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description=Optional array of tag names"`
}

// GetEntryParams represents parameters for get_entry tool
type GetEntryParams struct {
	EntryID string `json:"entry_id" validate:"required,uuid" jsonschema:"description=UUID of the time entry to retrieve"`
}

// UpdateEntryParams represents parameters for update_entry tool
type UpdateEntryParams struct {
	EntryID     string   `json:"entry_id" validate:"required,uuid" jsonschema:"description=UUID of the time entry to update"`
	Start       string   `json:"start,omitempty" jsonschema:"description=Start time in ISO 8601 format (optional)"`
	End         string   `json:"end,omitempty" jsonschema:"description=End time in ISO 8601 format (optional)"`
	Description string   `json:"description" validate:"max=500" jsonschema:"description=Description of the work performed (optional)"`
	ProjectID   string   `json:"project_id,omitempty" jsonschema:"description=UUID of the project (optional)"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description=Optional array of tag names"`
}

// DeleteEntryParams represents parameters for delete_entry tool
type DeleteEntryParams struct {
	EntryID string `json:"entry_id" validate:"required,uuid" jsonschema:"description=UUID of the time entry to delete"`
}

// ListEntriesParams represents parameters for list_entries tool
type ListEntriesParams struct {
	FromDate  string `json:"from_date,omitempty" jsonschema:"description=Start date filter in YYYY-MM-DD format (optional)"`
	ToDate    string `json:"to_date,omitempty" jsonschema:"description=End date filter in YYYY-MM-DD format (optional)"`
	ProjectID string `json:"project_id,omitempty" validate:"omitempty,uuid" jsonschema:"description=UUID of the project to filter by (optional)"`
}

// MCP tool handlers

// handleCreateEntry handles the create_entry MCP tool call
func (h *ActivityMCPHandlers) handleCreateEntry(ctx context.Context, req *mcp.CallToolRequest, params CreateEntryParams) (*mcp.CallToolResult, any, error) {
	// Extract principal from context
	principal := shared.MustPrincipalFromContext(ctx)

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

	// Validate business rules
	if endTime.Before(*startTime) || endTime.Equal(*startTime) {
		err := errors.New("end time must be after start time")
		shared.LogMCPError("create_entry", err, map[string]any{
			"start": params.Start,
			"end":   params.End,
		})
		return nil, nil, err
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

	// Create activity filter
	filter := &ActivityFilter{
		Timespan: TimespanCustom,
	}

	// Parse and set date filters if provided
	if params.FromDate != "" {
		fromDate, err := time_utils.ParseDate(params.FromDate)
		if err != nil {
			shared.LogMCPError("list_entries", err, map[string]any{"from_date": params.FromDate})
			return nil, nil, errors.Wrap(err, "invalid from_date format, expected YYYY-MM-DD")
		}
		filter.start = *fromDate
	}

	if params.ToDate != "" {
		toDate, err := time_utils.ParseDate(params.ToDate)
		if err != nil {
			shared.LogMCPError("list_entries", err, map[string]any{"to_date": params.ToDate})
			return nil, nil, errors.Wrap(err, "invalid to_date format, expected YYYY-MM-DD")
		}
		filter.end = *toDate
	}

	// Validate date range if both dates are provided
	if !filter.start.IsZero() && !filter.end.IsZero() && filter.end.Before(filter.start) {
		err := errors.New("to_date must be on or after from_date")
		shared.LogMCPError("list_entries", err, map[string]any{
			"from_date": params.FromDate,
			"to_date":   params.ToDate,
		})
		return nil, nil, err
	}

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
		if params.FromDate != "" || params.ToDate != "" {
			if params.FromDate != "" && params.ToDate != "" {
				resultText += fmt.Sprintf(" between %s and %s", params.FromDate, params.ToDate)
			} else if params.FromDate != "" {
				resultText += fmt.Sprintf(" from %s onwards", params.FromDate)
			} else {
				resultText += fmt.Sprintf(" up to %s", params.ToDate)
			}
		}
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
