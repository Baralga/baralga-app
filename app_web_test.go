package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

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
