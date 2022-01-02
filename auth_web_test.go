package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestHandleLoginPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config: &config{},
	}

	r, _ := http.NewRequest("GET", "/login", nil)

	a.HandleLoginPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Baralga # Sign In"))
}
