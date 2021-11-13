package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/baralga/hal"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestMapToActivity(t *testing.T) {
	is := is.New(t)

	activityModel := &activityModel{
		ID:    "d9fbfab6-2750-4703-8a7b-77498756d64a",
		Start: "2021-11-06T21:37:00",
		End:   "2021-11-06T21:37:00",
		Links: hal.NewLinks(
			hal.NewLink("project", "/api/projects/efa45cae-5dc7-412a-887f-945ddbb0a23f"),
		),
	}

	activity, err := mapToActivity(activityModel)

	is.NoErr(err)
	is.Equal(activityModel.ID, activity.ID.String())
	is.Equal(uuid.MustParse("efa45cae-5dc7-412a-887f-945ddbb0a23f"), activity.ProjectID)
	is.Equal(2021, activity.Start.Year())
	is.Equal(2021, activity.End.Year())
}

func TestMapToActivityIdNotValid(t *testing.T) {
	is := is.New(t)

	activityModel := &activityModel{
		ID:    "no-uuid",
		Start: "2021-11-06T21:37:00",
		End:   "2021-11-06T21:37:00",
		Links: hal.NewLinks(
			hal.NewLink("project", "/api/projects/efa45cae-5dc7-412a-887f-945ddbb0a23f"),
		),
	}

	_, err := mapToActivity(activityModel)

	is.True(err != nil)
}

func TestMapToActivityModel(t *testing.T) {
	is := is.New(t)

	start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

	activity := &Activity{
		ID:        uuid.New(),
		Start:     start,
		End:       end,
		ProjectID: uuid.New(),
	}

	activityModel := mapToActivityModel(activity)

	is.Equal(activity.ID.String(), activityModel.ID)
	is.True(strings.Contains(activityModel.Links.HrefOf("project"), activity.ProjectID.String()))
	is.True(strings.Contains(activityModel.Start, "2021"))
	is.True(strings.Contains(activityModel.End, "2021"))
}

func TestHandleGetActivity(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/activities/00000000-0000-0000-2222-000000000001", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleGetActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
}

func TestHandleGetActivityNotFound(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/activities/d9fbfab6-2750-4703-8a7b-77498756d64a", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "d9fbfab6-2750-4703-8a7b-77498756d64a")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleGetActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotFound)
}

func TestHandleGetActivityIdNotValid(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/activities/not-a-uuid", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "not-a-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleGetActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleGetActivitiesWithUrlParams(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ActivityRepository: NewInMemActivityRepository(),
		ProjectRepository:  NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/activities?start=2021-10-01&end=2022-10-01", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleGetActivities()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	activitiesModel := &activitiesModel{}
	err := json.NewDecoder(httpRec.Body).Decode(activitiesModel)
	is.NoErr(err)
	is.Equal(1, len(activitiesModel.ActivityModels))
}

func TestHandleCreateActivity(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	a := &app{
		Config:             &config{},
		ActivityRepository: repo,
	}

	countBefore := len(repo.activities)
	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description":"",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleCreateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusCreated)
	is.Equal(countBefore+1, len(repo.activities))
}

func TestHandleDeleteActivityAsAdmin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	a := &app{
		Config:             &config{},
		ActivityRepository: repo,
	}

	r, _ := http.NewRequest("DELETE", "/api/activities/00000000-0000-0000-2222-000000000001", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Username: "admin",
		Roles:    []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleDeleteActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(0, len(repo.activities))
}

func TestHandleDeleteActivityAsMatchingUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	a := &app{
		Config:             &config{},
		ActivityRepository: repo,
	}

	r, _ := http.NewRequest("DELETE", "/api/activities/00000000-0000-0000-2222-000000000001", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Username: "user1",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleDeleteActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(0, len(repo.activities))
}

func TestHandleDeleteActivityIdNotValid(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("DELETE", "/api/activities/not-a-uuid", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "not-a-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleDeleteActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleCreateActivityWithInvalidBody(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ActivityRepository: NewInMemActivityRepository(),
	}

	body := `
	{
		INVALID!!
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleCreateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleDeleteActivityAsNonMatchingUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	a := &app{
		Config:             &config{},
		ActivityRepository: repo,
	}

	r, _ := http.NewRequest("DELETE", "/api/activities/00000000-0000-0000-2222-000000000001", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Username: "otherUser",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleDeleteActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotFound)
	is.Equal(1, len(repo.activities))
}

func TestHandleUpdateActivity(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	a := &app{
		Config:             &config{},
		ActivityRepository: repo,
	}

	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description": "My updated Description",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	activityUpdate, err := repo.FindActivityByID(context.Background(), uuid.MustParse("00000000-0000-0000-2222-000000000001"), organizationIDSample)
	is.NoErr(err)
	is.Equal("My updated Description", activityUpdate.Description)
}

func TestHandleUpdateActivityAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	a := &app{
		Config:             &config{},
		ActivityRepository: repo,
	}

	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description": "My updated Description",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Username: "user1",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	activityUpdate, err := repo.FindActivityByID(context.Background(), uuid.MustParse("00000000-0000-0000-2222-000000000001"), organizationIDSample)
	is.NoErr(err)
	is.Equal("My updated Description", activityUpdate.Description)
}

func TestHandleUpdateActivityWithNonMatchingUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	a := &app{
		Config:             &config{},
		ActivityRepository: repo,
	}

	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description": "My updated Description",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Username: "otherUser",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotFound)
}

func TestHandleUpdateActivityWithInvalidBody(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ActivityRepository: NewInMemActivityRepository(),
	}

	body := `
	{
		INVALID!!
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{
		Username: "otherUser",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}

func TestHandleUpdateActivityIdNotValid(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ActivityRepository: NewInMemActivityRepository(),
	}

	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description": "My updated Description",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities/not-a-uuid", strings.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "not-a-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotAcceptable)
}
