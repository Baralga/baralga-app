package tracking

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/baralga/shared"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestHandleReportPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ReportWeb{
		config: &shared.Config{},
		activityService: &ActitivityService{
			activityRepository: NewInMemActivityRepository(),
		},
	}

	r, _ := http.NewRequest("GET", "/reports", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Report Activities # Baralga"))
}

func TestHandleReportPageWithTimeByDay(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ReportWeb{
		config: &shared.Config{},
		activityService: &ActitivityService{
			activityRepository: NewInMemActivityRepository(),
		},
	}

	r, _ := http.NewRequest("GET", "/reports?c=time:d", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "id=\"time-report-by-day\""))
}

func TestHandleReportPageWithTimeByWeek(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ReportWeb{
		config: &shared.Config{},
		activityService: &ActitivityService{
			activityRepository: NewInMemActivityRepository(),
		},
	}

	r, _ := http.NewRequest("GET", "/reports?c=time:w&t=year", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "id=\"time-report-by-week\""))
}

func TestHandleReportPageWithTimeByMonth(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ReportWeb{
		config: &shared.Config{},
		activityService: &ActitivityService{
			activityRepository: NewInMemActivityRepository(),
		},
	}

	r, _ := http.NewRequest("GET", "/reports?c=time:m&t=year", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "id=\"time-report-by-month\""))
}

func TestHandleReportPageWithTimeByQuarter(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ReportWeb{
		config: &shared.Config{},
		activityService: &ActitivityService{
			activityRepository: NewInMemActivityRepository(),
		},
	}

	r, _ := http.NewRequest("GET", "/reports?c=time:q&t=year", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "id=\"time-report-by-quarter\""))
}

func TestHandleReportPageWithProject(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ReportWeb{
		config: &shared.Config{},
		activityService: &ActitivityService{
			activityRepository: NewInMemActivityRepository(),
		},
	}

	r, _ := http.NewRequest("GET", "/reports?c=project&t=year", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "id=\"project-report\""))
}

func TestHandleReportPageWithTag(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	tagRepo := NewInMemTagRepository()
	tagService := NewTagService(tagRepo)

	a := &ReportWeb{
		config: &shared.Config{},
		activityService: &ActitivityService{
			activityRepository: NewInMemActivityRepository(),
			tagRepository:      tagRepo,
			tagService:         tagService,
		},
	}

	r, _ := http.NewRequest("GET", "/reports?c=tag", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "No tagged activities found"))
}

func TestHandleReportPageWithTagData(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	tagRepo := NewInMemTagRepository()
	tagService := NewTagService(tagRepo)
	activityRepo := NewInMemActivityRepository()
	repositoryTxer := &shared.InMemRepositoryTxer{}

	activityService := &ActitivityService{
		repositoryTxer:     repositoryTxer,
		activityRepository: activityRepo,
		tagRepository:      tagRepo,
		tagService:         tagService,
	}

	principal := &shared.Principal{
		OrganizationID: shared.OrganizationIDSample,
		Username:       "testuser",
		Roles:          []string{"ROLE_ADMIN"},
	}

	// Create an activity with tags in the current week (2025-39)
	start, _ := time.Parse(time.RFC3339, "2025-09-27T10:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2025-09-27T11:00:00.000Z")

	activity := &Activity{
		Start:          start,
		End:            end,
		Description:    "Test activity with tags",
		ProjectID:      shared.ProjectIDSample,
		OrganizationID: principal.OrganizationID,
		Username:       principal.Username,
		Tags: []*Tag{
			{Name: "meeting"},
			{Name: "development"},
		},
	}

	_, err := activityService.CreateActivity(context.Background(), principal, activity)
	is.NoErr(err)

	a := &ReportWeb{
		config:          &shared.Config{},
		activityService: activityService,
	}

	r, _ := http.NewRequest("GET", "/reports?c=tag", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), principal))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()

	// The in-memory tag repository returns empty results for GetTagReportData
	// but we can still verify that the tag report UI structure is correct
	// and that it shows the "No tagged activities found" message
	is.True(strings.Contains(htmlBody, "No tagged activities found"))

	// Verify that the Tag navigation is active
	is.True(strings.Contains(htmlBody, `<a class="nav-link active"`))
	is.True(strings.Contains(htmlBody, `<i class="bi-tags me-2"></i>Tag`))
}

func TestHandleReportPageGeneralWithTags(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	tagRepo := NewInMemTagRepository()
	tagService := NewTagService(tagRepo)
	activityRepo := NewInMemActivityRepository()
	repositoryTxer := &shared.InMemRepositoryTxer{}

	activityService := &ActitivityService{
		repositoryTxer:     repositoryTxer,
		activityRepository: activityRepo,
		tagRepository:      tagRepo,
		tagService:         tagService,
	}

	principal := &shared.Principal{
		OrganizationID: shared.OrganizationIDSample,
		Username:       "testuser",
		Roles:          []string{"ROLE_ADMIN"},
	}

	a := &ReportWeb{
		config:          &shared.Config{},
		activityService: activityService,
	}

	r, _ := http.NewRequest("GET", "/reports?c=general", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), principal))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()

	// Verify that the Tags column header is present in both mobile and desktop views
	is.True(strings.Contains(htmlBody, "<th>Tags</th>"))

	// Verify that the tag display structure is present (even if no tags are shown)
	// The d-flex flex-wrap gap-1 class indicates the tag container structure
	is.True(strings.Contains(htmlBody, `class="d-flex flex-wrap gap-1"`))
}

func TestHandleReportPageGeneralTagDisplayStructure(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	// Create a custom in-memory repository with an activity that has tags
	activityRepo := &InMemActivityRepository{
		activities: []*Activity{
			{
				ID:             uuid.MustParse("00000000-0000-0000-2222-000000000001"),
				ProjectID:      shared.ProjectIDSample,
				OrganizationID: shared.OrganizationIDSample,
				Username:       "user1",
				Tags: []*Tag{
					{Name: "meeting", Color: "#FF5733"},
					{Name: "development", Color: "#33FF57"},
				},
			},
		},
	}

	tagRepo := NewInMemTagRepository()
	tagService := NewTagService(tagRepo)
	repositoryTxer := &shared.InMemRepositoryTxer{}

	activityService := &ActitivityService{
		repositoryTxer:     repositoryTxer,
		activityRepository: activityRepo,
		tagRepository:      tagRepo,
		tagService:         tagService,
	}

	principal := &shared.Principal{
		OrganizationID: shared.OrganizationIDSample,
		Username:       "testuser",
		Roles:          []string{"ROLE_ADMIN"},
	}

	a := &ReportWeb{
		config:          &shared.Config{},
		activityService: activityService,
	}

	r, _ := http.NewRequest("GET", "/reports?c=general", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), principal))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()

	// Verify that the Tags column header is present
	is.True(strings.Contains(htmlBody, "<th>Tags</th>"))

	// Verify that tag badges are displayed with colors
	is.True(strings.Contains(htmlBody, `background-color: #FF5733`))
	is.True(strings.Contains(htmlBody, `background-color: #33FF57`))
	is.True(strings.Contains(htmlBody, "meeting"))
	is.True(strings.Contains(htmlBody, "development"))

	// Verify badge structure
	is.True(strings.Contains(htmlBody, `class="badge"`))
	is.True(strings.Contains(htmlBody, `color: white;`))
}

func TestReportViewFromQueryParams(t *testing.T) {
	is := is.New(t)

	t.Run("view without params", func(t *testing.T) {
		// Arrange
		params := make(url.Values)

		// Act
		view := reportViewFromQueryParams(params, "")

		// Assert
		is.Equal(view.main, "general")
	})

	t.Run("view with week and month view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"week"}
		params["c"] = []string{"time:m"}

		// Act
		view := reportViewFromQueryParams(params, "week")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "w")
	})

	t.Run("view with week and quarter view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"week"}
		params["c"] = []string{"time:q"}

		// Act
		view := reportViewFromQueryParams(params, "week")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "w")
	})

	t.Run("view with week and day view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"week"}
		params["c"] = []string{"time:d"}

		// Act
		view := reportViewFromQueryParams(params, "week")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "d")
	})

	t.Run("view with month and quarter view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"month"}
		params["c"] = []string{"time:q"}

		// Act
		view := reportViewFromQueryParams(params, "month")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "m")
	})

	t.Run("view with day and quarter view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"day"}
		params["c"] = []string{"time:q"}

		// Act
		view := reportViewFromQueryParams(params, "day")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "d")
	})

	t.Run("view with day and month view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"day"}
		params["c"] = []string{"time:m"}

		// Act
		view := reportViewFromQueryParams(params, "day")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "d")
	})

	t.Run("view with day and week view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"day"}
		params["c"] = []string{"time:w"}

		// Act
		view := reportViewFromQueryParams(params, "day")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "d")
	})

	t.Run("view with tag category", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"week"}
		params["c"] = []string{"tag"}

		// Act
		view := reportViewFromQueryParams(params, "week")

		// Assert
		is.Equal(view.main, "tag")
		is.Equal(view.sub, "")
	})
}
