package util

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
)

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
