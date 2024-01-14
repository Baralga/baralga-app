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
	"github.com/matryer/is"
)

func TestHandleTrackingPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	activityRepository := NewInMemActivityRepository()
	a := &ActivityWebHandlers{
		config:             &shared.Config{},
		activityRepository: activityRepository,
		projectRepository:  NewInMemProjectRepository(),
		activityService: &ActitivityService{
			activityRepository: activityRepository,
		},
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
		activityService: &ActitivityService{
			repositoryTxer:     shared.NewInMemRepositoryTxer(),
			activityRepository: repo,
		},
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
