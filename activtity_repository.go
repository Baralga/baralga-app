package main

import (
	"context"
	"fmt"
	"time"

	"github.com/baralga/paged"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

var ErrActivityNotFound = errors.New("activity not found")

type ActivitiesPaged struct {
	Activities []*Activity
	Page       *paged.Page
}

type ActivitiesFilter struct {
	Start          time.Time
	End            time.Time
	Username       string
	OrganizationID uuid.UUID
}

type ActivityRepository interface {
	FindActivities(ctx context.Context, filter *ActivitiesFilter, pageParams *paged.PageParams) (*ActivitiesPaged, error)
	InsertActivity(ctx context.Context, activity *Activity) (*Activity, error)
	FindActivityByID(ctx context.Context, activityID uuid.UUID, organizationID uuid.UUID) (*Activity, error)
	DeleteActivityByID(ctx context.Context, organizationID, activityID uuid.UUID) error
	DeleteActivityByIDAndUsername(ctx context.Context, organizationID, activityID uuid.UUID, username string) error
	UpdateActivity(ctx context.Context, organizationID uuid.UUID, activity *Activity) (*Activity, error)
	UpdateActivityByUsername(ctx context.Context, organizationID uuid.UUID, activity *Activity, username string) (*Activity, error)
}

// DbUserRepository is a SQL database repository for users
type DbActivityRepository struct {
	connPool *pgxpool.Pool
}

var _ ActivityRepository = (*DbActivityRepository)(nil)

// NewDbActivityRepository creates a new SQL database repository for users
func NewDbActivityRepository(connPool *pgxpool.Pool) *DbActivityRepository {
	return &DbActivityRepository{
		connPool: connPool,
	}
}

func (r *DbActivityRepository) FindActivities(ctx context.Context, filter *ActivitiesFilter, pageParams *paged.PageParams) (*ActivitiesPaged, error) {
	params := []interface{}{filter.OrganizationID, filter.Start, filter.End, pageParams.Size, pageParams.Offset()}
	filterSql := ""

	if filter.Username != "" {
		params = append(params, filter.Username)
		filterSql = " AND username = $6"
	}

	sql := fmt.Sprintf(
		`SELECT activity_id as id, description, start_time, end_time, username, org_id, project_id 
         FROM activities 
	     WHERE org_id = $1 %s AND $2 <= start_time AND start_time < $3
	     ORDER by start_time DESC 
	     LIMIT $4 OFFSET $5`,
		filterSql,
	)

	rows, err := r.connPool.Query(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*Activity
	for rows.Next() {
		var (
			id             string
			description    pgtype.Varchar
			startTime      time.Time
			endTime        time.Time
			username       string
			organizationID string
			projectID      string
		)

		err = rows.Scan(&id, &description, &startTime, &endTime, &username, &organizationID, &projectID)
		if err != nil {
			return nil, err
		}

		activity := &Activity{
			ID:             uuid.MustParse(id),
			Description:    description.String,
			Start:          startTime,
			End:            endTime,
			Username:       username,
			OrganizationID: uuid.MustParse(organizationID),
			ProjectID:      uuid.MustParse(projectID),
		}
		activities = append(activities, activity)
	}

	countParams := []interface{}{filter.OrganizationID, filter.Start, filter.End}
	countFilter := ""

	if filter.Username != "" {
		countParams = append(countParams, filter.Username)
		countFilter = " AND username = $4"
	}

	countSql := fmt.Sprintf(`
     	SELECT count(*) as total 
	    FROM activities
	    WHERE org_id = $1 %s AND $2 <= start_time AND start_time < $3`,
		countFilter)
	row := r.connPool.QueryRow(ctx, countSql, countParams...)
	var total int
	err = row.Scan(&total)
	if err != nil {
		return nil, err
	}

	actvtivitiesPaged := &ActivitiesPaged{
		Activities: activities,
		Page:       pageParams.PageOfTotal(total),
	}

	return actvtivitiesPaged, nil
}

func (r *DbActivityRepository) FindActivityByID(ctx context.Context, activityID, organizationID uuid.UUID) (*Activity, error) {
	row := r.connPool.QueryRow(ctx,
		`SELECT activity_id as id, description, start_time, end_time, username, org_id, project_id 
         FROM activities 
	     WHERE activity_id = $1 AND org_id = $2`,
		activityID, organizationID)

	var (
		id          string
		description pgtype.Varchar
		startTime   time.Time
		endTime     time.Time
		username    string
		orgID       string
		projectID   string
	)

	err := row.Scan(&id, &description, &startTime, &endTime, &username, &orgID, &projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrActivityNotFound
		}

		return nil, err
	}

	activity := &Activity{
		ID:             uuid.MustParse(id),
		Description:    description.String,
		Start:          startTime,
		End:            endTime,
		Username:       username,
		OrganizationID: uuid.MustParse(orgID),
		ProjectID:      uuid.MustParse(projectID),
	}

	return activity, nil
}

func (r *DbActivityRepository) DeleteActivityByID(ctx context.Context, organizationID, activityID uuid.UUID) error {
	row := r.connPool.QueryRow(ctx,
		`DELETE 
         FROM activities 
	     WHERE activity_id = $1 AND org_id = $2
		 RETURNING activity_id`,
		activityID, organizationID)

	var id string
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrActivityNotFound
		}

		return err
	}

	if id != activityID.String() {
		return ErrActivityNotFound
	}

	return nil
}

func (r *DbActivityRepository) DeleteActivityByIDAndUsername(ctx context.Context, organizationID, activityID uuid.UUID, username string) error {
	row := r.connPool.QueryRow(ctx,
		`DELETE 
         FROM activities 
	     WHERE activity_id = $1 AND org_id = $2 AND username = $3
		 RETURNING activity_id`,
		activityID, organizationID, username)

	var id string
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrActivityNotFound
		}

		return err
	}

	if id != activityID.String() {
		return ErrActivityNotFound
	}

	return nil
}

func (r *DbActivityRepository) UpdateActivity(ctx context.Context, organizationID uuid.UUID, activity *Activity) (*Activity, error) {
	row := r.connPool.QueryRow(ctx,
		`UPDATE activities 
		 SET start_time = $3, end_time = $4, description = $5, project_id = $6 
		 WHERE activity_id = $1 AND org_id = $2
		 RETURNING activity_id`,
		activity.ID, organizationID,
		activity.Start, activity.End, activity.Description, activity.ProjectID,
	)

	var id string
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrActivityNotFound
		}

		return nil, err
	}

	return activity, nil
}

func (r *DbActivityRepository) UpdateActivityByUsername(ctx context.Context, organizationID uuid.UUID, activity *Activity, username string) (*Activity, error) {
	row := r.connPool.QueryRow(ctx,
		`UPDATE activities 
		 SET start_time = $4, end_time = $5, description = $6, project_id = $7 
		 WHERE activity_id = $1 AND org_id = $2 AND username = $3
		 RETURNING activity_id`,
		activity.ID, organizationID, username,
		activity.Start, activity.End, activity.Description, activity.ProjectID,
	)

	var id string
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrActivityNotFound
		}

		return nil, err
	}

	return activity, nil
}

func (r *DbActivityRepository) InsertActivity(ctx context.Context, activity *Activity) (*Activity, error) {
	tx, err := r.connPool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO activities 
		   (activity_id, start_time, end_time, description, project_id, org_id, username) 
		 VALUES 
		   ($1, $2, $3, $4, $5, $6, $7)`,
		activity.ID,
		activity.Start,
		activity.End,
		activity.Description,
		activity.ProjectID,
		activity.OrganizationID,
		activity.Username,
	)
	if err != nil {
		rb := tx.Rollback(ctx)
		if rb != nil {
			return nil, errors.Wrap(rb, "rollback error")
		}
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return activity, nil
}
