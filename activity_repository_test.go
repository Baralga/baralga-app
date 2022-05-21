package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/baralga/paged"
	"github.com/baralga/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/matryer/is"
)

func TestActivityRepository(t *testing.T) {
	// skip in short mode
	if testing.Short() {
		return
	}

	is := is.New(t)

	// Setup database
	ctx := context.Background()
	dbContainer, connPool, err := setupDatabase(ctx)
	if err != nil {
		t.Error(err)
	}

	defer func() {
		err := dbContainer.Terminate(ctx)
		if err != nil {
			t.Log(err)
		}
	}()

	activityRepository := NewDbActivityRepository(connPool)
	repositoryTxer := NewDbRepositoryTxer(connPool)

	t.Run("FindActivitiesByOrganizationId", func(t *testing.T) {
		filter := &ActivitiesFilter{
			Start:          time.Now().AddDate(-1, 0, 0),
			End:            time.Now(),
			OrganizationID: organizationIDSample,
		}
		activityiesPage, err := activityRepository.FindActivities(
			context.Background(),
			filter,
			&paged.PageParams{
				Page: 0,
				Size: 50,
			},
		)

		is.NoErr(err)
		is.Equal(len(activityiesPage.Activities), 1)
		is.Equal(activityiesPage.Page.TotalElements, 1)
		is.True(activityiesPage != nil)
	})

	t.Run("FindActivitiesByOrganizationIdAndUsername", func(t *testing.T) {
		filter := &ActivitiesFilter{
			Start:          time.Now().AddDate(-1, 0, 0),
			End:            time.Now(),
			Username:       "admin",
			OrganizationID: organizationIDSample,
		}
		activityiesPage, err := activityRepository.FindActivities(
			context.Background(),
			filter,
			&paged.PageParams{
				Page: 0,
				Size: 50,
			},
		)

		is.NoErr(err)
		is.Equal(len(activityiesPage.Activities), 1)
		is.Equal(activityiesPage.Page.TotalElements, 1)
		is.True(activityiesPage != nil)
	})

	t.Run("InsertAndFindAndDeleteActivity", func(t *testing.T) {
		start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
		end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

		activtiy := &Activity{
			ID:             uuid.New(),
			ProjectID:      projectIDSample,
			OrganizationID: organizationIDSample,
			Start:          start,
			End:            end,
		}

		err := repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := activityRepository.InsertActivity(
					ctx,
					activtiy,
				)
				return err
			},
		)
		is.NoErr(err)

		activityFound, err := activityRepository.FindActivityByID(context.Background(), activtiy.ID, organizationIDSample)
		is.NoErr(err)
		is.Equal(activtiy.ID, activityFound.ID)
		is.Equal(activtiy.Description, activityFound.Description)

		err = activityRepository.DeleteActivityByID(context.Background(), activtiy.OrganizationID, activtiy.ID)
		is.NoErr(err)
	})

	t.Run("InsertAndFindAndDeleteActivityForUser", func(t *testing.T) {
		start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
		end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

		activtiy := &Activity{
			ID:             uuid.New(),
			ProjectID:      projectIDSample,
			OrganizationID: organizationIDSample,
			Start:          start,
			End:            end,
			Username:       "user1",
		}

		err := repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := activityRepository.InsertActivity(
					ctx,
					activtiy,
				)
				return err
			},
		)
		is.NoErr(err)

		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				return activityRepository.DeleteActivityByIDAndUsername(ctx, activtiy.OrganizationID, activtiy.ID, "user1")
			},
		)
		is.NoErr(err)
	})

	t.Run("FindNonExistingActivityByID", func(t *testing.T) {
		_, err := activityRepository.FindActivityByID(
			context.Background(),
			uuid.MustParse("f8d8a2ac-3f3e-11ec-9bbc-0242ac130002"),
			organizationIDSample,
		)

		is.True(errors.Is(err, ErrActivityNotFound))
	})

	t.Run("DeleteNonExistingActivityByID", func(t *testing.T) {
		err := activityRepository.DeleteActivityByID(
			context.Background(),
			organizationIDSample,
			uuid.MustParse("f8d8a2ac-3f3e-11ec-9bbc-0242ac130002"),
		)

		is.True(errors.Is(err, ErrActivityNotFound))
	})

	t.Run("DeleteNonExistingActivityByIDAndUsername", func(t *testing.T) {
		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				err := activityRepository.DeleteActivityByIDAndUsername(
					ctx,
					organizationIDSample,
					uuid.MustParse("f8d8a2ac-3f3e-11ec-9bbc-0242ac130002"),
					"user1",
				)
				return err
			},
		)

		is.True(errors.Is(err, ErrActivityNotFound))
	})

	t.Run("InsertAndUpdateActivity", func(t *testing.T) {
		start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
		end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

		activtiy := &Activity{
			ID:             uuid.New(),
			ProjectID:      projectIDSample,
			OrganizationID: organizationIDSample,
			Description:    "My Description",
			Start:          start,
			End:            end,
			Username:       "user1",
		}

		err := repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := activityRepository.InsertActivity(
					ctx,
					activtiy,
				)
				return err
			},
		)
		is.NoErr(err)

		activtiy.Description = "My updated Description"

		var activityUpdate *Activity
		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				a, err := activityRepository.UpdateActivity(ctx, organizationIDSample, activtiy)
				if err != nil {
					return err
				}
				activityUpdate = a
				return nil
			},
		)
		is.NoErr(err)
		is.Equal("My updated Description", activityUpdate.Description)
	})

	t.Run("InsertAndUpdateActivityForUser", func(t *testing.T) {
		start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
		end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

		activtiy := &Activity{
			ID:             uuid.New(),
			ProjectID:      projectIDSample,
			OrganizationID: organizationIDSample,
			Description:    "My Description",
			Start:          start,
			End:            end,
			Username:       "user1",
		}

		err := repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := activityRepository.InsertActivity(
					ctx,
					activtiy,
				)
				return err
			},
		)
		is.NoErr(err)

		activtiy.Description = "My updated Description"

		var activityUpdate *Activity
		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				a, err := activityRepository.UpdateActivityByUsername(ctx, organizationIDSample, activtiy, "user1")
				if err != nil {
					return err
				}
				activityUpdate = a
				return nil
			},
		)
		is.NoErr(err)
		is.Equal("My updated Description", activityUpdate.Description)
	})

	t.Run("InsertAndUpdateActivityForNonMatchingUser", func(t *testing.T) {
		start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
		end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

		activtiy := &Activity{
			ID:             uuid.New(),
			ProjectID:      projectIDSample,
			OrganizationID: organizationIDSample,
			Description:    "My Description",
			Start:          start,
			End:            end,
			Username:       "user1",
		}

		err := repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := activityRepository.InsertActivity(
					ctx,
					activtiy,
				)
				return err
			},
		)
		is.NoErr(err)

		activtiy.Description = "My updated Description"

		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := activityRepository.UpdateActivityByUsername(ctx, organizationIDSample, activtiy, "otherUser")
				return err
			},
		)
		is.True(errors.Is(err, ErrActivityNotFound))
	})
}

func TestActivityRepositoryReports(t *testing.T) {
	// skip in short mode
	if testing.Short() {
		return
	}

	is := is.New(t)

	// Setup database
	ctx := context.Background()
	dbContainer, connPool, err := setupDatabase(ctx)
	if err != nil {
		t.Error(err)
	}

	err = insertSampleActivitiesForReports(ctx, connPool)
	if err != nil {
		t.Error(err)
	}

	defer func() {
		err := dbContainer.Terminate(ctx)
		if err != nil {
			t.Log(err)
		}
	}()

	activityRepository := NewDbActivityRepository(connPool)

	start, _ := time.Parse(time.RFC3339, "2022-01-01T10:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2023-01-01T11:00:00.000Z")
	filter := &ActivitiesFilter{
		Start:          start,
		End:            end,
		OrganizationID: organizationIDSample,
	}

	t.Run("TimeReportByDay", func(t *testing.T) {
		// Arrange

		// Act
		reportItems, err := activityRepository.TimeReportByDay(
			context.Background(),
			filter,
		)

		// Assert
		is.NoErr(err)
		is.Equal(len(reportItems), 4)
		is.Equal(120, reportItems[3].DurationInMinutesTotal)
	})

	t.Run("TimeReportByWeek", func(t *testing.T) {
		// Arrange

		// Act
		reportItems, err := activityRepository.TimeReportByWeek(
			context.Background(),
			filter,
		)

		// Assert
		is.NoErr(err)
		is.Equal(len(reportItems), 4)
		is.Equal(120, reportItems[3].DurationInMinutesTotal)
	})

	t.Run("TimeReportByMonth", func(t *testing.T) {
		// Arrange

		// Act
		reportItems, err := activityRepository.TimeReportByMonth(
			context.Background(),
			filter,
		)

		// Assert
		is.NoErr(err)
		is.Equal(len(reportItems), 3)
		is.Equal(180, reportItems[2].DurationInMinutesTotal)
	})

	t.Run("TimeReportByQuarter", func(t *testing.T) {
		// Arrange

		// Act
		reportItems, err := activityRepository.TimeReportByQuarter(
			context.Background(),
			filter,
		)

		// Assert
		is.NoErr(err)
		is.Equal(len(reportItems), 2)
		is.Equal(240, reportItems[1].DurationInMinutesTotal)
	})

	t.Run("ProjectReport", func(t *testing.T) {
		// Arrange

		// Act
		reportItems, err := activityRepository.ProjectReport(
			context.Background(),
			filter,
		)

		// Assert
		is.NoErr(err)
		is.Equal(len(reportItems), 1)
		is.Equal(300, reportItems[0].DurationInMinutesTotal)
	})
}

type InMemActivityRepository struct {
	activities []*Activity
}

var _ ActivityRepository = (*InMemActivityRepository)(nil)

func NewInMemActivityRepository() *InMemActivityRepository {
	return &InMemActivityRepository{
		activities: []*Activity{
			{
				ID:             uuid.MustParse("00000000-0000-0000-2222-000000000001"),
				ProjectID:      projectIDSample,
				OrganizationID: organizationIDSample,
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
			Quarter:                util.Quarter(a.Start),
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
			Quarter:                util.Quarter(a.Start),
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

func (r *InMemActivityRepository) FindActivities(ctx context.Context, filter *ActivitiesFilter, pageParams *paged.PageParams) (*ActivitiesPaged, error) {
	activitiesPage := &ActivitiesPaged{
		Activities: r.activities,
		Page: &paged.Page{
			Size:          len(r.activities),
			Number:        0,
			TotalElements: len(r.activities),
			TotalPages:    1,
		},
	}
	return activitiesPage, nil
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

func insertSampleActivitiesForReports(ctx context.Context, connPool *pgxpool.Pool) error {
	_, err := connPool.Exec(
		ctx,
		fmt.Sprintf(
			`INSERT INTO activities 
			(activity_id, start_time, end_time, description, project_id, org_id, username) 
			VALUES 
			('%v', '2022-01-10 14:00:00-00', '2022-01-10 15:00:00-00', 'My Desc', '%v', '%v', 'admin')`,
			uuid.New().String(),
			projectIDSample,
			organizationIDSample,
		),
	)
	if err != nil {
		return err
	}

	_, err = connPool.Exec(
		ctx,
		fmt.Sprintf(
			`INSERT INTO activities 
			(activity_id, start_time, end_time, description, project_id, org_id, username) 
			VALUES 
			('%v', '2022-01-10 15:00:00-00', '2022-01-10 16:00:00-00', 'My Desc', '%v', '%v', 'admin')`,
			uuid.New().String(),
			projectIDSample,
			organizationIDSample,
		),
	)
	if err != nil {
		return err
	}

	_, err = connPool.Exec(
		ctx,
		fmt.Sprintf(
			`INSERT INTO activities 
			(activity_id, start_time, end_time, description, project_id, org_id, username) 
			VALUES 
			('%v', '2022-01-17 14:00:00-00', '2022-01-17 15:00:00-00', 'My Desc', '%v', '%v', 'admin')`,
			uuid.New().String(),
			projectIDSample,
			organizationIDSample,
		),
	)
	if err != nil {
		return err
	}

	_, err = connPool.Exec(
		ctx,
		fmt.Sprintf(
			`INSERT INTO activities 
			(activity_id, start_time, end_time, description, project_id, org_id, username) 
			VALUES 
			('%v', '2022-02-10 15:00:00-00', '2022-02-10 16:00:00-00', 'My Desc', '%v', '%v', 'admin')`,
			uuid.New().String(),
			projectIDSample,
			organizationIDSample,
		),
	)
	if err != nil {
		return err
	}

	_, err = connPool.Exec(
		ctx,
		fmt.Sprintf(
			`INSERT INTO activities 
			(activity_id, start_time, end_time, description, project_id, org_id, username) 
			VALUES 
			('%v', '2022-04-10 15:00:00-00', '2022-04-10 16:00:00-00', 'My Desc', '%v', '%v', 'admin')`,
			uuid.New().String(),
			projectIDSample,
			organizationIDSample,
		),
	)
	if err != nil {
		return err
	}
	return nil
}
