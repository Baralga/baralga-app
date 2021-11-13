package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/baralga/paged"
	"github.com/google/uuid"
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

		_, err := activityRepository.InsertActivity(
			context.Background(),
			activtiy,
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

		_, err := activityRepository.InsertActivity(
			context.Background(),
			activtiy,
		)
		is.NoErr(err)

		err = activityRepository.DeleteActivityByIDAndUsername(context.Background(), activtiy.OrganizationID, activtiy.ID, "user1")
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
		err := activityRepository.DeleteActivityByIDAndUsername(
			context.Background(),
			organizationIDSample,
			uuid.MustParse("f8d8a2ac-3f3e-11ec-9bbc-0242ac130002"),
			"user1",
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

		_, err := activityRepository.InsertActivity(
			context.Background(),
			activtiy,
		)
		is.NoErr(err)

		activtiy.Description = "My updated Description"

		activityUpdate, err := activityRepository.UpdateActivity(context.Background(), organizationIDSample, activtiy)
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

		_, err := activityRepository.InsertActivity(
			context.Background(),
			activtiy,
		)
		is.NoErr(err)

		activtiy.Description = "My updated Description"

		activityUpdate, err := activityRepository.UpdateActivityByUsername(context.Background(), organizationIDSample, activtiy, "user1")
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

		_, err := activityRepository.InsertActivity(
			context.Background(),
			activtiy,
		)
		is.NoErr(err)

		activtiy.Description = "My updated Description"

		_, err = activityRepository.UpdateActivityByUsername(context.Background(), organizationIDSample, activtiy, "otherUser")
		is.True(errors.Is(err, ErrActivityNotFound))
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
				ProjectID:      uuid.MustParse("00000000-0000-0000-1111-000000000001"),
				OrganizationID: organizationIDSample,
				Username:       "user1",
			},
		},
	}
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
