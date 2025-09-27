package shared

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type contextKey int

const (
	contextKeyPrincipal contextKey = 0
	contextKeyTx        contextKey = 1
)

type Principal struct {
	Name           string
	Username       string
	OrganizationID uuid.UUID
	Roles          []string
}

// MustPrincipalFromContext reads the current principal from the context or panics if not present
func MustPrincipalFromContext(ctx context.Context) *Principal {
	principal, ok := ctx.Value(contextKeyPrincipal).(*Principal)
	if !ok {
		panic("no principal found in context")
	}

	return principal
}

// ToContextWithPrincipal creates a new context with the principal as value
func ToContextWithPrincipal(ctx context.Context, principal *Principal) context.Context {
	return context.WithValue(ctx, contextKeyPrincipal, principal)
}

func (p *Principal) HasRole(role string) bool {
	for _, c := range p.Roles {
		if c == role {
			return true
		}
	}
	return false
}

type RepositoryTxer interface {
	InTx(ctx context.Context, txFuncs ...func(ctxWithTx context.Context) error) error
}

type MailResource interface {
	SendMail(to, subject, body string) error
}

// Common error types
var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = func(message string) error { return errors.New(message) }
	ErrConflict   = func(message string) error { return errors.New(message) }
)
