package shared

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func TestHandleWebManifest(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	r, _ := http.NewRequest("GET", "/manifest.webmanifest", nil)
	r = r.WithContext(ToContextWithPrincipal(r.Context(), &Principal{}))

	HandleWebManifest()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
}

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
