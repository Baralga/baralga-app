package tracking

import (
	"context"

	"github.com/baralga/shared"
	"github.com/baralga/shared/paged"
	"github.com/google/uuid"
)

type ProjectService struct {
	repositoryTxer    shared.RepositoryTxer
	projectRepository ProjectRepository
}

func NewProjectService(repositoryTxer shared.RepositoryTxer, projectRepository ProjectRepository) *ProjectService {
	return &ProjectService{
		repositoryTxer:    repositoryTxer,
		projectRepository: projectRepository,
	}
}

func (a *ProjectService) CreateProject(ctx context.Context, principal *shared.Principal, project *Project) (*Project, error) {
	project.ID = uuid.New()
	project.OrganizationID = principal.OrganizationID

	var projectCreated *Project
	err := a.repositoryTxer.InTx(
		context.Background(),
		func(ctx context.Context) error {
			a, err := a.projectRepository.InsertProject(ctx, project)
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

func (a *ProjectService) UpdateProject(ctx context.Context, organizationID uuid.UUID, project *Project) (*Project, error) {
	var projectUpdated *Project
	err := a.repositoryTxer.InTx(
		context.Background(),
		func(ctx context.Context) error {
			p, err := a.projectRepository.UpdateProject(ctx, organizationID, project)
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

func (a *ProjectService) ArchiveProject(ctx context.Context, organizationID, projectID uuid.UUID) error {
	err := a.repositoryTxer.InTx(
		context.Background(),
		func(ctx context.Context) error {
			err := a.projectRepository.ArchiveProjectByID(ctx, organizationID, projectID)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (a *ProjectService) DeleteProjectByID(ctx context.Context, principal *shared.Principal, projectID uuid.UUID) error {
	return a.repositoryTxer.InTx(
		ctx,
		func(ctx context.Context) error {
			return a.projectRepository.DeleteProjectByID(ctx, principal.OrganizationID, projectID)
		},
	)
}

func (a *ProjectService) ReadProjects(ctx context.Context, principal *shared.Principal, pageParams *paged.PageParams) (*ProjectsPaged, error) {
	return a.projectRepository.FindProjects(ctx, principal.OrganizationID, pageParams)
}

func (a *ProjectService) OrganizationInitializer() func(ctx context.Context, organizationID uuid.UUID) error {
	return func(ctx context.Context, organizationID uuid.UUID) error {
		// Create initial project
		project := &Project{
			ID:             uuid.New(),
			Title:          "My Project",
			Active:         true,
			OrganizationID: organizationID,
		}

		_, err := a.projectRepository.InsertProject(ctx, project)
		if err != nil {
			return err
		}
		return nil
	}
}
