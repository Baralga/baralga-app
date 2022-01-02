package util

import (
	"log"
	"net/http"

	g "github.com/maragudk/gomponents"
)

func RenderHTML(w http.ResponseWriter, n g.Node) {
	w.Header().Set("Content-Type", "text/html")
	err := n.Render(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RenderProblemHTML(w http.ResponseWriter, isProduction bool, err error) {
	log.Printf("internal server error: %s", err)

	if !isProduction {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, "internal server error", http.StatusInternalServerError)
}
