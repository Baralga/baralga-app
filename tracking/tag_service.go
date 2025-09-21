package tracking

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrTooManyTags is returned when more than 10 tags are provided
	ErrTooManyTags = errors.New("maximum 10 tags allowed per activity")
	// ErrTagTooLong is returned when a tag exceeds the maximum length
	ErrTagTooLong = errors.New("tag name cannot exceed 50 characters")
	// ErrInvalidTagFormat is returned when a tag contains invalid characters
	ErrInvalidTagFormat = errors.New("tag name can only contain alphanumeric characters, hyphens, and underscores")
	// ErrTagTooShort is returned when a tag is too short
	ErrTagTooShort = errors.New("tag name must be at least 1 character long")
)

// TagService handles business logic for tag operations
type TagService struct {
	tagRepository TagRepository
}

// NewTagService creates a new TagService instance
func NewTagService(tagRepository TagRepository) *TagService {
	return &TagService{
		tagRepository: tagRepository,
	}
}

// GetTagsForAutocomplete returns matching tags for autocomplete within organization
// Performs fuzzy matching on tag names using the provided query string
func (s *TagService) GetTagsForAutocomplete(ctx context.Context, organizationID uuid.UUID, query string) ([]*Tag, error) {
	if query == "" {
		return []*Tag{}, nil
	}

	// Normalize the query for consistent matching
	normalizedQuery := s.NormalizeTagName(query)

	return s.tagRepository.FindTagsByOrganization(ctx, organizationID, normalizedQuery)
}

// NormalizeTagName converts tags to lowercase for case-insensitive handling
// This ensures consistent storage and matching of tag names
func (s *TagService) NormalizeTagName(tagName string) string {
	return strings.ToLower(strings.TrimSpace(tagName))
}

// ParseTagsFromString splits comma/space separated tags into a slice
// Handles both comma and space separation, removes duplicates and empty strings
func (s *TagService) ParseTagsFromString(tagString string) []string {
	if tagString == "" {
		return []string{}
	}

	// Split by comma first, then by spaces within each part
	var tags []string
	parts := strings.Split(tagString, ",")

	for _, part := range parts {
		// Split each part by spaces and add non-empty tags
		spaceSplit := strings.Fields(strings.TrimSpace(part))
		for _, tag := range spaceSplit {
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// Remove duplicates and normalize
	return s.removeDuplicates(tags)
}

// ValidateTags validates tag names and enforces business rules
// Checks for maximum count (10), length limits (1-50 chars), and valid format
func (s *TagService) ValidateTags(tags []string) error {
	// Check maximum tag count
	if len(tags) > 10 {
		return ErrTooManyTags
	}

	// Validate each tag
	tagPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	for _, tag := range tags {
		normalizedTag := s.NormalizeTagName(tag)

		// Check minimum length
		if len(normalizedTag) < 1 {
			return ErrTagTooShort
		}

		// Check maximum length
		if len(normalizedTag) > 50 {
			return ErrTagTooLong
		}

		// Check format (alphanumeric, hyphens, underscores only)
		if !tagPattern.MatchString(normalizedTag) {
			return fmt.Errorf("%w: '%s'", ErrInvalidTagFormat, tag)
		}
	}

	return nil
}

// GetTagColor generates a consistent color for a tag based on its name
// Uses SHA256 hash of the normalized tag name to ensure consistency
func (s *TagService) GetTagColor(tagName string) string {
	// Predefined set of accessible colors with good contrast
	colors := []string{
		"#007bff", // Blue
		"#28a745", // Green
		"#dc3545", // Red
		"#ffc107", // Yellow
		"#17a2b8", // Cyan
		"#6f42c1", // Purple
		"#fd7e14", // Orange
		"#20c997", // Teal
		"#e83e8c", // Pink
		"#6c757d", // Gray
		"#343a40", // Dark
		"#495057", // Dark Gray
		"#0056b3", // Dark Blue
		"#155724", // Dark Green
		"#721c24", // Dark Red
		"#856404", // Dark Yellow
	}

	// Default fallback color
	if tagName == "" {
		return "#6c757d"
	}

	// Normalize tag name for consistent hashing
	normalized := s.NormalizeTagName(tagName)

	// Generate hash
	hash := sha256.Sum256([]byte(normalized))

	// Use first byte of hash to select color
	colorIndex := int(hash[0]) % len(colors)

	selectedColor := colors[colorIndex]

	// Safety check: ensure color is exactly 7 characters (#RRGGBB format)
	if len(selectedColor) != 7 || selectedColor[0] != '#' {
		return "#6c757d" // Return default gray if invalid color
	}

	return selectedColor
}

// removeDuplicates removes duplicate tags and normalizes them
func (s *TagService) removeDuplicates(tags []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, tag := range tags {
		normalized := s.NormalizeTagName(tag)
		if normalized != "" && !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	return result
}

// PrepareTagsWithColors takes tag names and returns Tag objects with generated colors
// This method should be called before SyncTagsForActivity to ensure all tags have colors
func (s *TagService) PrepareTagsWithColors(tagNames []string) []*Tag {
	if len(tagNames) == 0 {
		return []*Tag{}
	}

	// Normalize and deduplicate tag names
	normalizedTags := make(map[string]bool)
	for _, tagName := range tagNames {
		normalized := s.NormalizeTagName(tagName)
		if normalized != "" {
			normalizedTags[normalized] = true
		}
	}

	// Create Tag objects with generated colors
	var tags []*Tag
	for tagName := range normalizedTags {
		tag := &Tag{
			Name:  tagName,
			Color: s.GetTagColor(tagName),
		}
		tags = append(tags, tag)
	}

	return tags
}

// TagReportData represents comprehensive tag report data with filtering options
type TagReportData struct {
	Items        []*TagReportItem
	SelectedTags []string
	TotalTags    int
	TotalTime    int // in minutes
}

// GenerateTagReports generates comprehensive tag-based reports with time breakdowns
// Supports filtering by specific tags and aggregation by time periods
func (s *TagService) GenerateTagReports(ctx context.Context, organizationID uuid.UUID, filter *ActivitiesFilter, aggregateBy string, selectedTags []string) (*TagReportData, error) {
	// Create a copy of the filter with selected tags if provided
	reportFilter := &ActivitiesFilter{
		Start:          filter.Start,
		End:            filter.End,
		SortBy:         filter.SortBy,
		SortOrder:      filter.SortOrder,
		Username:       filter.Username,
		OrganizationID: organizationID,
		Tags:           selectedTags, // Use selected tags for filtering
	}

	// Get tag report data based on aggregation type
	items, err := s.GetTagReportData(ctx, reportFilter, aggregateBy)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	totalTime := 0
	tagSet := make(map[string]bool)
	for _, item := range items {
		totalTime += item.DurationInMinutesTotal
		tagSet[item.TagName] = true
	}

	return &TagReportData{
		Items:        items,
		SelectedTags: selectedTags,
		TotalTags:    len(tagSet),
		TotalTime:    totalTime,
	}, nil
}

// GetTagReportData retrieves filtered tag report data for specific date ranges and tag selections
// Handles activities with multiple tags by counting activity time for each associated tag
func (s *TagService) GetTagReportData(ctx context.Context, filter *ActivitiesFilter, aggregateBy string) ([]*TagReportItem, error) {
	// This method will need to be implemented by calling the repository layer
	// For now, we'll define the interface that the repository should implement

	// The repository implementation should:
	// 1. Join activities with their tags
	// 2. Filter by date range, organization, username, and selected tags
	// 3. Group by tag and time period (day/week/month/quarter)
	// 4. Calculate total duration and activity count per tag per time period
	// 5. Handle activities with multiple tags by including the activity's time for each tag

	return s.tagRepository.GetTagReportData(ctx, filter, aggregateBy)
}
