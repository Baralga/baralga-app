package tracking

import (
	"context"
	"strings"
	"time"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// DbTagRepository is a SQL database repository for tags
type DbTagRepository struct {
	connPool   *pgxpool.Pool
	tagService *TagService
}

var _ TagRepository = (*DbTagRepository)(nil)

// NewDbTagRepository creates a new SQL database repository for tags
func NewDbTagRepository(connPool *pgxpool.Pool) *DbTagRepository {
	return &DbTagRepository{
		connPool: connPool,
	}
}

// FindTagsByOrganization returns all tags for a specific organization for autocomplete
func (r *DbTagRepository) FindTagsByOrganization(ctx context.Context, organizationID uuid.UUID, query string) ([]*Tag, error) {
	var rows pgx.Rows
	var err error

	if query == "" {
		// Return all tags for the organization if no query provided
		rows, err = r.connPool.Query(ctx,
			`SELECT tag_id, name, color, org_id, created_at 
			 FROM tags 
			 WHERE org_id = $1 
			 ORDER BY name ASC`,
			organizationID)
	} else {
		// Use trigram similarity search for autocomplete
		normalizedQuery := strings.ToLower(strings.TrimSpace(query))
		rows, err = r.connPool.Query(ctx,
			`SELECT tag_id, name, color, org_id, created_at 
			 FROM tags 
			 WHERE org_id = $1 AND name % $2
			 ORDER BY similarity(name, $2) DESC, name ASC
			 LIMIT 20`,
			organizationID, normalizedQuery)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		var (
			tagID     string
			name      string
			color     string
			orgID     string
			createdAt time.Time
		)

		err = rows.Scan(&tagID, &name, &color, &orgID, &createdAt)
		if err != nil {
			return nil, err
		}

		tag := &Tag{
			ID:             uuid.MustParse(tagID),
			Name:           name,
			Color:          color,
			OrganizationID: uuid.MustParse(orgID),
			CreatedAt:      createdAt,
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// FindOrCreateTag gets existing or creates new tag for organization with default color
func (r *DbTagRepository) FindOrCreateTag(ctx context.Context, name string, organizationID uuid.UUID) (*Tag, error) {
	return r.FindOrCreateTagWithColor(ctx, name, organizationID, "#6c757d")
}

// FindOrCreateTagWithColor gets existing or creates new tag for organization with specified color
func (r *DbTagRepository) FindOrCreateTagWithColor(ctx context.Context, name string, organizationID uuid.UUID, color string) (*Tag, error) {
	// Normalize tag name to lowercase for case-insensitive handling
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return nil, errors.New("tag name cannot be empty")
	}

	tx := shared.MustTxFromContext(ctx)

	// First try to find existing tag
	row := tx.QueryRow(ctx,
		`SELECT tag_id, name, color, org_id, created_at 
		 FROM tags 
		 WHERE name = $1 AND org_id = $2`,
		normalizedName, organizationID)

	var (
		tagID     string
		tagName   string
		tagColor  string
		orgID     string
		createdAt time.Time
	)

	err := row.Scan(&tagID, &tagName, &tagColor, &orgID, &createdAt)
	if err == nil {
		// Tag exists, return it
		return &Tag{
			ID:             uuid.MustParse(tagID),
			Name:           tagName,
			Color:          tagColor,
			OrganizationID: uuid.MustParse(orgID),
			CreatedAt:      createdAt,
		}, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// Tag doesn't exist, create it
	newTagID := uuid.New()
	now := time.Now()

	_, err = tx.Exec(ctx,
		`INSERT INTO tags (tag_id, name, color, org_id, created_at) 
		 VALUES ($1, $2, $3, $4, $5)`,
		newTagID, normalizedName, color, organizationID, now)
	if err != nil {
		return nil, err
	}

	return &Tag{
		ID:             newTagID,
		Name:           normalizedName,
		Color:          color,
		OrganizationID: organizationID,
		CreatedAt:      now,
	}, nil
}

// SyncTagsForActivity creates/updates tag relationships when activity is saved
func (r *DbTagRepository) SyncTagsForActivity(ctx context.Context, activityID uuid.UUID, organizationID uuid.UUID, tagNames []string) error {
	tx := shared.MustTxFromContext(ctx)

	// First, delete all existing tag relationships for this activity
	_, err := tx.Exec(ctx,
		`DELETE FROM activity_tags WHERE activity_id = $1`,
		activityID)
	if err != nil {
		return err
	}

	// If no tags provided, we're done
	if len(tagNames) == 0 {
		return nil
	}

	// Normalize and deduplicate tag names
	normalizedTags := make(map[string]bool)
	for _, tagName := range tagNames {
		normalized := strings.ToLower(strings.TrimSpace(tagName))
		if normalized != "" {
			normalizedTags[normalized] = true
		}
	}

	// Create or find each tag and create the relationship
	for tagName := range normalizedTags {
		// Generate color for new tags (existing tags will keep their color)
		color := r.tagService.GetTagColor(tagName)
		tag, err := r.FindOrCreateTagWithColor(ctx, tagName, organizationID, color)
		if err != nil {
			return err
		}

		// Create the activity-tag relationship
		_, err = tx.Exec(ctx,
			`INSERT INTO activity_tags (activity_id, tag_id, org_id) 
			 VALUES ($1, $2, $3)`,
			activityID, tag.ID, organizationID)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteUnusedTags cleanup method for organization-level cleanup
func (r *DbTagRepository) DeleteUnusedTags(ctx context.Context, organizationID uuid.UUID) error {
	tx := shared.MustTxFromContext(ctx)

	// Delete tags that are not referenced by any activities in the organization
	_, err := tx.Exec(ctx,
		`DELETE FROM tags 
		 WHERE org_id = $1 
		 AND tag_id NOT IN (
			 SELECT DISTINCT tag_id 
			 FROM activity_tags 
			 WHERE org_id = $1
		 )`,
		organizationID)

	return err
}

// SetTagService sets the tag service for color generation
func (r *DbTagRepository) SetTagService(tagService *TagService) {
	r.tagService = tagService
}
