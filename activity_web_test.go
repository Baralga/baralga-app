package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/matryer/is"
)

func TestHandleActivityAddPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/activities/new", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleActivityAddPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "<form"))
}

func TestHandleActivityEditPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ProjectRepository:  NewInMemProjectRepository(),
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/activities/00000000-0000-0000-2222-000000000001/edit", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleActivityEditPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "<form"))
}
