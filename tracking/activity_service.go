package tracking

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/baralga/shared"
	"github.com/baralga/shared/paged"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type ActitivityService struct {
	repositoryTxer     shared.RepositoryTxer
	activityRepository ActivityRepository
	tagRepository      TagRepository
	tagService         *TagService
}

func NewActitivityService(repositoryTxer shared.RepositoryTxer, activityRepository ActivityRepository, tagRepository TagRepository, tagService *TagService) *ActitivityService {
	return &ActitivityService{
		repositoryTxer:     repositoryTxer,
		activityRepository: activityRepository,
		tagRepository:      tagRepository,
		tagService:         tagService,
	}
}

// ReadActivitiesWithProjects reads activities with their associated projects
func (a *ActitivityService) ReadActivitiesWithProjects(ctx context.Context, principal *shared.Principal, filter *ActivityFilter, pageParams *paged.PageParams) (*ActivitiesPaged, []*Project, error) {
	activitiesFilter := toFilter(principal, filter)

	activitiesPage, projects, err := a.activityRepository.FindActivities(ctx, activitiesFilter, pageParams)
	if err != nil {
		return nil, nil, err
	}

	return activitiesPage, projects, err
}

func (a *ActitivityService) TimeReports(ctx context.Context, principal *shared.Principal, filter *ActivityFilter, aggregateBy string) ([]*ActivityTimeReportItem, error) {
	activitiesFilter := toFilter(principal, filter)

	switch aggregateBy {
	case "week":
		return a.activityRepository.TimeReportByWeek(ctx, activitiesFilter)
	case "month":
		return a.activityRepository.TimeReportByMonth(ctx, activitiesFilter)
	case "quarter":
		return a.activityRepository.TimeReportByQuarter(ctx, activitiesFilter)
	case "day":
		return a.activityRepository.TimeReportByDay(ctx, activitiesFilter)
	default:
		return a.activityRepository.TimeReportByDay(ctx, activitiesFilter)
	}
}

func (a *ActitivityService) ProjectReports(ctx context.Context, principal *shared.Principal, filter *ActivityFilter) ([]*ActivityProjectReportItem, error) {
	activitiesFilter := toFilter(principal, filter)
	return a.activityRepository.ProjectReport(ctx, activitiesFilter)
}

// CreateActivity creates a new activity
func (a *ActitivityService) CreateActivity(ctx context.Context, principal *shared.Principal, activity *Activity) (*Activity, error) {
	activity.ID = uuid.New()
	activity.OrganizationID = principal.OrganizationID
	activity.Username = principal.Username

	// Extract tag names from Tag objects
	tagNames := make([]string, len(activity.Tags))
	for i, tag := range activity.Tags {
		tagNames[i] = tag.Name
	}

	// Validate tags
	if err := a.tagService.ValidateTags(tagNames); err != nil {
		return nil, err
	}

	// Prepare tags with colors for repository and update activity.Tags
	tagsWithColors := a.tagService.PrepareTagsWithColors(tagNames)
	activity.Tags = tagsWithColors

	var newActivity *Activity
	err := a.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			insertedActivity, err := a.activityRepository.InsertActivity(ctx, activity)
			if err != nil {
				return err
			}
			newActivity = insertedActivity

			// Sync tags after activity creation
			err = a.tagRepository.SyncTagsForActivity(ctx, activity.ID, activity.OrganizationID, tagsWithColors)
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return newActivity, nil
}

// DeleteActivityByID deletes an activity
func (a *ActitivityService) DeleteActivityByID(ctx context.Context, principal *shared.Principal, activityID uuid.UUID) error {
	if principal.HasRole("ROLE_ADMIN") {
		return a.repositoryTxer.InTx(
			ctx,
			func(ctx context.Context) error {
				return a.activityRepository.DeleteActivityByID(ctx, principal.OrganizationID, activityID)
			},
		)
	}
	return a.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			return a.activityRepository.DeleteActivityByIDAndUsername(ctx, principal.OrganizationID, activityID, principal.Username)
		},
	)
}

// UpdateActivity updates an activity
func (a *ActitivityService) UpdateActivity(ctx context.Context, principal *shared.Principal, activity *Activity) (*Activity, error) {
	// Extract tag names from Tag objects
	tagNames := make([]string, len(activity.Tags))
	for i, tag := range activity.Tags {
		tagNames[i] = tag.Name
	}

	// Validate tags
	if err := a.tagService.ValidateTags(tagNames); err != nil {
		return nil, err
	}

	// Prepare tags with colors for repository and update activity.Tags
	tagsWithColors := a.tagService.PrepareTagsWithColors(tagNames)
	activity.Tags = tagsWithColors

	var activityUpdate *Activity
	if principal.HasRole("ROLE_ADMIN") {
		err := a.repositoryTxer.InTx(
			ctx,
			func(ctx context.Context) error {
				updatedActivity, err := a.activityRepository.UpdateActivity(ctx, principal.OrganizationID, activity)
				if err != nil {
					return err
				}
				activityUpdate = updatedActivity

				// Sync tags after activity update
				err = a.tagRepository.SyncTagsForActivity(ctx, activity.ID, principal.OrganizationID, tagsWithColors)
				if err != nil {
					return err
				}

				return nil
			},
		)
		if err != nil {
			return nil, err
		}
		return activityUpdate, nil
	}
	err := a.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			updatedActivity, err := a.activityRepository.UpdateActivityByUsername(ctx, principal.OrganizationID, activity, principal.Username)
			if err != nil {
				return err
			}
			activityUpdate = updatedActivity

			// Sync tags after activity update
			err = a.tagRepository.SyncTagsForActivity(ctx, activity.ID, principal.OrganizationID, tagsWithColors)
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return activityUpdate, nil
}

func (a *ActitivityService) WriteAsCSV(activities []*Activity, projects []*Project, w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'

	defer csvWriter.Flush()

	headers := []string{"Date", "Start", "End", "Duration", "Project", "Description"}

	err := csvWriter.Write(headers)
	if err != nil {
		return err
	}

	// prepare projects
	projectsById := make(map[uuid.UUID]*Project)
	for _, project := range projects {
		projectsById[project.ID] = project
	}

	// write records for activities
	for _, activity := range activities {
		record := []string{
			activity.Start.Format("2006-01-02"),
			activity.Start.Format("15:04"),
			activity.End.Format("15:04"),
			activity.DurationFormatted(),
			projectsById[activity.ProjectID].Title,
			activity.Description,
		}
		err := csvWriter.Write(record)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *ActitivityService) WriteAsExcel(activities []*Activity, projects []*Project, w io.Writer) error {
	// prepare projects
	projectsById := make(map[uuid.UUID]*Project)
	for _, project := range projects {
		projectsById[project.ID] = project
	}

	f := excelize.NewFile()
	f.SetActiveSheet(0)
	err := f.SetSheetName("Sheet1", "Activities")
	if err != nil {
		return err
	}

	_ = f.SetCellValue("Activities", "A1", "Project")
	_ = f.SetCellValue("Activities", "B1", "Date")
	_ = f.SetCellValue("Activities", "C1", "Start")
	_ = f.SetCellValue("Activities", "D1", "End")
	_ = f.SetCellValue("Activities", "E1", "Hours")
	_ = f.SetCellValue("Activities", "F1", "Description")

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:  "color",
			Color: []string{"#adadad"},
		},
	})

	styleDuration, _ := f.NewStyle(&excelize.Style{
		NumFmt: 4,
	})
	_ = f.SetCellStyle("Activities", "A1", "F1", style)

	descriptionStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			WrapText: true,
		},
	})

	for i, activity := range activities {
		idx := i + 2

		duration, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", activity.DurationDecimal()), 64)

		_ = f.SetCellValue("Activities", fmt.Sprintf("A%v", idx), projectsById[activity.ProjectID].Title)
		_ = f.SetCellValue("Activities", fmt.Sprintf("B%v", idx), activity.Start.Format("2006-01-02"))
		_ = f.SetCellValue("Activities", fmt.Sprintf("C%v", idx), activity.Start.Format("15:04"))
		_ = f.SetCellValue("Activities", fmt.Sprintf("D%v", idx), activity.End.Format("15:04"))

		_ = f.SetCellValue("Activities", fmt.Sprintf("E%v", idx), duration)
		_ = f.SetCellStyle("Activities", fmt.Sprintf("E%v", idx), fmt.Sprintf("E%v", idx), styleDuration)

		_ = f.SetCellValue("Activities", fmt.Sprintf("F%v", idx), activity.Description)
		_ = f.SetCellStyle("Activities", fmt.Sprintf("F%v", idx), fmt.Sprintf("F%v", idx), descriptionStyle)
	}

	return f.Write(w)
}

// ParseTagsFromString parses a comma/space separated string of tags into a normalized slice
func (a *ActitivityService) ParseTagsFromString(tagString string) []string {
	return a.tagService.ParseTagsFromString(tagString)
}

// ValidateTags validates a slice of tag names
func (a *ActitivityService) ValidateTags(tags []string) error {
	return a.tagService.ValidateTags(tags)
}

// GetTagsForAutocomplete returns matching tags for autocomplete within organization
func (a *ActitivityService) GetTagsForAutocomplete(ctx context.Context, principal *shared.Principal, query string) ([]*Tag, error) {
	return a.tagService.GetTagsForAutocomplete(ctx, principal.OrganizationID, query)
}

func toFilter(principal *shared.Principal, filter *ActivityFilter) *ActivitiesFilter {
	activitiesFilter := &ActivitiesFilter{
		Start:          filter.Start(),
		End:            filter.End(),
		SortBy:         filter.sortBy,
		SortOrder:      filter.sortOrder,
		OrganizationID: principal.OrganizationID,
		Tags:           filter.Tags(),
	}

	if !principal.HasRole("ROLE_ADMIN") {
		activitiesFilter.Username = principal.Username
	}

	return activitiesFilter
}
