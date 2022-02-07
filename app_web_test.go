package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestHandleIndexPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ProjectRepository:  NewInMemProjectRepository(),
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleIndexPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Track Activities # Baralga"))
}

func TestHandleWebManifest(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ProjectRepository:  NewInMemProjectRepository(),
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/manifest.webmanifest", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleWebManifest()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
}

func TestHandleReportPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ProjectRepository:  NewInMemProjectRepository(),
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/reports", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Report Activities # Baralga"))
}
