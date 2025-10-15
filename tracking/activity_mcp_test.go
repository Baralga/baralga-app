package tracking

import (
	"context"
	"testing"
	"time"

	"github.com/baralga/shared"
	time_utils "github.com/baralga/tracking/time"
	"github.com/matryer/is"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestActivityMCPHandlers_DefaultValues(t *testing.T) {
	is := is.New(t)

	// Setup in-memory repositories
	activityRepo := NewInMemActivityRepository()
	projectRepo := NewInMemProjectRepository()
	tagRepo := NewInMemTagRepository()
	tagService := NewTagService(tagRepo)

	// Create services directly (like in existing tests)
	repositoryTxer := shared.NewInMemRepositoryTxer()
	activityService := &ActitivityService{
		repositoryTxer:     repositoryTxer,
		activityRepository: activityRepo,
		tagRepository:      tagRepo,
		tagService:         tagService,
	}
	projectService := &ProjectService{
		repositoryTxer:    repositoryTxer,
		projectRepository: projectRepo,
	}

	// Create MCP handlers
	handlers := NewActivityMCPHandlers(activityService, activityRepo, projectRepo, projectService)

	// Create test organization and user using sample UUIDs
	orgID := shared.OrganizationIDSample
	username := "test@example.com"
	principal := &shared.Principal{
		Username:       username,
		OrganizationID: orgID,
		Roles:          []string{"ROLE_USER"},
	}
	ctx := shared.ToContextWithPrincipal(context.Background(), principal)

	// Create a test project using sample UUID
	project := &Project{
		ID:             shared.ProjectIDSample,
		Title:          "Test Project",
		Description:    "Test project for MCP testing",
		Active:         true,
		OrganizationID: orgID,
	}
	_, err := projectRepo.InsertProject(ctx, project)
	is.NoErr(err)

	t.Run("CreateEntry with default times", func(t *testing.T) {
		is := is.New(t)

		// Test create_entry with missing start and end times
		params := CreateEntryParams{
			Description: "Test entry with default times",
			ProjectID:   project.ID.String(),
		}

		result, response, err := handlers.handleCreateEntry(ctx, &mcp.CallToolRequest{}, params)
		is.NoErr(err)
		is.True(result != nil)
		is.True(response != nil)

		// Verify that the response contains the created entry
		responseMap := response.(map[string]any)
		is.True(responseMap["id"] != nil)
		is.True(responseMap["start"] != nil)
		is.True(responseMap["end"] != nil)
		is.Equal(responseMap["description"], "Test entry with default times")
		is.Equal(responseMap["project_id"], project.ID.String())
	})

	t.Run("ListEntries with default date range", func(t *testing.T) {
		is := is.New(t)

		// Test list_entries with no date filters (should use current month defaults)
		params := ListEntriesParams{}

		result, response, err := handlers.handleListEntries(ctx, &mcp.CallToolRequest{}, params)
		is.NoErr(err)
		is.True(result != nil)
		is.True(response != nil)

		// Verify that the response contains filters with default dates
		responseMap := response.(map[string]any)
		filters := responseMap["filters"].(map[string]any)

		// Should have applied default date range (first day of current month to current date)
		now := time.Now()
		expectedFromDate := time_utils.FormatDate(time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()))
		expectedToDate := time_utils.FormatDate(now)

		is.Equal(filters["from_date"], expectedFromDate)
		is.Equal(filters["to_date"], expectedToDate)
	})

	t.Run("GetSummary with default date", func(t *testing.T) {
		is := is.New(t)

		// Test get_summary with missing date (should use current date default)
		params := GetSummaryParams{
			PeriodType: "day",
		}

		result, response, err := handlers.handleGetSummary(ctx, &mcp.CallToolRequest{}, params)
		is.NoErr(err)
		is.True(result != nil)
		is.True(response != nil)

		// Verify that the response contains the default date
		responseMap := response.(map[string]any)
		expectedDate := time_utils.FormatDate(time.Now())
		is.Equal(responseMap["date"], expectedDate)
	})

	t.Run("GetHoursByProject with default date range", func(t *testing.T) {
		is := is.New(t)

		// Test get_hours_by_project with no date filters (should use current month defaults)
		params := GetHoursByProjectParams{}

		result, response, err := handlers.handleGetHoursByProject(ctx, &mcp.CallToolRequest{}, params)
		is.NoErr(err)
		is.True(result != nil)
		is.True(response != nil)

		// Verify that the response contains default date range
		responseMap := response.(map[string]any)

		// Should have applied default date range (first day of current month to current date)
		now := time.Now()
		expectedFromDate := time_utils.FormatDate(time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()))
		expectedToDate := time_utils.FormatDate(time.Now())

		is.Equal(responseMap["from_date"], expectedFromDate)
		is.Equal(responseMap["to_date"], expectedToDate)
	})
}

func TestActivityMCPHandlers_ValidationAfterDefaults(t *testing.T) {
	is := is.New(t)

	// Setup in-memory repositories
	activityRepo := NewInMemActivityRepository()
	projectRepo := NewInMemProjectRepository()
	tagRepo := NewInMemTagRepository()
	tagService := NewTagService(tagRepo)

	// Create services directly (like in existing tests)
	repositoryTxer := shared.NewInMemRepositoryTxer()
	activityService := &ActitivityService{
		repositoryTxer:     repositoryTxer,
		activityRepository: activityRepo,
		tagRepository:      tagRepo,
		tagService:         tagService,
	}
	projectService := &ProjectService{
		repositoryTxer:    repositoryTxer,
		projectRepository: projectRepo,
	}

	// Create MCP handlers
	handlers := NewActivityMCPHandlers(activityService, activityRepo, projectRepo, projectService)

	// Create test organization and user using sample UUIDs
	orgID := shared.OrganizationIDSample
	username := "test@example.com"
	principal := &shared.Principal{
		Username:       username,
		OrganizationID: orgID,
		Roles:          []string{"ROLE_USER"},
	}
	ctx := shared.ToContextWithPrincipal(context.Background(), principal)

	// Create a test project using sample UUID
	project := &Project{
		ID:             shared.ProjectIDSample,
		Title:          "Test Project",
		Description:    "Test project for MCP testing",
		Active:         true,
		OrganizationID: orgID,
	}
	_, err := projectRepo.InsertProject(ctx, project)
	is.NoErr(err)

	t.Run("CreateEntry with invalid time combination after defaults", func(t *testing.T) {
		is := is.New(t)

		// Test create_entry with end time before start time
		futureTime := time.Now().Add(2 * time.Hour)
		pastTime := time.Now().Add(-1 * time.Hour)

		params := CreateEntryParams{
			Start:       time_utils.FormatDateTime(futureTime),
			End:         time_utils.FormatDateTime(pastTime),
			Description: "Test entry with invalid times",
			ProjectID:   project.ID.String(),
		}

		_, _, err := handlers.handleCreateEntry(ctx, &mcp.CallToolRequest{}, params)
		is.True(err != nil) // Should return an error
		is.True(err.Error() == "end time must be after start time")
	})
}
