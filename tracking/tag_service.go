package tracking

import (
	"context"
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