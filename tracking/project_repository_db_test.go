package tracking

import (
	"context"
	"testing"

	"github.com/baralga/shared"
	"github.com/baralga/shared/paged"
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
	connPool, err := shared.SetupTestDatabase(ctx)
	if err != nil {
		t.Error(err)
	}

	projectRepository := NewDbProjectRepository(connPool)
	repositoryTxer := shared.NewDbRepositoryTxer(connPool)

	t.Run("FindProjects", func(t *testing.T) {
		projectsPage, err := projectRepository.FindProjects(
			context.Background(),
			shared.OrganizationIDSample,
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
			shared.OrganizationIDSample,
			[]uuid.UUID{shared.ProjectIDSample},
		)

		is.NoErr(err)
		is.Equal(len(projects), 1)
	})

	t.Run("InsertAndUpdateProject", func(t *testing.T) {
		project := &Project{
			ID:             uuid.New(),
			Title:          "My Title",
			OrganizationID: shared.OrganizationIDSample,
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
				p, err := projectRepository.UpdateProject(ctx, shared.OrganizationIDSample, project)
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
			OrganizationID: shared.OrganizationIDSample,
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
				err := projectRepository.ArchiveProjectByID(ctx, shared.OrganizationIDSample, project.ID)
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
	connPool, err := shared.SetupTestDatabase(ctx)
	if err != nil {
		t.Error(err)
	}

	projectRepository := NewDbProjectRepository(connPool)
	repositoryTxer := shared.NewDbRepositoryTxer(connPool)

	t.Run("FindProjectByID", func(t *testing.T) {
		project, err := projectRepository.FindProjectByID(
			context.Background(),
			shared.OrganizationIDSample,
			shared.ProjectIDSample,
		)

		is.NoErr(err)
		is.Equal(shared.ProjectIDSample, project.ID)
	})

	t.Run("FindNonExistingProjectByID", func(t *testing.T) {
		_, err := projectRepository.FindProjectByID(
			context.Background(),
			shared.OrganizationIDSample,
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
					shared.OrganizationIDSample,
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
					shared.OrganizationIDSample,
					shared.ProjectIDSample,
				)
			},
		)

		is.NoErr(err)
	})
}
