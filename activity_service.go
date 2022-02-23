package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/baralga/paged"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// ReadActivitiesWithProjects reads activities with their associated projects
func (a *app) ReadActivitiesWithProjects(ctx context.Context, principal *Principal, filter *ActivityFilter, pageParams *paged.PageParams) (*ActivitiesPaged, []*Project, error) {
	activitiesFilter := &ActivitiesFilter{
		Start:          filter.Start(),
		End:            filter.End(),
		OrganizationID: principal.OrganizationID,
	}

	if !principal.HasRole("ROLE_ADMIN") {
		activitiesFilter.Username = principal.Username
	}

	activitiesPage, err := a.ActivityRepository.FindActivities(ctx, activitiesFilter, pageParams)
	if err != nil {
		return nil, nil, err
	}

	projectIDs := distinctProjectIds(activitiesPage)
	projects, err := a.ProjectRepository.FindProjectsByIDs(ctx, principal.OrganizationID, projectIDs)
	if err != nil {
		return nil, nil, err
	}

	return activitiesPage, projects, err
}

// CreateActivity creates a new activity
func (a *app) CreateActivity(ctx context.Context, principal *Principal, activity *Activity) (*Activity, error) {
	activity.ID = uuid.New()
	activity.OrganizationID = principal.OrganizationID
	activity.Username = principal.Username

	var newActivity *Activity
	err := a.RepositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			a, err := a.ActivityRepository.InsertActivity(ctx, activity)
			if err != nil {
				return err
			}
			newActivity = a
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return newActivity, nil
}

// DeleteActivityByID deletes an activity
func (a *app) DeleteActivityByID(ctx context.Context, principal *Principal, activityID uuid.UUID) error {
	if principal.HasRole("ROLE_ADMIN") {
		return a.RepositoryTxer.InTx(
			ctx,
			func(ctx context.Context) error {
				return a.ActivityRepository.DeleteActivityByID(ctx, principal.OrganizationID, activityID)
			},
		)
	}
	return a.RepositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			return a.ActivityRepository.DeleteActivityByIDAndUsername(ctx, principal.OrganizationID, activityID, principal.Username)
		},
	)
}

// UpdateActivity updates an activity
func (a *app) UpdateActivity(ctx context.Context, principal *Principal, activity *Activity) (*Activity, error) {
	var activityUpdate *Activity
	if principal.HasRole("ROLE_ADMIN") {
		err := a.RepositoryTxer.InTx(
			ctx,
			func(ctx context.Context) error {
				a, err := a.ActivityRepository.UpdateActivity(ctx, principal.OrganizationID, activity)
				if err != nil {
					return err
				}
				activityUpdate = a
				return nil
			},
		)
		if err != nil {
			return nil, err
		}
		return activityUpdate, nil
	}
	err := a.RepositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			a, err := a.ActivityRepository.UpdateActivityByUsername(ctx, principal.OrganizationID, activity, principal.Username)
			if err != nil {
				return err
			}
			activityUpdate = a
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return activityUpdate, nil
}

func (a *app) WriteAsCSV(activities []*Activity, projects []*Project, w io.Writer) error {
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

func (a *app) WriteAsExcel(activities []*Activity, projects []*Project, w io.Writer) error {
	// prepare projects
	projectsById := make(map[uuid.UUID]*Project)
	for _, project := range projects {
		projectsById[project.ID] = project
	}

	f := excelize.NewFile()
	f.SetActiveSheet(0)
	f.SetSheetName("Sheet1", "Activities")

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

func distinctProjectIds(activitiesPage *ActivitiesPaged) []uuid.UUID {
	pIDs := make(map[uuid.UUID]bool)

	for _, activity := range activitiesPage.Activities {
		pIDs[activity.ProjectID] = true
	}

	projectIDs := make([]uuid.UUID, len(pIDs))
	i := 0
	for projectID := range pIDs {
		projectIDs[i] = projectID
		i++
	}

	return projectIDs
}
