package main

import (
	"context"

	"github.com/google/uuid"
)

func (a *app) CreateProject(ctx context.Context, principal *Principal, project *Project) (*Project, error) {
	project.ID = uuid.New()
	project.OrganizationID = principal.OrganizationID
	return a.ProjectRepository.InsertProject(ctx, project)
}

func (a *app) DeleteProjectByID(ctx context.Context, principal *Principal, projectID uuid.UUID) error {
	return a.ProjectRepository.DeleteProjectByID(ctx, principal.OrganizationID, projectID)
}
