package shared

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

	a := &App{
		Config: &Config{},
	}

	r, _ := http.NewRequest("GET", "/manifest.webmanifest", nil)
	r = r.WithContext(context.WithValue(r.Context(), ContextKeyPrincipal, &Principal{}))

	a.HandleWebManifest()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
}
