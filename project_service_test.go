package main

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestArchiveProject(t *testing.T) {
	// Arrange
	is := is.New(t)

	projectRepository := NewInMemProjectRepository()
	a := &app{
		ProjectRepository: projectRepository,
		RepositoryTxer:    NewInMemRepositoryTxer(),
	}

	// Act
	err := a.ArchiveProject(context.Background(), organizationIDSample, projectIDSample)

	// Assert
	is.NoErr(err)
	is.Equal(projectRepository.projects[0].Active, false)
}
