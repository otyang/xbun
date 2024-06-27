package xbun

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/schema"
)

type (
	SelectCriteria func(q *bun.SelectQuery) *bun.SelectQuery
	UpdateCriteria func(q *bun.UpdateQuery) *bun.UpdateQuery
	DeleteCriteria func(q *bun.DeleteQuery) *bun.DeleteQuery
)

func Create[T any](ctx context.Context, db bun.IDB, ignoreDuplicates bool, model *T) error {
	if ignoreDuplicates {
		_, err := db.NewInsert().Model(model).Ignore().Exec(ctx)
		return err
	}
	_, err := db.NewInsert().Model(model).Exec(ctx)
	return err
}

func Delete[T any](ctx context.Context, db bun.IDB, modelPtr *T, where DeleteCriteria) (int64, error) {
	q := db.NewDelete().Model(modelPtr)

	if where == nil {
		q.WherePK()
	} else {
		where(q)
	}

	r, err := q.Exec(ctx)
	if err != nil {
		return 0, err
	}

	noOfRowsAffected, err := r.RowsAffected()
	if err != nil {
		return 0, err
	}

	return noOfRowsAffected, nil
}

func Upsert[T any](ctx context.Context, db bun.IDB, modelsPtr *T) (int64, error) {
	r, err := db.NewInsert().Model(modelsPtr).On("CONFLICT DO UPDATE").Exec(ctx)
	if err != nil {
		return 0, err
	}

	noOfRowsAffected, err := r.RowsAffected()
	if err != nil {
		return 0, err
	}

	return noOfRowsAffected, nil
}

func Update[T any](ctx context.Context, db bun.IDB, where UpdateCriteria, modelPtr ...T) error {
	q := db.NewUpdate().Model(modelPtr)

	if len(modelPtr) > 1 {
		q.Bulk()
	}

	if where == nil {
		q.WherePK()
	} else {
		where(q)
	}

	_, err := q.Exec(ctx)
	return err
}

func FindOne[T any](ctx context.Context, db bun.IDB, modelPtr *T, where SelectCriteria) (T, error) {
	q := db.NewSelect().Model(modelPtr)

	if where == nil {
		q.WherePK()
	} else {
		where(q)
	}

	err := q.Limit(1).Scan(ctx)
	return *modelPtr, err
}

func FindMany[T any](ctx context.Context, db bun.IDB, limit int, where SelectCriteria) (results []T, hasMore bool, e error) {
	if limit < 0 {
		limit = 0
	}

	var modelsPtr []T

	q := db.NewSelect().Model(&modelsPtr)

	if where != nil {
		where(q)
	}

	if err := q.Limit(limit + 1).Scan(ctx); err != nil {
		return nil, false, err
	}

	if len(modelsPtr) < (limit + 1) {
		return modelsPtr, false, nil
	}

	return modelsPtr[:limit], true, nil
}

func Transaction(ctx context.Context, db *bun.DB, fn func(ctx context.Context, tx bun.Tx) error) error {
	return db.RunInTx(ctx, nil, fn)
}

func NewDBConn(driver string, dataSourceName string, poolMax int, printQueries bool) (*bun.DB, error) {
	driver = strings.ToLower(driver)
	var dialect schema.Dialect
	{
		if driver == "pg" {
			dialect = pgdialect.New()
		}

		if driver == sqliteshim.ShimName {
			dialect = sqlitedialect.New()
		}

		if dialect == nil {
			return nil, fmt.Errorf("unknown database driver %s", driver)
		}
	}

	conn, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	db := bun.NewDB(conn, dialect, bun.WithDiscardUnknownColumns())

	// Verify successful connection
	if db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	// Configure query logging
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(printQueries)))

	// Set connection pool settings
	db.SetMaxOpenConns(poolMax * runtime.GOMAXPROCS(0))
	db.SetMaxIdleConns(poolMax * runtime.GOMAXPROCS(0))

	return db, nil
}
