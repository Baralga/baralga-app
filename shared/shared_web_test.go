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

	r, _ := http.NewRequest("GET", "/manifest.webmanifest", nil)
	r = r.WithContext(context.WithValue(r.Context(), ContextKeyPrincipal, &Principal{}))

	HandleWebManifest()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
}
