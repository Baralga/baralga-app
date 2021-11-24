package main

import (
	"context"

	"github.com/baralga/paged"
	"github.com/google/uuid"
)

// ReadActivitiesWithProjects reads activities with their associated projects
func (a *app) ReadActivitiesWithProjects(ctx context.Context, principal *Principal, filter *ActivityFilter, pageParams *paged.PageParams) ([]*Activity, []*Project, error) {
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

	return activitiesPage.Activities, projects, err
}

// CreateActivity creates a new activity
func (a *app) CreateActivity(ctx context.Context, principal *Principal, activity *Activity) (*Activity, error) {
	activity.ID = uuid.New()
	activity.OrganizationID = principal.OrganizationID
	activity.Username = principal.Username
	return a.ActivityRepository.InsertActivity(ctx, activity)
}

// DeleteActivityByID deletes an activity
func (a *app) DeleteActivityByID(ctx context.Context, principal *Principal, activityID uuid.UUID) error {
	if principal.HasRole("ROLE_ADMIN") {
		return a.ActivityRepository.DeleteActivityByID(ctx, principal.OrganizationID, activityID)
	}
	return a.ActivityRepository.DeleteActivityByIDAndUsername(ctx, principal.OrganizationID, activityID, principal.Username)
}

// UpdateActivity updates an activity
func (a *app) UpdateActivity(ctx context.Context, principal *Principal, activity *Activity) (*Activity, error) {
	if principal.HasRole("ROLE_ADMIN") {
		return a.ActivityRepository.UpdateActivity(ctx, principal.OrganizationID, activity)
	}
	return a.ActivityRepository.UpdateActivityByUsername(ctx, principal.OrganizationID, activity, principal.Username)
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
