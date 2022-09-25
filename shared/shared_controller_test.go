package shared

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestRenderJSON(t *testing.T) {
	is := is.New(t)

	w := httptest.NewRecorder()

	c := struct {
		Name string
		Type string
	}{
		Name: "Sammy",
		Type: "Shark",
	}

	RenderJSON(w, c)

	is.True(strings.Contains(w.Body.String(), "Shark"))
}

func TestRenderProblemJSONInProduction(t *testing.T) {
	is := is.New(t)

	isProduction := true
	err := errors.New("my error")
	w := httptest.NewRecorder()

	RenderProblemJSON(w, isProduction, err)

	is.True(!strings.Contains(w.Body.String(), "my error"))
}

func TestRenderProblemJSON(t *testing.T) {
	is := is.New(t)

	isProduction := false
	err := errors.New("my error")
	w := httptest.NewRecorder()

	RenderProblemJSON(w, isProduction, err)

	is.True(strings.Contains(w.Body.String(), "my error"))
}
