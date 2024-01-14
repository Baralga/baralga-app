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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pkg/errors"
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

// MustTxFromContext reads the current transaction from the context or panics if not present
func MustTxFromContext(ctx context.Context) pgx.Tx {
	tx, ok := ctx.Value(contextKeyTx).(pgx.Tx)
	if !ok {
		panic("no tx found in context")
	}

	return tx
}

func toContextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, contextKeyTx, tx)
}

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

	ctxWithTx := toContextWithTx(ctx, tx)

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

func SetupTestDatabase(ctx context.Context) (func() error, *pgxpool.Pool, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres", Tag: "14", Env: []string{"POSTGRES_DB=baralga", "POSTGRES_PASSWORD=postgres", "POSTGRES_USER=postgres", "listen_addresses = '*'"}},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	dbURI := fmt.Sprintf("postgres://postgres:postgres@localhost:%v/baralga", resource.GetPort("5432/tcp"))

	err = pool.Retry(func() error {
		var err error

		_, err = Connect(dbURI, 1)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	connPool, err := Connect(dbURI, 1)
	if err != nil {
		return nil, nil, err
	}

	err = insertSampleContent(ctx, connPool)

	return func() error {
		return pool.Purge(resource)
	}, connPool, err
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

	pgxConfig.MaxConns = maxConns

	conn, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
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
