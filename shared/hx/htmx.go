package hx

import (
	"net/http"

	g "github.com/maragudk/gomponents"
)

func Delete(action string) g.Node {
	return g.Attr("hx-delete", action)
}

func Confirm(message string) g.Node {
	return g.Attr("hx-confirm", message)
}

func IsHXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func IsHXTargetRequest(r *http.Request, target string) bool {
	return IsHXRequest(r) && r.Header.Get("HX-Target") == target
}
