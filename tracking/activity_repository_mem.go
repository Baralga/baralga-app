package tracking

import (
	"context"

	"github.com/baralga/shared"
	"github.com/baralga/shared/paged"
	time_utils "github.com/baralga/shared/time"
	"github.com/google/uuid"
)

type InMemActivityRepository struct {
	activities []*Activity
}

var _ ActivityRepository = (*InMemActivityRepository)(nil)

func NewInMemActivityRepository() *InMemActivityRepository {
	return &InMemActivityRepository{
		activities: []*Activity{
			{
				ID:             uuid.MustParse("00000000-0000-0000-2222-000000000001"),
				ProjectID:      shared.ProjectIDSample,
				OrganizationID: shared.OrganizationIDSample,
				Username:       "user1",
			},
		},
	}
}

func (r *InMemActivityRepository) TimeReportByDay(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error) {
	var reportItems []*ActivityTimeReportItem
	for _, a := range r.activities {
		_, w := a.Start.ISOWeek()
		reportItem := &ActivityTimeReportItem{
			Year:                   a.Start.Year(),
			Month:                  int(a.Start.Month()),
			Quarter:                time_utils.Quarter(a.Start),
			Week:                   w,
			Day:                    a.Start.Day(),
			DurationInMinutesTotal: 60,
		}
		reportItems = append(reportItems, reportItem)
	}
	return reportItems, nil
}

func (r *InMemActivityRepository) TimeReportByWeek(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error) {
	var reportItems []*ActivityTimeReportItem
	for _, a := range r.activities {
		_, w := a.Start.ISOWeek()
		reportItem := &ActivityTimeReportItem{
			Year:                   a.Start.Year(),
			Week:                   w,
			DurationInMinutesTotal: 60,
		}
		reportItems = append(reportItems, reportItem)
	}
	return reportItems, nil
}

func (r *InMemActivityRepository) TimeReportByMonth(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error) {
	var reportItems []*ActivityTimeReportItem
	for _, a := range r.activities {
		reportItem := &ActivityTimeReportItem{
			Year:                   a.Start.Year(),
			Month:                  int(a.Start.Month()),
			DurationInMinutesTotal: 60,
		}
		reportItems = append(reportItems, reportItem)
	}
	return reportItems, nil
}

func (r *InMemActivityRepository) TimeReportByQuarter(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error) {
	var reportItems []*ActivityTimeReportItem
	for _, a := range r.activities {
		reportItem := &ActivityTimeReportItem{
			Year:                   a.Start.Year(),
			Quarter:                time_utils.Quarter(a.Start),
			DurationInMinutesTotal: 60,
		}
		reportItems = append(reportItems, reportItem)
	}
	return reportItems, nil
}

func (r *InMemActivityRepository) ProjectReport(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityProjectReportItem, error) {
	var reportItems []*ActivityProjectReportItem
	for _, a := range r.activities {
		reportItem := &ActivityProjectReportItem{
			ProjectID:              a.ProjectID,
			ProjectTitle:           "My Project",
			DurationInMinutesTotal: 60,
		}
		reportItems = append(reportItems, reportItem)
	}
	return reportItems, nil
}

func (r *InMemActivityRepository) FindActivities(ctx context.Context, filter *ActivitiesFilter, pageParams *paged.PageParams) (*ActivitiesPaged, []*Project, error) {
	activitiesPage := &ActivitiesPaged{
		Activities: r.activities,
		Page: &paged.Page{
			Size:          len(r.activities),
			Number:        0,
			TotalElements: len(r.activities),
			TotalPages:    1,
		},
	}
	projects := []*Project{
		{
			ID:             shared.ProjectIDSample,
			Title:          "My Project",
			OrganizationID: shared.OrganizationIDSample,
		},
	}

	return activitiesPage, projects, nil
}

func (r *InMemActivityRepository) FindActivityByID(ctx context.Context, activityID, organizationID uuid.UUID) (*Activity, error) {
	for _, a := range r.activities {
		if a.ID == activityID {
			return a, nil
		}
	}
	return nil, ErrActivityNotFound
}

func (r *InMemActivityRepository) InsertActivity(ctx context.Context, activity *Activity) (*Activity, error) {
	r.activities = append(r.activities, activity)
	return activity, nil
}

func (r *InMemActivityRepository) DeleteActivityByID(ctx context.Context, organizationID, activityID uuid.UUID) error {
	for i, a := range r.activities {
		if a.ID == activityID {
			r.activities = append(r.activities[:i], r.activities[i+1:]...)
			return nil
		}
	}
	return ErrActivityNotFound
}

func (r *InMemActivityRepository) DeleteActivityByIDAndUsername(ctx context.Context, organizationID, activityID uuid.UUID, username string) error {
	for i, a := range r.activities {
		if a.ID == activityID && a.Username == username {
			r.activities = append(r.activities[:i], r.activities[i+1:]...)
			return nil
		}
	}
	return ErrActivityNotFound
}

func (r *InMemActivityRepository) UpdateActivity(ctx context.Context, organizationID uuid.UUID, activity *Activity) (*Activity, error) {
	for i, a := range r.activities {
		if a.ID == activity.ID {
			r.activities[i] = activity
			return activity, nil
		}
	}
	return nil, ErrActivityNotFound
}

func (r *InMemActivityRepository) UpdateActivityByUsername(ctx context.Context, organizationID uuid.UUID, activity *Activity, username string) (*Activity, error) {
	for i, a := range r.activities {
		if a.ID == activity.ID && a.Username == username {
			r.activities[i] = activity
			return activity, nil
		}
	}
	return nil, ErrActivityNotFound
}
