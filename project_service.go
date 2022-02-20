package main

import (
	"context"

	"github.com/google/uuid"
)

func (a *app) CreateProject(ctx context.Context, principal *Principal, project *Project) (*Project, error) {
	project.ID = uuid.New()
	project.OrganizationID = principal.OrganizationID

	var projectCreated *Project
	err := a.RepositoryTxer.InTx(
		context.Background(),
		func(ctx context.Context) error {
			a, err := a.ProjectRepository.InsertProject(ctx, project)
			if err != nil {
				return err
			}
			projectCreated = a
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return projectCreated, nil
}

func (a *app) UpdateProject(ctx context.Context, organizationID uuid.UUID, project *Project) (*Project, error) {
	var projectUpdated *Project
	err := a.RepositoryTxer.InTx(
		context.Background(),
		func(ctx context.Context) error {
			p, err := a.ProjectRepository.UpdateProject(ctx, organizationID, project)
			if err != nil {
				return err
			}
			projectUpdated = p
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return projectUpdated, nil
}

func (a *app) DeleteProjectByID(ctx context.Context, principal *Principal, projectID uuid.UUID) error {
	return a.RepositoryTxer.InTx(
		ctx,
		func(ctxWithTx context.Context) error {
			return a.ProjectRepository.DeleteProjectByID(ctx, principal.OrganizationID, projectID)
		},
	)
}
