package seeder

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

var (
	ErrCreateTablesPrefix     = "create table error: %w"
	ErrDropTablesPrefix       = "drop table error: %w"
	ErrDropCreateTablesPrefix = "drop and create tables error: %w"
)

func RegisterModels(ctx context.Context, db *bun.DB, models []any, intermediaryModels []any) {
	db.RegisterModel(intermediaryModels...)
	db.RegisterModel(models...)
}

func CreateTables(ctx context.Context, db *bun.DB, models []any, intermediaryModels []any) error {
	err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		ms := append(models, intermediaryModels...)
		for _, model := range ms {
			if _, err := tx.NewCreateTable().Model(model).Exec(ctx); err != nil {
				return fmt.Errorf(ErrCreateTablesPrefix, err)
			}
		}

		return nil
	})
	return err
}

func DropTables(ctx context.Context, db *bun.DB, models []any, intermediaryModels []any) error {
	err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		ms := append(models, intermediaryModels...)
		for _, model := range ms {
			if _, err := tx.NewDropTable().Model(model).Cascade().IfExists().Exec(ctx); err != nil {
				return fmt.Errorf(ErrDropTablesPrefix, err)
			}
		}
		return nil
	})

	return err
}

func DropAndCreateTables(ctx context.Context, db *bun.DB, models []any, intermediaryModels []any) error {
	err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		ms := append(models, intermediaryModels...)

		for _, model := range ms {
			if _, err := tx.NewDropTable().Model(model).Cascade().IfExists().Exec(ctx); err != nil {
				return fmt.Errorf(ErrDropCreateTablesPrefix, err)
			}
		}
		for _, model := range ms {
			if _, err := tx.NewCreateTable().Model(model).Exec(ctx); err != nil {
				return fmt.Errorf(ErrDropCreateTablesPrefix, err)
			}
		}
		return nil
	})

	return err
}

func CreateIndex(ctx context.Context, db *bun.DB, modelPtr any, indexName string, indexColumn string) error {
	err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := db.NewCreateIndex().Model(modelPtr).Index(indexName).Column(indexColumn).Exec(ctx)
		return err
	})
	return err
}
