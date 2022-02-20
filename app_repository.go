package main

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

const contextKeyTx contextKey = 1

type RepositoryTxer interface {
	InTx(ctx context.Context, txFuncs ...func(ctxWithTx context.Context) error) error
}

// DbUserRepository is a SQL database repository for users
type DbRepositoryTxer struct {
	connPool *pgxpool.Pool
}

var _ RepositoryTxer = (*DbRepositoryTxer)(nil)

func NewDbRepositoryTxer(connPool *pgxpool.Pool) *DbRepositoryTxer {
	return &DbRepositoryTxer{
		connPool: connPool,
	}
}

func (txer *DbRepositoryTxer) InTx(ctx context.Context, txFuncs ...func(ctxWithTx context.Context) error) error {
	tx, err := txer.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	ctxWithTx := context.WithValue(ctx, contextKeyTx, tx)

	for _, txFunc := range txFuncs {
		err = txFunc(ctxWithTx)
		if err != nil {
			rb := tx.Rollback(ctx)
			if rb != nil {
				return errors.Wrap(rb, "rollback error")
			}
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
