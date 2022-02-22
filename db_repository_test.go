package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	organizationIDSample uuid.UUID
	projectIDSample      uuid.UUID
	userIDAdminSample    uuid.UUID
	confirmationIDError  uuid.UUID
	confirmationIdSample uuid.UUID
)

func init() {
	organizationIDSample = uuid.MustParse("4ed0c11d-3d6a-41c1-9873-558e86084591")
	projectIDSample = uuid.MustParse("f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea")
	userIDAdminSample = uuid.MustParse("eeeeeb80-33f3-4d3f-befe-58694d2ac841")
	confirmationIDError = uuid.MustParse("4303e5b2-8124-4d9c-aea1-91aae2be562f")
	confirmationIdSample = uuid.MustParse("eeeeeb80-33f3-4d3f-befe-58694d2ac841")
}

func insertSampleContent(ctx context.Context, connPool *pgxpool.Pool) error {
	_, err := connPool.Exec(
		ctx,
		`INSERT INTO activities 
		  (activity_id, start_time, end_time, description, project_id, org_id, username) 
	    VALUES 
		  ('2a52852c-3f36-11ec-9bbc-0242ac130002', '2021-10-14 14:00:00-00', '2021-10-14 14:10:00-00', 'My Desc', 'f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea', '4ed0c11d-3d6a-41c1-9873-558e86084591', 'admin')`,
	)
	return err
}

func setupDatabase(ctx context.Context) (testcontainers.Container, *pgxpool.Pool, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "baralga",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
		},
	}
	dbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}
	port, err := dbContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, err
	}
	host, err := dbContainer.Host(ctx)
	if err != nil {
		return nil, nil, err
	}

	dbURI := fmt.Sprintf("postgres://postgres:postgres@%v:%v/baralga", host, port.Port())

	connPool, err := connect(dbURI)
	if err != nil {
		return nil, nil, err
	}

	err = insertSampleContent(ctx, connPool)

	return dbContainer, connPool, err
}
