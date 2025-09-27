package tracking

import (
	"context"
	"strings"
	"time"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// InMemTagRepository is an in-memory repository for tags
type InMemTagRepository struct {
	tags         []*Tag
	activityTags map[uuid.UUID][]uuid.UUID // activityID -> []tagID
	tagService   *TagService
}

var _ TagRepository = (*InMemTagRepository)(nil)

// NewInMemTagRepository creates a new in-memory repository for tags
func NewInMemTagRepository() *InMemTagRepository {
	// Create some sample tags
	sampleTags := []*Tag{
		{
			ID:             uuid.MustParse("00000000-0000-0000-3333-000000000001"),
			Name:           "development",
			Color:          "#28a745",
			OrganizationID: shared.OrganizationIDSample,
			CreatedAt:      time.Now().Add(-24 * time.Hour),
		},
		{
			ID:             uuid.MustParse("00000000-0000-0000-3333-000000000002"),
			Name:           "meeting",
			Color:          "#007bff",
			OrganizationID: shared.OrganizationIDSample,
			CreatedAt:      time.Now().Add(-12 * time.Hour),
		},
		{
			ID:             uuid.MustParse("00000000-0000-0000-3333-000000000003"),
			Name:           "testing",
			Color:          "#dc3545",
			OrganizationID: shared.OrganizationIDSample,
			CreatedAt:      time.Now().Add(-6 * time.Hour),
		},
	}

	return &InMemTagRepository{
		tags:         sampleTags,
		activityTags: make(map[uuid.UUID][]uuid.UUID),
	}
}

// FindTagsByOrganization returns all tags for a specific organization for autocomplete
func (r *InMemTagRepository) FindTagsByOrganization(ctx context.Context, organizationID uuid.UUID, query string) ([]*Tag, error) {
	var matchingTags []*Tag
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	for _, tag := range r.tags {
		if tag.OrganizationID != organizationID {
			continue
		}

		// If no query, include all tags for the organization
		if normalizedQuery == "" {
			matchingTags = append(matchingTags, tag)
			continue
		}

		// Simple substring matching for in-memory implementation
		if strings.Contains(tag.Name, normalizedQuery) {
			matchingTags = append(matchingTags, tag)
		}
	}

	// Limit results to 20 for consistency with database implementation
	if len(matchingTags) > 20 {
		matchingTags = matchingTags[:20]
	}

	return matchingTags, nil
}

// FindOrCreateTag gets existing or creates new tag for organization
func (r *InMemTagRepository) FindOrCreateTag(ctx context.Context, name string, organizationID uuid.UUID) (*Tag, error) {
	// Normalize tag name to lowercase for case-insensitive handling
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return nil, errors.New("tag name cannot be empty")
	}

	// First try to find existing tag
	for _, tag := range r.tags {
		if tag.Name == normalizedName && tag.OrganizationID == organizationID {
			return tag, nil
		}
	}

	// Tag doesn't exist, create it with default color
	newTag := &Tag{
		ID:             uuid.New(),
		Name:           normalizedName,
		Color:          "#6c757d", // Default gray color
		OrganizationID: organizationID,
		CreatedAt:      time.Now(),
	}

	r.tags = append(r.tags, newTag)
	return newTag, nil
}

// SyncTagsForActivity creates/updates tag relationships when activity is saved
func (r *InMemTagRepository) SyncTagsForActivity(ctx context.Context, activityID uuid.UUID, organizationID uuid.UUID, tagNames []string) error {
	// First, clear existing relationships for this activity
	delete(r.activityTags, activityID)

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
	var tagIDs []uuid.UUID
	for tagName := range normalizedTags {
		// Generate color for new tags (existing tags will keep their color)
		color := r.tagService.GetTagColor(tagName)
		tag, err := r.findOrCreateTagWithColor(ctx, tagName, organizationID, color)
		if err != nil {
			return err
		}
		tagIDs = append(tagIDs, tag.ID)
	}

	// Store the relationships
	if len(tagIDs) > 0 {
		r.activityTags[activityID] = tagIDs
	}

	return nil
}

// findOrCreateTagWithColor gets existing or creates new tag for organization with specified color
func (r *InMemTagRepository) findOrCreateTagWithColor(ctx context.Context, name string, organizationID uuid.UUID, color string) (*Tag, error) {
	// Normalize tag name to lowercase for case-insensitive handling
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return nil, errors.New("tag name cannot be empty")
	}

	// First try to find existing tag
	for _, tag := range r.tags {
		if tag.Name == normalizedName && tag.OrganizationID == organizationID {
			return tag, nil
		}
	}

	// Tag doesn't exist, create it with the provided color
	newTag := &Tag{
		ID:             uuid.New(),
		Name:           normalizedName,
		Color:          color,
		OrganizationID: organizationID,
		CreatedAt:      time.Now(),
	}

	r.tags = append(r.tags, newTag)
	return newTag, nil
}

// DeleteUnusedTags cleanup method for organization-level cleanup
func (r *InMemTagRepository) DeleteUnusedTags(ctx context.Context, organizationID uuid.UUID) error {
	// Collect all tag IDs that are currently in use
	usedTagIDs := make(map[uuid.UUID]bool)
	for _, tagIDs := range r.activityTags {
		for _, tagID := range tagIDs {
			usedTagIDs[tagID] = true
		}
	}

	// Remove unused tags for the organization
	var remainingTags []*Tag
	for _, tag := range r.tags {
		if tag.OrganizationID == organizationID {
			// Keep tag if it's used or belongs to a different organization
			if usedTagIDs[tag.ID] {
				remainingTags = append(remainingTags, tag)
			}
		} else {
			// Keep tags from other organizations
			remainingTags = append(remainingTags, tag)
		}
	}

	r.tags = remainingTags
	return nil
}

// SetTagService sets the tag service for color generation
func (r *InMemTagRepository) SetTagService(tagService *TagService) {
	r.tagService = tagService
}
