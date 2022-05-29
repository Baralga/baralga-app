package hx

import (
	"net/http"

	g "github.com/maragudk/gomponents"
)

func Boost() g.Node {
	return g.Attr("hx-boost", "true")
}

func PushURLTrue() g.Node {
	return PushURL("true")
}

func PushURL(url string) g.Node {
	return g.Attr("hx-push-url", url)
}

func Post(action string) g.Node {
	return g.Attr("hx-post", action)
}

func Delete(action string) g.Node {
	return g.Attr("hx-delete", action)
}

func Get(action string) g.Node {
	return g.Attr("hx-get", action)
}

func Target(action string) g.Node {
	return g.Attr("hx-target", action)
}

func Swap(action string) g.Node {
	return g.Attr("hx-swap", action)
}

func Trigger(action string) g.Node {
	return g.Attr("hx-trigger", action)
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
