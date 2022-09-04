package util

import (
	"log"
	"net/http"

	"schneider.vip/problem"
)

func RenderProblemJSON(w http.ResponseWriter, isProduction bool, err error) {
	log.Printf("internal server error: %s", err)

	if !isProduction {
		http.Error(w, problem.New(problem.Title("internal server error"), problem.Wrap(err)).JSONString(), http.StatusInternalServerError)
		return
	}

	http.Error(w, problem.New(problem.Title("internal server error")).JSONString(), http.StatusInternalServerError)
}
