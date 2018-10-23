package models

import (
	"context"
	"database/sql"

	"github.com/keydotcat/keycatd/util"
)

type contextType int

const (
	contextDBKey = contextType(0)
)

func GetDB(ctx context.Context) *sql.DB {
	d, ok := ctx.Value(contextDBKey).(*sql.DB)
	if !ok {
		panic("No db defined in context")
	}
	return d
}

func doTx(ctx context.Context, ftor func(*sql.Tx) error) error {
	tx, err := GetDB(ctx).BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	if err = ftor(tx); err != nil {
		if util.CheckErr(err, sql.ErrTxDone) || util.CheckErr(err, sql.ErrConnDone) {
			return err
		}
		if rerr := tx.Rollback(); rerr != nil {
			return util.NewErrorf("Could not rollback transaction: %s (prev error was %s)", rerr, err)
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return util.NewErrorf("Could not commit transaction: %s", err)
	}
	return nil
}

func AddDBToContext(ctx context.Context, d *sql.DB) context.Context {
	return context.WithValue(ctx, contextDBKey, d)
}
