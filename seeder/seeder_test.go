package seeder

import (
	"context"
	"database/sql"
	"log"
	"testing"

	dbstore "github.com/otyang/go-dbstore"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

var (
	test_driver = dbstore.DriverSqlite
	test_dsn    = "file::memory:?cache=shared"
)

type (
	Animal struct {
		Id   string `bun:",pk"`
		Name string `bun:",notnull"`
	}

	Car struct {
		Id   string `bun:",pk"`
		Area int    `bun:",notnull"`
	}

	AnimalToCar struct {
		Id    string `bun:",pk"`
		Brand string `bun:",notnull"`
	}
)

var (
	mmodels   = []any{(*Animal)(nil), (*Car)(nil)}
	imodels   = []any{(*AnimalToCar)(nil)}
	allModels = append(mmodels, imodels...)
)

func setUp(t *testing.T, test_driver dbstore.DBDriver, test_dsn string) (context.Context, *bun.DB) {
	ctx := context.Background()

	// connect
	db, err := dbstore.NewDBConnection(test_driver, test_dsn, 1, true)
	assert.NoError(t, err)

	return ctx, db
}

func tearDown(db *bun.DB, modelsPtr ...any) {
	for _, model := range modelsPtr {
		_, err := db.NewDropTable().Model(model).Exec(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func TestSeeder_CreateTables(t *testing.T) {
	ctx, db := setUp(t, test_driver, test_dsn)
	defer tearDown(db, allModels...)

	err := CreateTables(ctx, db, mmodels, imodels)
	assert.NoError(t, err)

	a := Animal{Id: "987654321"}
	err = db.NewSelect().Model(&a).WherePK().Scan(ctx)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSeeder_DropTables(t *testing.T) {
	ctx, db := setUp(t, test_driver, test_dsn)

	err := CreateTables(ctx, db, mmodels, imodels)
	assert.NoError(t, err)

	err = DropTables(ctx, db, mmodels, imodels)
	assert.NoError(t, err)
}

func TestSeeder_DropAndCreateTables(t *testing.T) {
	ctx, db := setUp(t, test_driver, test_dsn)

	err := CreateTables(ctx, db, mmodels, imodels)
	assert.NoError(t, err)

	err = DropAndCreateTables(ctx, db, mmodels, imodels)
	assert.NoError(t, err)
}

func TestSeeder_CreateIndex(t *testing.T) {
	type Movies struct {
		Id   string `bun:",pk"`
		ISBN string `bun:",notnull"`
	}

	ctx, db := setUp(t, test_driver, test_dsn)
	defer tearDown(db, (*Movies)(nil))

	err := CreateTables(ctx, db, []any{(*Movies)(nil)}, nil)
	assert.NoError(t, err)

	err = CreateIndex(ctx, db, (*Movies)(nil), "animal_id_index", "animal_id_index")
	assert.NoError(t, err)
}
