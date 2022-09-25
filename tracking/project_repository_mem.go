package tracking

import (
	"context"

	"github.com/baralga/shared"
	"github.com/baralga/shared/paged"
	"github.com/google/uuid"
)

type InMemProjectRepository struct {
	projects []*Project
}

var _ ProjectRepository = (*InMemProjectRepository)(nil)

func NewInMemProjectRepository() *InMemProjectRepository {
	return &InMemProjectRepository{
		projects: []*Project{
			{
				ID:             shared.ProjectIDSample,
				Title:          "My Project",
				OrganizationID: shared.OrganizationIDSample,
			},
		},
	}
}

func (r *InMemProjectRepository) InsertProject(ctx context.Context, project *Project) (*Project, error) {
	r.projects = append(r.projects, project)
	return project, nil
}

func (r *InMemProjectRepository) FindProjects(ctx context.Context, organizationID uuid.UUID, pageParams *paged.PageParams) (*ProjectsPaged, error) {
	projectsPaged := &ProjectsPaged{
		Projects: r.projects,
		Page:     pageParams.PageOfTotal(len(r.projects)),
	}
	return projectsPaged, nil
}

func (r *InMemProjectRepository) FindProjectsByIDs(ctx context.Context, organizationID uuid.UUID, projectIDs []uuid.UUID) ([]*Project, error) {
	var projects []*Project

	for _, projectID := range projectIDs {
		for _, p := range r.projects {
			if p.ID == projectID {
				projects = append(projects, p)
				break
			}
		}
	}

	return projects, nil
}

func (r *InMemProjectRepository) UpdateProject(ctx context.Context, organizationID uuid.UUID, project *Project) (*Project, error) {
	for i, p := range r.projects {
		if p.ID == project.ID {
			r.projects[i] = project
			return project, nil
		}
	}
	return nil, ErrProjectNotFound
}

func (r *InMemProjectRepository) DeleteProjectByID(ctx context.Context, organizationID, projectID uuid.UUID) error {
	for i, a := range r.projects {
		if a.ID == projectID {
			r.projects = append(r.projects[:i], r.projects[i+1:]...)
			return nil
		}
	}
	return ErrProjectNotFound
}

func (r *InMemProjectRepository) ArchiveProjectByID(ctx context.Context, organizationID, projectID uuid.UUID) error {
	for i, a := range r.projects {
		if a.ID == projectID {
			r.projects[i].Active = false
			return nil
		}
	}
	return ErrProjectNotFound
}

func (r *InMemProjectRepository) FindProjectByID(ctx context.Context, organizationID, projectID uuid.UUID) (*Project, error) {
	for _, a := range r.projects {
		if a.ID == projectID {
			return a, nil
		}
	}
	return nil, ErrProjectNotFound
}
