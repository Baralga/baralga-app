package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestHandleProjectsPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:            &config{},
		ProjectRepository: NewInMemProjectRepository(),
	}

	r, _ := http.NewRequest("GET", "/projects", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleProjectsPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "<form"))
}
