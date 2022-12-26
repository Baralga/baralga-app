package tracking

import (
	"context"
	"database/sql"

	"github.com/baralga/shared"
	"github.com/baralga/shared/paged"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// DbProjectRepository is a SQL database repository for projects
type DbProjectRepository struct {
	connPool *pgxpool.Pool
}

var _ ProjectRepository = (*DbProjectRepository)(nil)

// NewDbProjectRepository creates a new SQL database repository for projects
func NewDbProjectRepository(connPool *pgxpool.Pool) *DbProjectRepository {
	return &DbProjectRepository{
		connPool: connPool,
	}
}

func (r *DbProjectRepository) FindProjects(ctx context.Context, organizationID uuid.UUID, pageParams *paged.PageParams) (*ProjectsPaged, error) {
	rows, err := r.connPool.Query(
		ctx,
		`SELECT project_id as id, title, description, active 
		 FROM projects 
		 WHERE org_id = $1 AND active = true
		 ORDER BY title ASC 
		 LIMIT $2 OFFSET $3`,
		organizationID, pageParams.Size, pageParams.Offset(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var (
			id          string
			title       string
			description sql.NullString
			active      bool
		)

		err = rows.Scan(&id, &title, &description, &active)
		if err != nil {
			return nil, err
		}

		project := &Project{
			ID:          uuid.MustParse(id),
			Title:       title,
			Description: description.String,
			Active:      active,
		}
		projects = append(projects, project)
	}

	row := r.connPool.QueryRow(
		ctx,
		`SELECT count(*) as total 
		 FROM projects 
		 WHERE org_id = $1 AND active = true`,
		organizationID,
	)
	var total int
	err = row.Scan(&total)
	if err != nil {
		return nil, err
	}

	projectsPaged := &ProjectsPaged{
		Projects: projects,
		Page:     pageParams.PageOfTotal(total),
	}

	return projectsPaged, nil
}

func (r *DbProjectRepository) FindProjectsByIDs(ctx context.Context, organizationID uuid.UUID, projectIDs []uuid.UUID) ([]*Project, error) {
	rows, err := r.connPool.Query(
		ctx,
		`SELECT project_id as id, title, description, active 
		 FROM projects 
		 WHERE org_id = $1 AND project_id = any($2) 
		 ORDER by title ASC`,
		organizationID, projectIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var (
			id          string
			title       string
			description sql.NullString
			active      bool
		)

		err = rows.Scan(&id, &title, &description, &active)
		if err != nil {
			return nil, err
		}

		project := &Project{
			ID:          uuid.MustParse(id),
			Title:       title,
			Description: description.String,
			Active:      active,
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (r *DbProjectRepository) FindProjectByID(ctx context.Context, organizationID, projectID uuid.UUID) (*Project, error) {
	row := r.connPool.QueryRow(ctx,
		`SELECT project_id as id, title, description, active  
         FROM projects 
	     WHERE project_id = $1 AND org_id = $2`,
		projectID, organizationID)

	var (
		id          string
		title       string
		description sql.NullString
		active      bool
	)

	err := row.Scan(&id, &title, &description, &active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}

		return nil, err
	}

	project := &Project{
		ID:          uuid.MustParse(id),
		Title:       title,
		Description: description.String,
		Active:      active,
	}

	return project, nil
}

func (r *DbProjectRepository) InsertProject(ctx context.Context, project *Project) (*Project, error) {
	tx := ctx.Value(shared.ContextKeyTx).(pgx.Tx)

	_, err := tx.Exec(
		ctx,
		`INSERT INTO projects 
		   (project_id, title, active, description, org_id) 
		 VALUES 
		   ($1, $2, $3, $4, $5)`,
		project.ID,
		project.Title,
		project.Active,
		project.Description,
		project.OrganizationID,
	)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *DbProjectRepository) UpdateProject(ctx context.Context, organizationID uuid.UUID, project *Project) (*Project, error) {
	tx := ctx.Value(shared.ContextKeyTx).(pgx.Tx)

	row := tx.QueryRow(ctx,
		`UPDATE projects 
		 SET title = $3, description = $4, active = $5 
		 WHERE project_id = $1 AND org_id = $2
		 RETURNING project_id`,
		project.ID, organizationID,
		project.Title, project.Description, project.Active,
	)

	var id string
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}

		return nil, err
	}

	return project, nil
}

func (r *DbProjectRepository) DeleteProjectByID(ctx context.Context, organizationID, projectID uuid.UUID) error {
	tx := ctx.Value(shared.ContextKeyTx).(pgx.Tx)

	_, err := tx.Exec(
		ctx,
		`DELETE FROM activities
		 WHERE project_id = $1 AND org_id = $2`,
		projectID, organizationID,
	)
	if err != nil {
		return err
	}

	row := tx.QueryRow(ctx,
		`DELETE 
         FROM projects 
	     WHERE project_id = $1 AND org_id = $2
		 RETURNING project_id`,
		projectID, organizationID)

	var id string
	err = row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProjectNotFound
		}

		return err
	}

	if id != projectID.String() {
		return ErrProjectNotFound
	}

	return nil
}

func (r *DbProjectRepository) ArchiveProjectByID(ctx context.Context, organizationID, projectID uuid.UUID) error {
	tx := ctx.Value(shared.ContextKeyTx).(pgx.Tx)

	row := tx.QueryRow(ctx,
		`UPDATE projects 
		 SET active = false 
		 WHERE project_id = $1 AND org_id = $2
		 RETURNING project_id`,
		projectID, organizationID)

	var id string
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProjectNotFound
		}

		return err
	}

	if id != projectID.String() {
		return ErrProjectNotFound
	}

	return nil
}
