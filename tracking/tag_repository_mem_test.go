package tracking

import (
	"context"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInMemTagRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemTagRepository()
	tagService := NewTagService(repo)

	t.Run("FindTagsByOrganization returns sample tags", func(t *testing.T) {
		tags, err := repo.FindTagsByOrganization(ctx, shared.OrganizationIDSample, "")
		assert.NoError(t, err)
		assert.Len(t, tags, 3)
		assert.Equal(t, "development", tags[0].Name)
		assert.Equal(t, "meeting", tags[1].Name)
		assert.Equal(t, "testing", tags[2].Name)
	})

	t.Run("FindTagsByOrganization with query filters results", func(t *testing.T) {
		tags, err := repo.FindTagsByOrganization(ctx, shared.OrganizationIDSample, "dev")
		assert.NoError(t, err)
		assert.Len(t, tags, 1)
		assert.Equal(t, "development", tags[0].Name)
	})

	t.Run("FindOrCreateTag finds existing tag", func(t *testing.T) {
		tag, err := repo.FindOrCreateTag(ctx, "Development", shared.OrganizationIDSample)
		assert.NoError(t, err)
		assert.Equal(t, "development", tag.Name) // normalized to lowercase
		assert.Equal(t, shared.OrganizationIDSample, tag.OrganizationID)
	})

	t.Run("FindOrCreateTag creates new tag", func(t *testing.T) {
		tag, err := repo.FindOrCreateTag(ctx, "New Tag", shared.OrganizationIDSample)
		assert.NoError(t, err)
		assert.Equal(t, "new tag", tag.Name) // normalized to lowercase
		assert.Equal(t, shared.OrganizationIDSample, tag.OrganizationID)
		assert.NotEqual(t, uuid.Nil, tag.ID)
	})

	t.Run("SyncTagsForActivity creates relationships", func(t *testing.T) {
		activityID := uuid.New()
		tagNames := []string{"development", "New Feature"}

		// Prepare tags with colors using the service
		tagsWithColors := tagService.PrepareTagsWithColors(tagNames)

		err := repo.SyncTagsForActivity(ctx, activityID, shared.OrganizationIDSample, tagsWithColors)
		assert.NoError(t, err)

		// Verify the tags were created/found
		tags, err := repo.FindTagsByOrganization(ctx, shared.OrganizationIDSample, "")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(tags), 4) // original 3 + at least 1 new

		// Check that "new feature" was created
		found := false
		for _, tag := range tags {
			if tag.Name == "new feature" {
				found = true
				break
			}
		}
		assert.True(t, found, "new feature tag should be created")
	})

	t.Run("DeleteUnusedTags removes unused tags", func(t *testing.T) {
		// Create a tag that won't be used
		unusedTag, err := repo.FindOrCreateTag(ctx, "unused tag", shared.OrganizationIDSample)
		assert.NoError(t, err)

		// Create an activity with some tags
		activityID := uuid.New()
		tagsWithColors := tagService.PrepareTagsWithColors([]string{"development"})
		err = repo.SyncTagsForActivity(ctx, activityID, shared.OrganizationIDSample, tagsWithColors)
		assert.NoError(t, err)

		// Delete unused tags
		err = repo.DeleteUnusedTags(ctx, shared.OrganizationIDSample)
		assert.NoError(t, err)

		// Verify unused tag was removed
		tags, err := repo.FindTagsByOrganization(ctx, shared.OrganizationIDSample, "")
		assert.NoError(t, err)

		for _, tag := range tags {
			assert.NotEqual(t, unusedTag.ID, tag.ID, "unused tag should be deleted")
		}
	})

	t.Run("SyncTagsForActivity generates colors for new tags", func(t *testing.T) {
		repo := NewInMemTagRepository()
		tagService := NewTagService(repo)
		ctx := context.Background()
		activityID := uuid.New()
		tagNames := []string{"bug-fix", "new-feature"}

		// Prepare tags with colors using the service
		tagsWithColors := tagService.PrepareTagsWithColors(tagNames)

		err := repo.SyncTagsForActivity(ctx, activityID, shared.OrganizationIDSample, tagsWithColors)
		assert.NoError(t, err)

		// Verify tags were created with generated colors
		tags, err := repo.FindTagsByOrganization(ctx, shared.OrganizationIDSample, "")
		assert.NoError(t, err)

		// Should have original 3 sample tags + 2 new ones = 5 total
		assert.Equal(t, 5, len(tags))

		// Find the new tags and verify they have colors (not default gray)
		for _, tag := range tags {
			if tag.Name == "bug-fix" || tag.Name == "new-feature" {
				assert.NotEqual(t, "", tag.Color)
				assert.True(t, strings.HasPrefix(tag.Color, "#"))
				assert.Equal(t, 7, len(tag.Color)) // #RRGGBB format
				// Verify it's not the default color (should be generated)
				expectedColor := tagService.GetTagColor(tag.Name)
				assert.Equal(t, expectedColor, tag.Color)
			}
		}
	})
}
