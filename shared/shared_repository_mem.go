package shared

import (
	"context"

	_ "github.com/golang-migrate/migrate/v4/database/pgx"
)

type InMemRepositoryTxer struct {
}

var _ RepositoryTxer = (*InMemRepositoryTxer)(nil)

func NewInMemRepositoryTxer() *InMemRepositoryTxer {
	return &InMemRepositoryTxer{}
}

func (txer *InMemRepositoryTxer) InTx(ctx context.Context, txFuncs ...func(ctxWithTx context.Context) error) error {
	for _, txFunc := range txFuncs {
		err := txFunc(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
