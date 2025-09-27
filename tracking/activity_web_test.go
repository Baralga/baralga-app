package tracking

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/baralga/shared"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

// Helper function to create a properly initialized ActivityService for web tests
func createTestActivityServiceForWeb(repo ActivityRepository) *ActitivityService {
	tagRepo := NewInMemTagRepository()
	tagService := NewTagService(tagRepo)
	tagRepo.SetTagService(tagService)
	return &ActitivityService{
		repositoryTxer:     shared.NewInMemRepositoryTxer(),
		activityRepository: repo,
		tagRepository:      tagRepo,
		tagService:         tagService,
	}
}

func TestHandleTrackingPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	activityRepository := NewInMemActivityRepository()
	a := &ActivityWebHandlers{
		config:             &shared.Config{},
		activityRepository: activityRepository,
		projectRepository:  NewInMemProjectRepository(),
		activityService:    createTestActivityServiceForWeb(activityRepository),
	}

	r, _ := http.NewRequest("GET", "/", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleTrackingPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Track Activities # Baralga"))
}

func TestHandleActivityAddPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ActivityWebHandlers{
		config:            &shared.Config{},
		projectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/activities/new", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleActivityAddPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "<form"))
	// Verify that the Tags input field is present
	is.True(strings.Contains(htmlBody, `name="Tags"`))
	is.True(strings.Contains(htmlBody, `placeholder="meeting, development, bug-fix"`))
	is.True(strings.Contains(htmlBody, "Separate tags with commas or spaces"))
}

func TestHandleActivityEditPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ActivityWebHandlers{
		config:             &shared.Config{},
		activityRepository: NewInMemActivityRepository(),
		projectRepository:  NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/activities/00000000-0000-0000-2222-000000000001/edit", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleActivityEditPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "<form"))
	// Verify that the Tags input field is present in edit form
	is.True(strings.Contains(htmlBody, `name="Tags"`))
	// The sample activity should have tags "meeting, development" as defined in the in-memory repository
	is.True(strings.Contains(htmlBody, "meeting, development"))
}

func TestHandleCreateActivtiyWithValidActivtiy(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	config := &shared.Config{}

	w := &ActivityWebHandlers{
		config:             config,
		activityRepository: repo,
		projectRepository:  NewInMemProjectRepository(),
		activityService:    createTestActivityServiceForWeb(repo),
	}

	countBefore := len(repo.activities)

	data := url.Values{}
	data["ProjectID"] = []string{shared.ProjectIDSample.String()}
	data["Date"] = []string{"21.12.2021"}
	data["StartTime"] = []string{"10:00"}
	data["EndTime"] = []string{"11:00"}
	data["Description"] = []string{"My description"}

	r, _ := http.NewRequest("POST", "/activities/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	w.HandleActivityForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(countBefore+1, len(repo.activities))
}

func TestHandleCreateActivityWithTags(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	config := &shared.Config{}

	w := &ActivityWebHandlers{
		config:             config,
		activityRepository: repo,
		projectRepository:  NewInMemProjectRepository(),
		activityService:    createTestActivityServiceForWeb(repo),
	}

	countBefore := len(repo.activities)

	data := url.Values{}
	data["ProjectID"] = []string{shared.ProjectIDSample.String()}
	data["Date"] = []string{"21.12.2021"}
	data["StartTime"] = []string{"10:00"}
	data["EndTime"] = []string{"11:00"}
	data["Description"] = []string{"My description"}
	data["Tags"] = []string{"meeting, development, bug-fix"}

	r, _ := http.NewRequest("POST", "/activities/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	w.HandleActivityForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(countBefore+1, len(repo.activities))

	// Verify the activity was created with tags
	createdActivity := repo.activities[len(repo.activities)-1]
	is.Equal(len(createdActivity.Tags), 3)
	is.True(containsTag(createdActivity.Tags, "meeting"))
	is.True(containsTag(createdActivity.Tags, "development"))
	is.True(containsTag(createdActivity.Tags, "bug-fix"))
}

func TestMapFormToActivityWithTags(t *testing.T) {
	is := is.New(t)

	formModel := activityFormModel{
		ProjectID:   shared.ProjectIDSample.String(),
		Date:        "21.12.2021",
		StartTime:   "10:00",
		EndTime:     "11:00",
		Description: "My description",
		Tags:        "meeting, development, bug-fix",
	}

	activity, err := mapFormToActivity(formModel)
	is.NoErr(err)
	is.Equal(len(activity.Tags), 3)
	is.True(containsTag(activity.Tags, "meeting"))
	is.True(containsTag(activity.Tags, "development"))
	is.True(containsTag(activity.Tags, "bug-fix"))
}

func TestMapFormToActivityWithSpaceSeparatedTags(t *testing.T) {
	is := is.New(t)

	formModel := activityFormModel{
		ProjectID:   shared.ProjectIDSample.String(),
		Date:        "21.12.2021",
		StartTime:   "10:00",
		EndTime:     "11:00",
		Description: "My description",
		Tags:        "meeting development bug-fix",
	}

	activity, err := mapFormToActivity(formModel)
	is.NoErr(err)
	is.Equal(len(activity.Tags), 3)
	is.True(containsTag(activity.Tags, "meeting"))
	is.True(containsTag(activity.Tags, "development"))
	is.True(containsTag(activity.Tags, "bug-fix"))
}

func TestMapFormToActivityWithDuplicateTags(t *testing.T) {
	is := is.New(t)

	formModel := activityFormModel{
		ProjectID:   shared.ProjectIDSample.String(),
		Date:        "21.12.2021",
		StartTime:   "10:00",
		EndTime:     "11:00",
		Description: "My description",
		Tags:        "meeting, Meeting, MEETING, development",
	}

	activity, err := mapFormToActivity(formModel)
	is.NoErr(err)
	is.Equal(len(activity.Tags), 2) // Should deduplicate case-insensitive
	is.True(containsTag(activity.Tags, "meeting"))
	is.True(containsTag(activity.Tags, "development"))
}

func TestMapActivityToFormWithTags(t *testing.T) {
	is := is.New(t)

	activityID := uuid.MustParse("00000000-0000-0000-2222-000000000001")
	activity := Activity{
		ID:          activityID,
		ProjectID:   shared.ProjectIDSample,
		Description: "My description",
		Tags: []*Tag{
			{Name: "meeting"},
			{Name: "development"},
			{Name: "bug-fix"},
		},
	}

	formModel := mapActivityToForm(activity)
	is.Equal(formModel.Tags, "meeting, development, bug-fix")
}

func TestHandleTrackingPageDisplaysTagsWithoutFiltering(t *testing.T) {
	// Arrange
	is := is.New(t)

	config := &shared.Config{}
	activityRepository := NewInMemActivityRepository()
	projectRepository := NewInMemProjectRepository()
	tagRepository := NewInMemTagRepository()
	tagService := NewTagService(tagRepository)
	tagRepository.SetTagService(tagService)
	repositoryTxer := shared.NewInMemRepositoryTxer()

	activityService := NewActitivityService(repositoryTxer, activityRepository, tagRepository, tagService)

	handlers := NewActivityWebHandlers(config, activityService, activityRepository, projectRepository)

	// Create a simple request (no tag filtering on web page)
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(shared.ToContextWithPrincipal(req.Context(), &shared.Principal{
		OrganizationID: uuid.New(),
		Username:       "test@example.com",
		Roles:          []string{"ROLE_USER"},
	}))

	w := httptest.NewRecorder()

	// Act
	handlers.HandleTrackingPage()(w, req)

	// Assert
	is.Equal(w.Code, http.StatusOK)
	// The response should not contain tag filter UI (removed from web page)
	responseBody := w.Body.String()
	is.True(!strings.Contains(responseBody, "Filter by tags"))
}

func TestHandleCreateActivtiyWithInvalidActivtiy(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	a := &ActivityWebHandlers{
		config:             &shared.Config{},
		activityRepository: repo,
		projectRepository:  NewInMemProjectRepository(),
	}

	countBefore := len(repo.activities)

	data := url.Values{}
	data["ProjectID"] = []string{shared.ProjectIDSample.String()}
	data["Date"] = []string{"2"}
	data["StartTime"] = []string{"1"}
	data["EndTime"] = []string{"1"}

	r, _ := http.NewRequest("POST", "/activities/new", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	a.HandleActivityForm()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(countBefore, len(repo.activities))
}

func TestHandleStartTimeValidation(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ActivityWebHandlers{
		config: &shared.Config{},
	}

	data := url.Values{}
	data["StartTime"] = []string{"10"}

	r, _ := http.NewRequest("POST", "/activities/validation-start-time", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleStartTimeValidation()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "10:00"))
}

func TestHandleEndTimeValidation(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ActivityWebHandlers{
		config: &shared.Config{},
	}

	data := url.Values{}
	data["StartTime"] = []string{"10"}

	r, _ := http.NewRequest("POST", "/activities/validation-end-time", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleEndTimeValidation()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "10:00"))
}

// Helper function to check if a tag name exists in a slice of Tag objects
func containsTag(tags []*Tag, tagName string) bool {
	for _, tag := range tags {
		if tag.Name == tagName {
			return true
		}
	}
	return false
}
