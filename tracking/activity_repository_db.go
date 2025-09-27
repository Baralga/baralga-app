package tracking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/baralga/shared"
	"github.com/baralga/shared/paged"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

// DbActivityRepository is a SQL database repository for activities
type DbActivityRepository struct {
	connPool *pgxpool.Pool
}

var _ ActivityRepository = (*DbActivityRepository)(nil)

// NewDbActivityRepository creates a new SQL database repository for activities
func NewDbActivityRepository(connPool *pgxpool.Pool) *DbActivityRepository {
	return &DbActivityRepository{
		connPool: connPool,
	}
}

func (r *DbActivityRepository) TimeReportByDay(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error) {
	params := []interface{}{filter.OrganizationID, filter.Start, filter.End}
	filterSql := ""

	if filter.Username != "" {
		params = append(params, filter.Username)
		filterSql = " AND username = $4"
	}

	sql := fmt.Sprintf(
		`SELECT year, quarter, month, week, day, sum(duration_minutes_total) as duration_minutes_total  
		 FROM activities_agg
	     WHERE org_id = $1 AND $2 <= start_time AND start_time < $3 %s
		 GROUP BY year, quarter, month, week, day
         ORDER BY (year, quarter, month, week, day) desc`,
		filterSql,
	)

	rows, err := r.connPool.Query(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*ActivityTimeReportItem
	for rows.Next() {
		var (
			year              int
			quarter           int
			month             int
			week              int
			day               int
			durationInMinutes int
		)

		err = rows.Scan(&year, &quarter, &month, &week, &day, &durationInMinutes)
		if err != nil {
			return nil, err
		}

		activity := &ActivityTimeReportItem{
			Year:                   year,
			Quarter:                quarter,
			Month:                  month,
			Week:                   week,
			Day:                    day,
			DurationInMinutesTotal: durationInMinutes,
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func (r *DbActivityRepository) TimeReportByWeek(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error) {
	params := []interface{}{filter.OrganizationID, filter.Start, filter.End}
	filterSql := ""

	if filter.Username != "" {
		params = append(params, filter.Username)
		filterSql = " AND username = $4"
	}

	sql := fmt.Sprintf(
		`SELECT year, week, sum(duration_minutes_total) as duration_minutes_total  
		 FROM activities_agg
	     WHERE org_id = $1 AND $2 <= start_time AND start_time < $3 %s
		 GROUP BY year, week
         ORDER BY (year, week) desc`,
		filterSql,
	)

	rows, err := r.connPool.Query(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*ActivityTimeReportItem
	for rows.Next() {
		var (
			year              int
			week              int
			durationInMinutes int
		)

		err = rows.Scan(&year, &week, &durationInMinutes)
		if err != nil {
			return nil, err
		}

		activity := &ActivityTimeReportItem{
			Year:                   year,
			Week:                   week,
			DurationInMinutesTotal: durationInMinutes,
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func (r *DbActivityRepository) TimeReportByMonth(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error) {
	params := []interface{}{filter.OrganizationID, filter.Start, filter.End}
	filterSql := ""

	if filter.Username != "" {
		params = append(params, filter.Username)
		filterSql = " AND username = $4"
	}

	sql := fmt.Sprintf(
		`SELECT year, month, sum(duration_minutes_total) as duration_minutes_total  
		 FROM activities_agg
	     WHERE org_id = $1 AND $2 <= start_time AND start_time < $3 %s
		 GROUP BY year, month
         ORDER BY (year, month) desc`,
		filterSql,
	)

	rows, err := r.connPool.Query(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*ActivityTimeReportItem
	for rows.Next() {
		var (
			year              int
			month             int
			durationInMinutes int
		)

		err = rows.Scan(&year, &month, &durationInMinutes)
		if err != nil {
			return nil, err
		}

		activity := &ActivityTimeReportItem{
			Day:                    1,
			Year:                   year,
			Month:                  month,
			DurationInMinutesTotal: durationInMinutes,
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func (r *DbActivityRepository) TimeReportByQuarter(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error) {
	params := []interface{}{filter.OrganizationID, filter.Start, filter.End}
	filterSql := ""

	if filter.Username != "" {
		params = append(params, filter.Username)
		filterSql = " AND username = $4"
	}

	sql := fmt.Sprintf(
		`SELECT year, quarter, sum(duration_minutes_total) as duration_minutes_total  
		 FROM activities_agg
	     WHERE org_id = $1 AND $2 <= start_time AND start_time < $3 %s
		 GROUP BY year, quarter
         ORDER BY (year, quarter) desc`,
		filterSql,
	)

	rows, err := r.connPool.Query(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*ActivityTimeReportItem
	for rows.Next() {
		var (
			year              int
			quarter           int
			durationInMinutes int
		)

		err = rows.Scan(&year, &quarter, &durationInMinutes)
		if err != nil {
			return nil, err
		}

		activity := &ActivityTimeReportItem{
			Day:                    1,
			Year:                   year,
			Quarter:                quarter,
			DurationInMinutesTotal: durationInMinutes,
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func (r *DbActivityRepository) ProjectReport(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityProjectReportItem, error) {
	params := []interface{}{filter.OrganizationID, filter.Start, filter.End}
	filterSql := ""

	if filter.Username != "" {
		params = append(params, filter.Username)
		filterSql = " AND username = $4"
	}

	sql := fmt.Sprintf(
		`SELECT ag.project_id, projects.title as title, ag.duration_minutes_total FROM 
		  (SELECT project_id, sum(duration_minutes_total) as duration_minutes_total  
		   FROM activities_agg
	       WHERE org_id = $1 AND $2 <= start_time AND start_time < $3 %s
		   GROUP BY project_id
		  ) ag
		INNER JOIN projects
		ON projects.project_id = ag.project_id
		ORDER BY (title) asc`,
		filterSql,
	)

	rows, err := r.connPool.Query(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*ActivityProjectReportItem
	for rows.Next() {
		var (
			projectID         uuid.UUID
			projectTitle      string
			durationInMinutes int
		)

		err = rows.Scan(&projectID, &projectTitle, &durationInMinutes)
		if err != nil {
			return nil, err
		}

		activity := &ActivityProjectReportItem{
			ProjectID:              projectID,
			ProjectTitle:           projectTitle,
			DurationInMinutesTotal: durationInMinutes,
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func (r *DbActivityRepository) FindActivities(ctx context.Context, filter *ActivitiesFilter, pageParams *paged.PageParams) (*ActivitiesPaged, []*Project, error) {
	params := []interface{}{filter.OrganizationID, filter.Start, filter.End, pageParams.Size, pageParams.Offset()}
	filterSql := ""
	paramIndex := 6

	if filter.Username != "" {
		params = append(params, filter.Username)
		filterSql += fmt.Sprintf(" AND username = $%d", paramIndex)
		paramIndex++
	}

	// Add tag filtering support
	tagJoin := ""
	if len(filter.Tags) > 0 {
		tagJoin = `INNER JOIN activity_tags at ON a.id = at.activity_id
				   INNER JOIN tags t ON at.tag_id = t.tag_id`

		// Create placeholders for tag names
		tagPlaceholders := make([]string, len(filter.Tags))
		for i, tag := range filter.Tags {
			params = append(params, strings.ToLower(strings.TrimSpace(tag)))
			tagPlaceholders[i] = fmt.Sprintf("$%d", paramIndex)
			paramIndex++
		}
		filterSql += fmt.Sprintf(" AND t.name IN (%s)", strings.Join(tagPlaceholders, ","))
	}

	sortBy := "start"
	if filter.SortBy != "" {
		sortBy = strings.ToLower(filter.SortBy)
	}

	sortOrder := "DESC"
	if filter.SortOrder != "" {
		sortOrder = strings.ToUpper(filter.SortOrder)
	}

	sql := fmt.Sprintf(
		`SELECT DISTINCT a.*, projects.title as project FROM
		   (SELECT activity_id as id, description, start_time as start, end_time as end, username, org_id, project_id
			FROM activities 
			WHERE org_id = $1 %s AND $2 <= start_time AND start_time < $3
		   ) a
		 %s
         INNER JOIN projects
	     ON projects.project_id = a.project_id
		 ORDER by %s %s 
	     LIMIT $4 OFFSET $5`,
		filterSql,
		tagJoin,
		sortBy,
		sortOrder,
	)

	rows, err := r.connPool.Query(ctx, sql, params...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var activities []*Activity
	projectsById := make(map[uuid.UUID]*Project)
	for rows.Next() {
		var (
			id             string
			description    string
			startTime      time.Time
			endTime        time.Time
			username       string
			organizationID string
			projectID      string
			projectTitle   string
		)

		err = rows.Scan(&id, &description, &startTime, &endTime, &username, &organizationID, &projectID, &projectTitle)
		if err != nil {
			return nil, nil, err
		}

		projectUUID := uuid.MustParse(projectID)
		activityID := uuid.MustParse(id)

		activity := &Activity{
			ID:             activityID,
			Description:    description,
			Start:          startTime,
			End:            endTime,
			Username:       username,
			OrganizationID: uuid.MustParse(organizationID),
			ProjectID:      projectUUID,
		}
		activities = append(activities, activity)

		if _, ok := projectsById[projectUUID]; !ok {
			project := &Project{
				ID:             projectUUID,
				OrganizationID: uuid.MustParse(organizationID),
				Title:          projectTitle,
			}
			projectsById[projectUUID] = project
		}
	}

	// Load tags for all activities
	err = r.loadTagsForActivities(ctx, activities)
	if err != nil {
		return nil, nil, err
	}

	projects := maps.Values(projectsById)

	// Count query with same filtering logic
	countParams := []interface{}{filter.OrganizationID, filter.Start, filter.End}
	countFilter := ""
	countParamIndex := 4

	if filter.Username != "" {
		countParams = append(countParams, filter.Username)
		countFilter += fmt.Sprintf(" AND username = $%d", countParamIndex)
		countParamIndex++
	}

	countTagJoin := ""
	if len(filter.Tags) > 0 {
		countTagJoin = `INNER JOIN activity_tags at ON activities.activity_id = at.activity_id
						INNER JOIN tags t ON at.tag_id = t.tag_id`

		// Add tag parameters for count query
		tagPlaceholders := make([]string, len(filter.Tags))
		for i, tag := range filter.Tags {
			countParams = append(countParams, strings.ToLower(strings.TrimSpace(tag)))
			tagPlaceholders[i] = fmt.Sprintf("$%d", countParamIndex)
			countParamIndex++
		}
		countFilter += fmt.Sprintf(" AND t.name IN (%s)", strings.Join(tagPlaceholders, ","))
	}

	countSql := fmt.Sprintf(`
     	SELECT count(DISTINCT activities.activity_id) as total 
	    FROM activities
		%s
	    WHERE org_id = $1 %s AND $2 <= start_time AND start_time < $3`,
		countTagJoin, countFilter)
	row := r.connPool.QueryRow(ctx, countSql, countParams...)
	var total int
	err = row.Scan(&total)
	if err != nil {
		return nil, nil, err
	}

	actvtivitiesPaged := &ActivitiesPaged{
		Activities: activities,
		Page:       pageParams.PageOfTotal(total),
	}

	return actvtivitiesPaged, projects, nil
}

func (r *DbActivityRepository) FindActivityByID(ctx context.Context, activityID, organizationID uuid.UUID) (*Activity, error) {
	row := r.connPool.QueryRow(ctx,
		`SELECT activity_id as id, description, start_time, end_time, username, org_id, project_id 
         FROM activities 
	     WHERE activity_id = $1 AND org_id = $2`,
		activityID, organizationID)

	var (
		id          string
		description string
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
		Description:    description,
		Start:          startTime,
		End:            endTime,
		Username:       username,
		OrganizationID: uuid.MustParse(orgID),
		ProjectID:      uuid.MustParse(projectID),
	}

	// Load tags for this activity
	err = r.loadTagsForActivities(ctx, []*Activity{activity})
	if err != nil {
		return nil, err
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
	tx := shared.MustTxFromContext(ctx)

	row := tx.QueryRow(ctx,
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
	tx := shared.MustTxFromContext(ctx)

	row := tx.QueryRow(ctx,
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
	tx := shared.MustTxFromContext(ctx)

	row := tx.QueryRow(ctx,
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
	tx := shared.MustTxFromContext(ctx)

	_, err := tx.Exec(
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
		return nil, err
	}

	return activity, nil
}

// loadTagsForActivities loads tags for a slice of activities
func (r *DbActivityRepository) loadTagsForActivities(ctx context.Context, activities []*Activity) error {
	if len(activities) == 0 {
		return nil
	}

	// Create a map of activity IDs for quick lookup
	activityMap := make(map[uuid.UUID]*Activity)
	activityIDs := make([]interface{}, len(activities))
	placeholders := make([]string, len(activities))

	for i, activity := range activities {
		activityMap[activity.ID] = activity
		activityIDs[i] = activity.ID
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	// Query to get all tags for these activities
	sql := fmt.Sprintf(`
		SELECT at.activity_id, t.tag_id, t.name, t.color, t.org_id, t.created_at
		FROM activity_tags at
		INNER JOIN tags t ON at.tag_id = t.tag_id
		WHERE at.activity_id IN (%s)
		ORDER BY at.activity_id, t.name`,
		strings.Join(placeholders, ","))

	rows, err := r.connPool.Query(ctx, sql, activityIDs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Group tags by activity ID
	for rows.Next() {
		var (
			activityID string
			tagID      string
			tagName    string
			tagColor   string
			orgID      string
			createdAt  time.Time
		)

		err = rows.Scan(&activityID, &tagID, &tagName, &tagColor, &orgID, &createdAt)
		if err != nil {
			return err
		}

		activityUUID := uuid.MustParse(activityID)
		if activity, exists := activityMap[activityUUID]; exists {
			tag := &Tag{
				ID:             uuid.MustParse(tagID),
				Name:           tagName,
				Color:          tagColor,
				OrganizationID: uuid.MustParse(orgID),
				CreatedAt:      createdAt,
			}
			activity.Tags = append(activity.Tags, tag)
		}
	}

	return nil
}
