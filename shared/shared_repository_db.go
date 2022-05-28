package shared

import (
	"context"
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed migrations
var migrations embed.FS

var (
	OrganizationIDSample uuid.UUID
	ProjectIDSample      uuid.UUID
	UserIDAdminSample    uuid.UUID
	ConfirmationIDError  uuid.UUID
	ConfirmationIdSample uuid.UUID
)

func init() {
	OrganizationIDSample = uuid.MustParse("4ed0c11d-3d6a-41c1-9873-558e86084591")
	ProjectIDSample = uuid.MustParse("f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea")
	UserIDAdminSample = uuid.MustParse("eeeeeb80-33f3-4d3f-befe-58694d2ac841")
	ConfirmationIDError = uuid.MustParse("4303e5b2-8124-4d9c-aea1-91aae2be562f")
}

// DbUserRepository is a SQL database repository for users
type DbRepositoryTxer struct {
	connPool *pgxpool.Pool
}

var _ RepositoryTxer = (*DbRepositoryTxer)(nil)

func NewDbRepositoryTxer(connPool *pgxpool.Pool) *DbRepositoryTxer {
	return &DbRepositoryTxer{
		connPool: connPool,
	}
}

func (txer *DbRepositoryTxer) InTx(ctx context.Context, txFuncs ...func(ctxWithTx context.Context) error) error {
	tx, err := txer.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	ctxWithTx := context.WithValue(ctx, ContextKeyTx, tx)

	for _, txFunc := range txFuncs {
		err = txFunc(ctxWithTx)
		if err != nil {
			rb := tx.Rollback(ctx)
			if rb != nil {
				return errors.Wrap(rb, "rollback error")
			}
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
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

func SetupTestDatabase(ctx context.Context) (testcontainers.Container, *pgxpool.Pool, error) {
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

	connPool, err := Connect(dbURI, 1)
	if err != nil {
		return nil, nil, err
	}

	err = insertSampleContent(ctx, connPool)

	return dbContainer, connPool, err
}

func Connect(dbURL string, maxConns int32) (*pgxpool.Pool, error) {
	err := migrateDb(dbURL)
	if err != nil {
		return nil, err
	}

	pgxConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	pgxConfig.LazyConnect = true
	pgxConfig.MaxConns = maxConns

	conn, err := pgxpool.ConnectConfig(context.Background(), pgxConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func migrateDb(dbURL string) error {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, strings.Replace(dbURL, "postgres://", "pgx://", 1))
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		return err
	}

	log.Printf("running database version %v (dirty: %v)", version, dirty)
	return nil
}
