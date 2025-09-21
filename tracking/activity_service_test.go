package tracking

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestTimeReportsByDay(t *testing.T) {
	// Arrange
	is := is.New(t)

	activityRepository := NewInMemActivityRepository()
	a := &ActitivityService{
		activityRepository: activityRepository,
	}

	start1, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end1, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	start2, _ := time.Parse(time.RFC3339, "2021-01-08T10:00:00.000Z")
	end2, _ := time.Parse(time.RFC3339, "2021-01-08T11:00:00.000Z")

	activityRepository.activities = []*Activity{
		{
			Start: start1,
			End:   end1,
		},
		{
			Start: start2,
			End:   end2,
		},
	}

	principal := &shared.Principal{}
	filter := &ActivityFilter{}

	// Act
	timeReports, err := a.TimeReports(context.Background(), principal, filter, "day")

	// Assert
	is.NoErr(err)
	is.Equal(len(timeReports), 2)

	day1 := timeReports[0]
	is.Equal(day1.Year, 2021)
	is.Equal(day1.Day, 1)
	is.Equal(day1.DurationInMinutesTotal, 60)

	day2 := timeReports[1]
	is.Equal(day2.Year, 2021)
	is.Equal(day2.Day, 8)
	is.Equal(day2.DurationInMinutesTotal, 60)
}

func TestTimeReportsByWeek(t *testing.T) {
	// Arrange
	is := is.New(t)

	activityRepository := NewInMemActivityRepository()
	a := &ActitivityService{
		activityRepository: activityRepository,
	}

	start1, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end1, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	start2, _ := time.Parse(time.RFC3339, "2021-01-08T10:00:00.000Z")
	end2, _ := time.Parse(time.RFC3339, "2021-01-08T11:00:00.000Z")

	activityRepository.activities = []*Activity{
		{
			Start: start1,
			End:   end1,
		},
		{
			Start: start2,
			End:   end2,
		},
	}

	principal := &shared.Principal{}
	filter := &ActivityFilter{}

	// Act
	timeReports, err := a.TimeReports(context.Background(), principal, filter, "week")

	// Assert
	is.NoErr(err)
	is.Equal(len(timeReports), 2)

	week1 := timeReports[0]
	is.Equal(week1.Year, 2021)
	is.Equal(week1.Week, 53)
	is.Equal(week1.DurationInMinutesTotal, 60)

	week2 := timeReports[1]
	is.Equal(week2.Year, 2021)
	is.Equal(week2.Week, 1)
	is.Equal(week2.DurationInMinutesTotal, 60)
}

func TestTimeReportsByMonth(t *testing.T) {
	// Arrange
	is := is.New(t)

	activityRepository := NewInMemActivityRepository()
	a := &ActitivityService{
		activityRepository: activityRepository,
	}

	start1, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end1, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	start2, _ := time.Parse(time.RFC3339, "2021-02-01T10:00:00.000Z")
	end2, _ := time.Parse(time.RFC3339, "2021-02-01T11:00:00.000Z")

	activityRepository.activities = []*Activity{
		{
			Start: start1,
			End:   end1,
		},
		{
			Start: start2,
			End:   end2,
		},
	}

	principal := &shared.Principal{}
	filter := &ActivityFilter{}

	// Act
	timeReports, err := a.TimeReports(context.Background(), principal, filter, "month")

	// Assert
	is.NoErr(err)
	is.Equal(len(timeReports), 2)

	monthJan := timeReports[0]
	is.Equal(monthJan.Year, 2021)
	is.Equal(monthJan.Month, 1)
	is.Equal(monthJan.DurationInMinutesTotal, 60)

	monthFeb := timeReports[1]
	is.Equal(monthFeb.Year, 2021)
	is.Equal(monthFeb.Month, 2)
	is.Equal(monthFeb.DurationInMinutesTotal, 60)
}

func TestTimeReportsByQuarter(t *testing.T) {
	// Arrange
	is := is.New(t)

	activityRepository := NewInMemActivityRepository()
	a := &ActitivityService{
		activityRepository: activityRepository,
	}

	start1, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end1, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	start2, _ := time.Parse(time.RFC3339, "2021-04-01T10:00:00.000Z")
	end2, _ := time.Parse(time.RFC3339, "2021-04-01T11:00:00.000Z")

	activityRepository.activities = []*Activity{
		{
			Start: start1,
			End:   end1,
		},
		{
			Start: start2,
			End:   end2,
		},
	}

	principal := &shared.Principal{}
	filter := &ActivityFilter{}

	// Act
	timeReports, err := a.TimeReports(context.Background(), principal, filter, "quarter")

	// Assert
	is.NoErr(err)
	is.Equal(len(timeReports), 2)

	q1 := timeReports[0]
	is.Equal(q1.Year, 2021)
	is.Equal(q1.Quarter, 1)
	is.Equal(q1.DurationInMinutesTotal, 60)

	q2 := timeReports[1]
	is.Equal(q2.Year, 2021)
	is.Equal(q2.Quarter, 2)
	is.Equal(q2.DurationInMinutesTotal, 60)
}

func TestProjectReports(t *testing.T) {
	// Arrange
	is := is.New(t)

	activityRepository := NewInMemActivityRepository()
	a := &ActitivityService{
		activityRepository: activityRepository,
	}

	projectId1 := uuid.New()
	start1, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end1, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	projectId2 := uuid.New()
	start2, _ := time.Parse(time.RFC3339, "2021-01-08T10:00:00.000Z")
	end2, _ := time.Parse(time.RFC3339, "2021-01-08T11:00:00.000Z")

	activityRepository.activities = []*Activity{
		{
			ProjectID: projectId1,
			Start:     start1,
			End:       end1,
		},
		{
			ProjectID: projectId2,
			Start:     start2,
			End:       end2,
		},
	}

	principal := &shared.Principal{}
	filter := &ActivityFilter{}

	// Act
	projectReports, err := a.ProjectReports(context.Background(), principal, filter)

	// Assert
	is.NoErr(err)
	is.Equal(len(projectReports), 2)

	reportItem1 := projectReports[0]
	is.Equal(reportItem1.ProjectID, projectId1)
	is.Equal(reportItem1.ProjectTitle, "My Project")
	is.Equal(reportItem1.DurationInMinutesTotal, 60)

	reportItem2 := projectReports[1]
	is.Equal(reportItem2.ProjectID, projectId2)
	is.Equal(reportItem2.ProjectTitle, "My Project")
	is.Equal(reportItem2.DurationInMinutesTotal, 60)
}

func TestWriteAsCSV(t *testing.T) {
	is := is.New(t)

	a := &ActitivityService{}

	start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

	activity := &Activity{
		Start:     start,
		End:       end,
		ProjectID: uuid.New(),
	}
	activities := []*Activity{activity}

	project := &Project{
		ID:    activity.ProjectID,
		Title: "My Project",
	}
	projects := []*Project{project}

	var buffer bytes.Buffer

	err := a.WriteAsCSV(activities, projects, &buffer)

	is.NoErr(err)
	csv := buffer.String()

	is.True(strings.Contains(csv, "Date"))
	is.True(strings.Contains(csv, "My Project"))
	is.True(strings.Contains(csv, "11:00"))
	is.True(strings.Contains(csv, "11:30"))
}

func TestWriteAsExcel(t *testing.T) {
	is := is.New(t)

	a := &ActitivityService{}

	start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

	activity := &Activity{
		Start:     start,
		End:       end,
		ProjectID: uuid.New(),
	}
	activities := []*Activity{activity}

	project := &Project{
		ID:    activity.ProjectID,
		Title: "My Project",
	}
	projects := []*Project{project}

	var buffer bytes.Buffer

	err := a.WriteAsExcel(activities, projects, &buffer)

	is.NoErr(err)
}

func TestActivityService_TagIntegration(t *testing.T) {
	// Arrange
	is := is.New(t)

	tagRepository := NewInMemTagRepository()
	tagService := NewTagService(tagRepository)
	activityRepository := NewInMemActivityRepository()
	
	a := &ActitivityService{
		activityRepository: activityRepository,
		tagRepository:      tagRepository,
		tagService:         tagService,
	}

	// Test ParseTagsFromString
	tags := a.ParseTagsFromString("Meeting, Development, bug-fix")
	is.Equal(len(tags), 3)
	is.Equal(tags[0], "meeting")
	is.Equal(tags[1], "development")
	is.Equal(tags[2], "bug-fix")

	// Test ValidateTags with valid tags
	err := a.ValidateTags([]string{"meeting", "development", "bug-fix"})
	is.NoErr(err)

	// Test ValidateTags with too many tags
	tooManyTags := make([]string, 11)
	for i := 0; i < 11; i++ {
		tooManyTags[i] = "tag" + string(rune(i))
	}
	err = a.ValidateTags(tooManyTags)
	is.True(err != nil)
	is.Equal(err, ErrTooManyTags)

	// Test GetTagsForAutocomplete
	principal := &shared.Principal{
		OrganizationID: uuid.New(),
	}
	
	// Add some tags to the repository
	tagRepository.tags = []*Tag{
		{ID: uuid.New(), Name: "meeting", OrganizationID: principal.OrganizationID},
		{ID: uuid.New(), Name: "development", OrganizationID: principal.OrganizationID},
		{ID: uuid.New(), Name: "testing", OrganizationID: principal.OrganizationID},
	}

	autocompleteResults, err := a.GetTagsForAutocomplete(context.Background(), principal, "meet")
	is.NoErr(err)
	is.Equal(len(autocompleteResults), 1)
	is.Equal(autocompleteResults[0].Name, "meeting")
}

func TestActivityService_CreateActivityWithTags(t *testing.T) {
	// Arrange
	is := is.New(t)

	tagRepository := NewInMemTagRepository()
	tagService := NewTagService(tagRepository)
	activityRepository := NewInMemActivityRepository()
	repositoryTxer := &shared.InMemRepositoryTxer{}
	
	a := &ActitivityService{
		repositoryTxer:     repositoryTxer,
		activityRepository: activityRepository,
		tagRepository:      tagRepository,
		tagService:         tagService,
	}

	principal := &shared.Principal{
		OrganizationID: uuid.New(),
		Username:       "testuser",
	}

	start, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	activity := &Activity{
		Start:       start,
		End:         end,
		Description: "Test activity",
		ProjectID:   uuid.New(),
		Tags:        []string{"Meeting", "DEVELOPMENT", "bug-fix"},
	}

	// Act
	createdActivity, err := a.CreateActivity(context.Background(), principal, activity)

	// Assert
	is.NoErr(err)
	is.True(createdActivity != nil)
	is.Equal(len(createdActivity.Tags), 3)
	// Tags should be normalized to lowercase
	is.Equal(createdActivity.Tags[0], "meeting")
	is.Equal(createdActivity.Tags[1], "development")
	is.Equal(createdActivity.Tags[2], "bug-fix")
	is.Equal(createdActivity.OrganizationID, principal.OrganizationID)
	is.Equal(createdActivity.Username, principal.Username)
}

func TestActivityService_UpdateActivityWithTags(t *testing.T) {
	// Arrange
	is := is.New(t)

	tagRepository := NewInMemTagRepository()
	tagService := NewTagService(tagRepository)
	activityRepository := NewInMemActivityRepository()
	repositoryTxer := &shared.InMemRepositoryTxer{}
	
	a := &ActitivityService{
		repositoryTxer:     repositoryTxer,
		activityRepository: activityRepository,
		tagRepository:      tagRepository,
		tagService:         tagService,
	}

	principal := &shared.Principal{
		OrganizationID: uuid.New(),
		Username:       "testuser",
		Roles:          []string{"ROLE_ADMIN"},
	}

	activityID := uuid.New()
	start, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	// First create an activity to update
	existingActivity := &Activity{
		ID:             activityID,
		Start:          start,
		End:            end,
		Description:    "Original activity",
		ProjectID:      uuid.New(),
		OrganizationID: principal.OrganizationID,
		Username:       principal.Username,
		Tags:           []string{"original"},
	}
	
	// Add the activity to the repository
	activityRepository.activities = []*Activity{existingActivity}

	activity := &Activity{
		ID:          activityID,
		Start:       start,
		End:         end,
		Description: "Updated activity",
		ProjectID:   existingActivity.ProjectID,
		Tags:        []string{"UPDATED", "tags"},
	}

	// Act
	updatedActivity, err := a.UpdateActivity(context.Background(), principal, activity)

	// Assert
	is.NoErr(err)
	is.True(updatedActivity != nil)
	is.Equal(len(updatedActivity.Tags), 2)
	// Tags should be normalized to lowercase
	is.Equal(updatedActivity.Tags[0], "updated")
	is.Equal(updatedActivity.Tags[1], "tags")
}

func TestActivityFilter_WithTags(t *testing.T) {
	// Arrange
	is := is.New(t)

	start, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	filter := &ActivityFilter{
		Timespan: TimespanCustom,
		start:    start,
		end:      end,
	}

	// Act
	filterWithTags := filter.WithTags([]string{"meeting", "development"})

	// Assert
	is.Equal(len(filterWithTags.Tags()), 2)
	is.Equal(filterWithTags.Tags()[0], "meeting")
	is.Equal(filterWithTags.Tags()[1], "development")
	is.Equal(filterWithTags.Timespan, TimespanCustom)
	is.Equal(filterWithTags.Start(), start)
	is.Equal(filterWithTags.End(), end)
}

func TestToFilter_WithTags(t *testing.T) {
	// Arrange
	is := is.New(t)

	principal := &shared.Principal{
		OrganizationID: uuid.New(),
		Username:       "testuser",
	}

	start, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	filter := &ActivityFilter{
		Timespan: TimespanWeek,
		start:    start,
		end:      end,
		tags:     []string{"meeting", "development"},
	}

	// Act
	activitiesFilter := toFilter(principal, filter)

	// Assert
	is.Equal(len(activitiesFilter.Tags), 2)
	is.Equal(activitiesFilter.Tags[0], "meeting")
	is.Equal(activitiesFilter.Tags[1], "development")
	is.Equal(activitiesFilter.OrganizationID, principal.OrganizationID)
	is.Equal(activitiesFilter.Username, principal.Username)
}
