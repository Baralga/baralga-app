package hx

import (
	"net/http"
)

func IsHXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func IsHXTargetRequest(r *http.Request, target string) bool {
	return IsHXRequest(r) && r.Header.Get("HX-Target") == target
}
