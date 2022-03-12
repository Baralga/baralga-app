package main

import (
	"context"
	"testing"

	"github.com/baralga/paged"
	"github.com/google/uuid"
	"github.com/matryer/is"
	"github.com/pkg/errors"
)

func TestProjectRepository(t *testing.T) {
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

	projectRepository := NewDbProjectRepository(connPool)
	repositoryTxer := NewDbRepositoryTxer(connPool)

	t.Run("FindProjects", func(t *testing.T) {
		projectsPage, err := projectRepository.FindProjects(
			context.Background(),
			organizationIDSample,
			&paged.PageParams{
				Page: 0,
				Size: 50,
			},
		)

		is.NoErr(err)
		is.Equal(len(projectsPage.Projects), 1)
		is.Equal(projectsPage.Page.TotalElements, 1)
		is.True(projectsPage != nil)
	})

	t.Run("FindProjectsByIDs", func(t *testing.T) {
		projects, err := projectRepository.FindProjectsByIDs(
			context.Background(),
			organizationIDSample,
			[]uuid.UUID{projectIDSample},
		)

		is.NoErr(err)
		is.Equal(len(projects), 1)
	})

	t.Run("InsertAndUpdateProject", func(t *testing.T) {
		project := &Project{
			ID:             uuid.New(),
			Title:          "My Title",
			OrganizationID: organizationIDSample,
			Description:    "My Description",
		}

		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := projectRepository.InsertProject(
					ctx,
					project,
				)
				return err
			},
		)
		is.NoErr(err)

		project.Description = "My updated Description"

		var projectUpdate *Project
		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				p, err := projectRepository.UpdateProject(ctx, organizationIDSample, project)
				if err != nil {
					return err
				}
				projectUpdate = p
				return nil
			},
		)

		is.NoErr(err)
		is.Equal("My updated Description", projectUpdate.Description)
	})

	t.Run("ArchiveProject", func(t *testing.T) {
		// Arrange
		project := &Project{
			ID:             uuid.New(),
			Title:          "My Title",
			OrganizationID: organizationIDSample,
			Description:    "My Description",
			Active:         true,
		}

		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				_, err := projectRepository.InsertProject(
					ctx,
					project,
				)
				return err
			},
		)
		is.NoErr(err)

		// Act
		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				err := projectRepository.ArchiveProjectByID(ctx, organizationIDSample, project.ID)
				if err != nil {
					return err
				}
				return nil
			},
		)

		// Assert
		is.NoErr(err)
	})
}

func TestProjectRepositoryDeleteProject(t *testing.T) {
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

	projectRepository := NewDbProjectRepository(connPool)
	repositoryTxer := NewDbRepositoryTxer(connPool)

	t.Run("FindProjectByID", func(t *testing.T) {
		project, err := projectRepository.FindProjectByID(
			context.Background(),
			organizationIDSample,
			projectIDSample,
		)

		is.NoErr(err)
		is.Equal(projectIDSample, project.ID)
	})

	t.Run("FindNonExistingProjectByID", func(t *testing.T) {
		_, err := projectRepository.FindProjectByID(
			context.Background(),
			organizationIDSample,
			uuid.MustParse("f8d8a2ac-3f3e-11ec-9bbc-0242ac130002"),
		)

		is.True(errors.Is(err, ErrProjectNotFound))
	})

	t.Run("DeleteNonExistingProject", func(t *testing.T) {
		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				return projectRepository.DeleteProjectByID(
					ctx,
					organizationIDSample,
					uuid.MustParse("f8d8a2ac-3f3e-11ec-9bbc-0242ac130002"),
				)
			},
		)
		is.True(errors.Is(err, ErrProjectNotFound))
	})

	t.Run("DeleteExistingProject", func(t *testing.T) {
		err = repositoryTxer.InTx(
			context.Background(),
			func(ctx context.Context) error {
				return projectRepository.DeleteProjectByID(
					ctx,
					organizationIDSample,
					projectIDSample,
				)
			},
		)

		is.NoErr(err)
	})
}

type InMemProjectRepository struct {
	projects []*Project
}

var _ ProjectRepository = (*InMemProjectRepository)(nil)

func NewInMemProjectRepository() *InMemProjectRepository {
	return &InMemProjectRepository{
		projects: []*Project{
			{
				ID:             projectIDSample,
				Title:          "My Project",
				OrganizationID: organizationIDSample,
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
