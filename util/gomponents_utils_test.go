package util

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
	"github.com/matryer/is"
)

func TestRenderHTML(t *testing.T) {
	// Arrange
	is := is.New(t)
	w := httptest.NewRecorder()
	n := Div(g.Text("Hello HTML!"))

	// Act
	RenderHTML(w, n)

	// Assert
	is.Equal(w.Body.String(), "<div>Hello HTML!</div>")
}

func TestRenderProblemHTMLInProduction(t *testing.T) {
	// Arrange
	is := is.New(t)
	w := httptest.NewRecorder()
	e := errors.New("BAM")
	isProduction := true

	// Act
	RenderProblemHTML(w, isProduction, e)

	// Assert
	is.True(strings.Contains(w.Body.String(), "internal server error"))
	is.True(!strings.Contains(w.Body.String(), "BAM"))
}

func TestRenderProblemHTMLInDevelopment(t *testing.T) {
	// Arrange
	is := is.New(t)
	w := httptest.NewRecorder()
	e := errors.New("BAM")
	isProduction := false

	// Act
	RenderProblemHTML(w, isProduction, e)

	// Assert
	is.True(strings.Contains(w.Body.String(), "internal server error: BAM"))
}
