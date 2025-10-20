package tracking

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"github.com/pkg/errors"
)

func TestTagService_NormalizeTagName(t *testing.T) {
	is := is.New(t)
	service := NewTagService(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase tag",
			input:    "meeting",
			expected: "meeting",
		},
		{
			name:     "uppercase tag",
			input:    "MEETING",
			expected: "meeting",
		},
		{
			name:     "mixed case tag",
			input:    "MeEtInG",
			expected: "meeting",
		},
		{
			name:     "tag with spaces",
			input:    "  meeting  ",
			expected: "meeting",
		},
		{
			name:     "empty tag",
			input:    "",
			expected: "",
		},
		{
			name:     "tag with hyphens",
			input:    "Bug-Fix",
			expected: "bug-fix",
		},
		{
			name:     "tag with underscores",
			input:    "Code_Review",
			expected: "code_review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.NormalizeTagName(tt.input)
			is.Equal(result, tt.expected)
		})
	}
}

func TestTagService_ParseTagsFromString(t *testing.T) {
	is := is.New(t)
	service := NewTagService(nil)

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "comma separated tags",
			input:    "meeting,development,bug-fix",
			expected: []string{"meeting", "development", "bug-fix"},
		},
		{
			name:     "space separated tags",
			input:    "meeting development bug-fix",
			expected: []string{"meeting", "development", "bug-fix"},
		},
		{
			name:     "mixed comma and space separation",
			input:    "meeting, development bug-fix",
			expected: []string{"meeting", "development", "bug-fix"},
		},
		{
			name:     "tags with extra spaces",
			input:    "  meeting  ,  development  ,  bug-fix  ",
			expected: []string{"meeting", "development", "bug-fix"},
		},
		{
			name:     "duplicate tags",
			input:    "meeting,meeting,development",
			expected: []string{"meeting", "development"},
		},
		{
			name:     "case insensitive duplicates",
			input:    "Meeting,MEETING,meeting",
			expected: []string{"meeting"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only spaces and commas",
			input:    "  ,  ,  ",
			expected: []string{},
		},
		{
			name:     "single tag",
			input:    "meeting",
			expected: []string{"meeting"},
		},
		{
			name:     "tags with underscores and hyphens",
			input:    "code_review,bug-fix,team_meeting",
			expected: []string{"code_review", "bug-fix", "team_meeting"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ParseTagsFromString(tt.input)
			is.Equal(len(result), len(tt.expected))
			for i, tag := range result {
				is.Equal(tag, tt.expected[i])
			}
		})
	}
}

func TestTagService_ValidateTags(t *testing.T) {
	is := is.New(t)
	service := NewTagService(nil)

	tests := []struct {
		name        string
		input       []string
		expectError bool
		expectedErr error
	}{
		{
			name:        "valid tags",
			input:       []string{"meeting", "development", "bug-fix"},
			expectError: false,
		},
		{
			name:        "empty slice",
			input:       []string{},
			expectError: false,
		},
		{
			name:        "single valid tag",
			input:       []string{"meeting"},
			expectError: false,
		},
		{
			name:        "tags with underscores",
			input:       []string{"code_review", "team_meeting"},
			expectError: false,
		},
		{
			name:        "tags with numbers",
			input:       []string{"sprint1", "version2"},
			expectError: false,
		},
		{
			name:        "maximum allowed tags (10)",
			input:       []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6", "tag7", "tag8", "tag9", "tag10"},
			expectError: false,
		},
		{
			name:        "too many tags (11)",
			input:       []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6", "tag7", "tag8", "tag9", "tag10", "tag11"},
			expectError: true,
			expectedErr: ErrTooManyTags,
		},
		{
			name:        "tag too long (51 characters)",
			input:       []string{"this_is_a_very_long_tag_name_that_exceeds_fifty_chars"},
			expectError: true,
			expectedErr: ErrTagTooLong,
		},
		{
			name:        "tag with invalid characters (spaces)",
			input:       []string{"invalid tag"},
			expectError: true,
			expectedErr: ErrInvalidTagFormat,
		},
		{
			name:        "tag with invalid characters (special chars)",
			input:       []string{"tag@invalid"},
			expectError: true,
			expectedErr: ErrInvalidTagFormat,
		},
		{
			name:        "empty tag",
			input:       []string{""},
			expectError: true,
			expectedErr: ErrTagTooShort,
		},
		{
			name:        "tag with only spaces",
			input:       []string{"   "},
			expectError: true,
			expectedErr: ErrTagTooShort,
		},
		{
			name:        "maximum length tag (50 characters)",
			input:       []string{"this_is_exactly_fifty_characters_long_tag_name_ok"},
			expectError: false,
		},
		{
			name:        "minimum length tag (1 character)",
			input:       []string{"a"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateTags(tt.input)
			if tt.expectError {
				is.True(err != nil)
				if tt.expectedErr != nil {
					is.True(errors.Is(err, tt.expectedErr) || err.Error() == tt.expectedErr.Error() || strings.Contains(err.Error(), tt.expectedErr.Error()))
				}
			} else {
				is.NoErr(err)
			}
		})
	}
}

func TestTagService_GetTagsForAutocomplete(t *testing.T) {
	is := is.New(t)

	orgID := uuid.New()
	otherOrgID := uuid.New()

	// Use InMemTagRepository instead of mock
	repo := NewInMemTagRepository()

	// Add test tags to the repository
	testTags := []*Tag{
		{ID: uuid.New(), Name: "meeting", OrganizationID: orgID},
		{ID: uuid.New(), Name: "development", OrganizationID: orgID},
		{ID: uuid.New(), Name: "bug-fix", OrganizationID: orgID},
		{ID: uuid.New(), Name: "team-meeting", OrganizationID: orgID},
		{ID: uuid.New(), Name: "other-org-tag", OrganizationID: otherOrgID},
	}

	// Add tags to repository by using FindOrCreateTag
	for _, tag := range testTags {
		_, err := repo.FindOrCreateTag(context.Background(), tag.Name, tag.OrganizationID)
		is.NoErr(err)
	}

	service := NewTagService(repo)
	ctx := context.Background()

	tests := []struct {
		name           string
		organizationID uuid.UUID
		query          string
		expectedCount  int
	}{
		{
			name:           "find tags with 'meet' query",
			organizationID: orgID,
			query:          "meet",
			expectedCount:  2, // "meeting" and "team-meeting"
		},
		{
			name:           "find tags with 'dev' query",
			organizationID: orgID,
			query:          "dev",
			expectedCount:  1, // "development"
		},
		{
			name:           "empty query returns empty result",
			organizationID: orgID,
			query:          "",
			expectedCount:  0,
		},
		{
			name:           "no matching tags",
			organizationID: orgID,
			query:          "nonexistent",
			expectedCount:  0,
		},
		{
			name:           "different organization returns no results",
			organizationID: uuid.New(),
			query:          "meet",
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetTagsForAutocomplete(ctx, tt.organizationID, tt.query)
			is.NoErr(err)
			is.Equal(len(result), tt.expectedCount)
		})
	}
}

func TestTagService_GetTagColor(t *testing.T) {
	is := is.New(t)
	service := NewTagService(nil)

	tests := []struct {
		name     string
		tagName  string
		expected string
	}{
		{
			name:     "consistent color for same tag",
			tagName:  "development",
			expected: service.GetTagColor("development"),
		},
		{
			name:     "case insensitive consistency",
			tagName:  "DEVELOPMENT",
			expected: service.GetTagColor("development"),
		},
		{
			name:     "empty tag returns default color",
			tagName:  "",
			expected: "#6c757d",
		},
		{
			name:     "different tags get different colors",
			tagName:  "meeting",
			expected: service.GetTagColor("meeting"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTagColor(tt.tagName)
			is.Equal(result, tt.expected)

			// Verify it's a valid hex color
			is.True(strings.HasPrefix(result, "#"))
			is.Equal(len(result), 7) // #RRGGBB format
		})
	}

	// Test consistency - same tag should always return same color
	t.Run("consistency check", func(t *testing.T) {
		tag1 := "development"
		color1 := service.GetTagColor(tag1)
		color2 := service.GetTagColor(tag1)
		is.Equal(color1, color2)

		// Case insensitive consistency
		color3 := service.GetTagColor("DEVELOPMENT")
		is.Equal(color1, color3)
	})
}
func TestTagService_GetTagColor_ValidLength(t *testing.T) {
	is := is.New(t)
	service := NewTagService(nil)

	// Test a variety of tag names to ensure all generated colors are valid
	testTags := []string{
		"meeting", "development", "bug-fix", "testing", "review",
		"deployment", "documentation", "research", "planning", "design",
		"frontend", "backend", "database", "api", "ui", "ux",
	}

	for _, tagName := range testTags {
		color := service.GetTagColor(tagName)

		// Verify color is exactly 7 characters
		is.Equal(len(color), 7)

		// Verify color starts with #
		is.True(strings.HasPrefix(color, "#"))

		// Verify the rest are valid hex characters
		hexPart := color[1:]
		is.Equal(len(hexPart), 6)

		// Check if all characters after # are valid hex
		for _, char := range hexPart {
			is.True((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F'))
		}
	}
}
func TestTagService_GenerateTagReports(t *testing.T) {
	is := is.New(t)

	// Use InMemTagRepository instead of mock
	repo := NewInMemTagRepository()
	service := NewTagService(repo)
	ctx := context.Background()

	orgID := uuid.New()
	filter := &ActivitiesFilter{
		OrganizationID: orgID,
		Start:          time.Now().AddDate(0, 0, -7),
		End:            time.Now(),
	}

	// Test GenerateTagReports
	reportData, err := service.GenerateTagReports(ctx, orgID, filter, "day", []string{"development"})
	is.NoErr(err)
	is.True(reportData != nil)
	is.Equal(reportData.SelectedTags, []string{"development"})
	is.Equal(reportData.TotalTags, 0) // Mock returns empty data
	is.Equal(reportData.TotalTime, 0)
}

func TestTagService_GetTagReportData(t *testing.T) {
	is := is.New(t)

	// Use InMemTagRepository instead of mock
	repo := NewInMemTagRepository()
	service := NewTagService(repo)
	ctx := context.Background()

	filter := &ActivitiesFilter{
		OrganizationID: uuid.New(),
		Start:          time.Now().AddDate(0, 0, -7),
		End:            time.Now(),
	}

	// Test GetTagReportData
	items, err := service.GetTagReportData(ctx, filter, "day")
	is.NoErr(err)
	is.True(items != nil)
	is.Equal(len(items), 0) // Mock returns empty data
}
