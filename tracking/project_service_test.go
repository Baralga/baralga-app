package tracking

import (
	"context"
	"testing"

	"github.com/baralga/shared"
	"github.com/matryer/is"
)

func TestArchiveProject(t *testing.T) {
	// Arrange
	is := is.New(t)

	projectRepository := NewInMemProjectRepository()
	a := &ProjectService{
		app:               &shared.App{},
		repositoryTxer:    shared.NewInMemRepositoryTxer(),
		projectRepository: projectRepository,
	}

	// Act
	err := a.ArchiveProject(context.Background(), shared.OrganizationIDSample, shared.ProjectIDSample)

	// Assert
	is.NoErr(err)
	is.Equal(projectRepository.projects[0].Active, false)
}
