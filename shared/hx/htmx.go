package hx

import (
	"net/http"

	ghttp "github.com/maragudk/gomponents-htmx/http"
)

func IsHXRequest(r *http.Request) bool {
	return ghttp.IsRequest(r.Header)
}

func IsHXTargetRequest(r *http.Request, target string) bool {
	return IsHXRequest(r) && r.Header.Get("HX-Target") == target
}
