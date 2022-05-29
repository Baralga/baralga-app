package shared

import (
	"context"

	"github.com/google/uuid"
)

type contextKey int

const (
	ContextKeyPrincipal contextKey = 0
	ContextKeyTx        contextKey = 1
)

type Principal struct {
	Name           string
	Username       string
	OrganizationID uuid.UUID
	Roles          []string
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
