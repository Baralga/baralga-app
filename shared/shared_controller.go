package shared

import "github.com/go-chi/chi/v5"

type DomainHandler interface {
	RegisterProtected(router chi.Router)
	RegisterOpen(router chi.Router)
}
