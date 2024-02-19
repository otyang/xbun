package xbun

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	dbstore "github.com/otyang/go-dbstore"
	"github.com/otyang/go-dbstore/filter"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

// Use constants for configuration
const (
	testDBDriver = dbstore.DriverSqlite
	testDSN      = "file::memory:?cache=shared"
)

var seed = []Book{
	{Id: "1", Title: "Title 1"},
	{Id: "2", Title: "Title 2"},
	{Id: "3", Title: "Title 3"},
	{Id: "4", Title: "Title 4"},
}

type Book struct {
	Id    string `bun:",pk"`
	Title string `bun:",notnull"`
}

func setUpMigrateAndTearDown(t *testing.T, modelsPtr ...any) (context.Context, *bun.DB, func()) {
	ctx := context.TODO()

	// connect
	db, err := dbstore.NewDBConnection(testDBDriver, testDSN, 1, true)
	assert.NoError(t, err)

	// migrate
	for _, model := range modelsPtr {
		_, err := db.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
		assert.NoError(t, err)
	}

	// tearDown
	teardownFunc := func() {
		for _, model := range modelsPtr {
			_, err := db.NewDropTable().Model(model).Exec(ctx)
			assert.NoError(t, err)
		}
	}

	return ctx, db, teardownFunc
}

func TestRepository_Create(t *testing.T) {
	ctx, db, tearDown := setUpMigrateAndTearDown(t, (*Book)(nil))
	defer tearDown()

	data := Book{
		Id:    "_1234asdf",
		Title: "the unknown",
	}

	err := Create(ctx, db, &data, false)
	assert.NoError(t, err)

	// re-inserting the same data should create an error
	// since the primary key already exists
	err = Create(ctx, db, &data, false)
	assert.Error(t, err)

	// ignore duplicates
	err = Create(ctx, db, &data, true)
	assert.NoError(t, err)
}

func TestRepository_CreateBulk(t *testing.T) {
	ctx, db, tearDown := setUpMigrateAndTearDown(t, (*Book)(nil))
	defer tearDown()

	err := CreateBulk(ctx, db, &seed, false)
	assert.NoError(t, err)

	err = CreateBulk(ctx, db, &seed, false)
	assert.Error(t, err)

	// ignore duplicates
	err = CreateBulk(ctx, db, &seed, true)
	assert.NoError(t, err)
}

func TestRepository_SelectOneByPK(t *testing.T) {
	var (
		ctx, db, tearDown = setUpMigrateAndTearDown(t, (*Book)(nil))
		err               = CreateBulk(ctx, db, &seed, false)
	)

	defer tearDown()
	assert.NoError(t, err)

	data := Book{Id: "1"}
	err = SelectOneByPK(ctx, db, &data)

	assert.NoError(t, err)
	assert.Equal(t, seed[0], data)
}

func TestRepository_SelectOneWhere(t *testing.T) {
	var (
		ctx, db, tearDown = setUpMigrateAndTearDown(t, (*Book)(nil))
		err               = CreateBulk(ctx, db, &seed, false)
		bookFromDB        Book
	)

	defer tearDown()
	assert.NoError(t, err)

	err = SelectOneWhere(ctx, db, &bookFromDB)
	assert.NoError(t, err)
	assert.NotEmpty(t, bookFromDB.Id)

	err = SelectOneWhere(ctx, db, &bookFromDB, func(q *bun.SelectQuery) *bun.SelectQuery {
		filter.Where(q, filter.Equal("id", "2"))
		return q
	})
	assert.NoError(t, err)
	assert.Equal(t, seed[1], bookFromDB)
}

func TestRepository_SelectManyWhere(t *testing.T) {
	var (
		ctx, db, tearDown = setUpMigrateAndTearDown(t, (*Book)(nil))
		err               = CreateBulk(ctx, db, &seed, false)
	)

	defer tearDown()
	assert.NoError(t, err)

	t.Run("SelectManyWhere without select criterias", func(t *testing.T) {
		books, err := SelectManyWhere[Book](ctx, db, 100, nil)
		assert.NoError(t, err)
		assert.Equal(t, seed, books)
		assert.Equal(t, len(seed), len(books))
	})

	t.Run("SelectManyWhere with select criterias", func(t *testing.T) {
		got, err := SelectManyWhere[Book](ctx, db, 100, func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("id >= ?", 2)
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, len(got))
	})
}

func TestRepository_UpdateByPK_oneAndMany(t *testing.T) {
	var (
		ctx, db, tearDown = setUpMigrateAndTearDown(t, (*Book)(nil))
		err               = CreateBulk(ctx, db, &seed, false)
	)

	defer tearDown()
	assert.NoError(t, err)

	t.Run("UpdateOne By PK", func(t *testing.T) {
		want := seed[0]
		want.Title = "Updated Title 1..."
		err := UpdateOneByPK(ctx, db, &want)
		assert.NoError(t, err)

		got := Book{Id: "1"}
		err = SelectOneByPK(ctx, db, &got)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title 1...", got.Title)
	})

	t.Run("Update Many By PK", func(t *testing.T) {
		updatedBooks := seed
		updatedBooks[2].Title = "bulk update 3"
		updatedBooks[3].Title = "bulk update 4"
		err := UpdateManyByPK(ctx, db, &updatedBooks)
		assert.NoError(t, err)

		got := Book{Id: "3"}
		err = SelectOneByPK(ctx, db, &got)
		assert.NoError(t, err)
		assert.Equal(t, "bulk update 3", got.Title)
	})

	t.Run("UpdateOneWhere", func(t *testing.T) {
		want := seed[0]
		want.Title = "one where"
		err = UpdateOneWhere(ctx, db, &want, func(q *bun.UpdateQuery) *bun.UpdateQuery {
			return q.Where("id = ?", seed[0].Id)
		})

		got := Book{Id: "1"}
		SelectOneByPK(ctx, db, &got)
		assert.NoError(t, err)
		assert.Equal(t, "one where", got.Title)
	})
}

func TestRepository_Upsert(t *testing.T) {
	var (
		ctx, db, tearDown = setUpMigrateAndTearDown(t, (*Book)(nil))
		err               = CreateBulk(ctx, db, &seed, false)
	)

	defer tearDown()
	assert.NoError(t, err)

	upsertedBooks := seed
	upsertedBooks[3].Title = "bulk update 4 9"

	err = Upsert(ctx, db, &upsertedBooks)
	assert.NoError(t, err)

	gotListOfUpsertedBooks, err := SelectManyWhere[Book](ctx, db, 200, nil)
	assert.NoError(t, err)
	assert.Equal(t, seed[3], gotListOfUpsertedBooks[3])
}

func TestRepository_DeleteByPK(t *testing.T) {
	var (
		ctx, db, tearDown = setUpMigrateAndTearDown(t, (*Book)(nil))
		err               = CreateBulk(ctx, db, &seed, false)
	)

	defer tearDown()
	assert.NoError(t, err)

	t.Run("DeleteByPK", func(t *testing.T) {
		err = DeleteByPK(ctx, db, &seed[0])
		assert.NoError(t, err)

		err = SelectOneByPK(ctx, db, &seed[0])
		assert.Equal(t, sql.ErrNoRows.Error(), err.Error())
	})

	t.Run("DeleteByPK  many", func(t *testing.T) {
		err = DeleteByPK(ctx, db, &[]Book{seed[1], seed[2]})
		assert.NoError(t, err)

		err = SelectOneByPK(ctx, db, &seed[1])
		assert.Equal(t, sql.ErrNoRows.Error(), err.Error())

		err = SelectOneByPK(ctx, db, &seed[2])
		assert.Equal(t, sql.ErrNoRows.Error(), err.Error())
	})
}

func TestRepository_DeleteWhere(t *testing.T) {
	var (
		ctx, db, tearDown = setUpMigrateAndTearDown(t, (*Book)(nil))
		err               = CreateBulk(ctx, db, &seed, false)
	)

	defer tearDown()
	assert.NoError(t, err)

	t.Run("DeleteWhere", func(t *testing.T) {
		err = DeleteWhere(ctx, db, (*Book)(nil), func(q *bun.DeleteQuery) *bun.DeleteQuery {
			filter.Where(q, filter.Equal("id", 1))
			return q
		})

		assert.NoError(t, err)

		err = SelectOneByPK(ctx, db, &seed[0])
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows.Error(), err.Error())
	})
}

func TestRepository_Transaction(t *testing.T) {
	ctx, db, tearDown := setUpMigrateAndTearDown(t, (*Book)(nil))
	defer tearDown()

	t.Run("transactions: no errors", func(t *testing.T) {
		err := Transaction(ctx, db, func(ctx context.Context, tx bun.Tx) error {
			if err := Create(ctx, tx, &seed[0], false); err != nil {
				return err
			}
			if err := Create(ctx, tx, &seed[1], false); err != nil {
				return err
			}
			return Create(ctx, tx, &seed[2], false)
		},
		)
		assert.NoError(t, err)
	})

	t.Run("transactions: with deliberate error to abort transactions", func(t *testing.T) {
		err := Transaction(ctx, db, func(ctx context.Context, tx bun.Tx) error {
			err := Create(ctx, tx, &seed[3], false)
			if err != nil {
				return err
			}

			return errors.New("deliberate-wrong-data")
		},
		)
		assert.Error(t, err)
	})
}
