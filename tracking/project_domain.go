package tracking

import (
	"context"

	"github.com/baralga/shared/util/paged"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var ErrProjectNotFound = errors.New("project not found")

type Project struct {
	ID             uuid.UUID
	Title          string
	Description    string
	Active         bool
	OrganizationID uuid.UUID
}

type ProjectsPaged struct {
	Projects []*Project
	Page     *paged.Page
}

type ProjectRepository interface {
	FindProjects(ctx context.Context, organizationID uuid.UUID, pageParams *paged.PageParams) (*ProjectsPaged, error)
	FindProjectsByIDs(ctx context.Context, organizationID uuid.UUID, projectIDs []uuid.UUID) ([]*Project, error)
	FindProjectByID(ctx context.Context, organizationID, projectID uuid.UUID) (*Project, error)
	InsertProject(ctx context.Context, project *Project) (*Project, error)
	UpdateProject(ctx context.Context, organizationID uuid.UUID, project *Project) (*Project, error)
	ArchiveProjectByID(ctx context.Context, organizationID, projectID uuid.UUID) error
	DeleteProjectByID(ctx context.Context, organizationID, projectID uuid.UUID) error
}
