package xbun

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/sqliteshim"
)

var (
	test_driver = sqliteshim.ShimName
	test_dsn    = "file::memory:?cache=shared"
)

func setUpMigrateAndTearDown(driver string, dataSourceName string, poolMax int, printQueries bool, modelsPtr ...any,
) (*bun.DB, func()) {
	// connect
	db, err := NewDBConn(driver, dataSourceName, poolMax, printQueries)
	if err != nil {
		log.Fatal(err)
	}

	// migrate
	for _, model := range modelsPtr {
		_, err := db.NewCreateTable().Model(model).IfNotExists().Exec(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}

	// tearDown
	teardownFunc := func() {
		for _, model := range modelsPtr {
			if model == nil {
				continue
			}
			_, err := db.NewDropTable().Model(model).Exec(context.TODO())
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return db, teardownFunc
}

type Book struct {
	Id    string `bun:",pk"`
	Title string `bun:",notnull"`
}

var seed = []Book{
	{Id: "1", Title: "Title 1"},
	{Id: "2", Title: "Title 2"},
	{Id: "3", Title: "Title 3"},
	{Id: "4", Title: "Title 4"},
}

func TestCreate(t *testing.T) {
	// same process for single or bulk (slice of type)
	var (
		ctx          context.Context = context.Background()
		db, tearDown                 = setUpMigrateAndTearDown(test_driver, test_dsn, 1, true, (*Book)(nil))
	)
	defer tearDown()

	data := Book{Id: "123", Title: "the unknown"}

	err := Create(ctx, db, false, &data) // create fresh
	assert.NoError(t, err)

	err = Create(ctx, db, false, &data) // dont ignore duplicates. re-inserting should  error
	assert.Error(t, err)

	err = Create(ctx, db, true, &data) // ignore duplicates
	assert.NoError(t, err)

	// ================================================== bulk
	err = Create(ctx, db, false, &[]Book{seed[0], seed[1]}) // create fresh
	assert.NoError(t, err)

	err = Create(ctx, db, false, &[]Book{seed[0], seed[1]}) // dont ignore duplicates
	assert.Error(t, err)

	err = Create(ctx, db, true, &[]Book{seed[0], seed[1]}) // ignore duplicates
	assert.NoError(t, err)
}

func TestFindOne(t *testing.T) {
	var (
		ctx          context.Context = context.Background()
		db, tearDown                 = setUpMigrateAndTearDown(test_driver, test_dsn, 1, true, (*Book)(nil))
		err                          = Create(ctx, db, false, &seed)
	)
	defer tearDown()
	assert.NoError(t, err)

	m, err := FindOne(ctx, db, &Book{Id: "1"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, seed[0], m)

	// ============================================================ where
	m, err = FindOne(ctx, db, &Book{Id: "2"}, func(q *bun.SelectQuery) *bun.SelectQuery {
		q.Where("id = ?", 1)
		return q
	})
	assert.NoError(t, err)
	assert.Equal(t, seed[0], m)
}

func TestFindMany(t *testing.T) {
	var (
		ctx          context.Context = context.Background()
		db, tearDown                 = setUpMigrateAndTearDown(test_driver, test_dsn, 1, true, (*Book)(nil))
		err                          = Create(ctx, db, false, &seed)
	)
	defer tearDown()
	assert.NoError(t, err)

	results, hasMore, err := FindMany[Book](ctx, db, 1, nil)
	assert.NoError(t, err)
	assert.True(t, hasMore)
	assert.True(t, len(results) == 1)

	results, hasMore, err = FindMany[Book](ctx, db, 4, nil)
	assert.NoError(t, err)
	assert.False(t, hasMore)
	assert.True(t, len(results) == 4)
}

func TestUpsert(t *testing.T) {
	// same process for single or bulk (slice of type)
	var (
		createBook                   = Book{Id: "a1", Title: "the book title"}
		ctx          context.Context = context.Background()
		db, tearDown                 = setUpMigrateAndTearDown(test_driver, test_dsn, 1, true, (*Book)(nil))
		err                          = Create(ctx, db, false, &seed)
	)
	defer tearDown()
	assert.NoError(t, err)

	createBook.Id = "a1"

	n, err := Upsert(ctx, db, &createBook)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestUpdate(t *testing.T) {
	// same process for single or bulk (slice of type)
	var (
		ctx          context.Context = context.Background()
		db, tearDown                 = setUpMigrateAndTearDown(test_driver, test_dsn, 1, true, (*Book)(nil))
		err                          = Create(ctx, db, false, &seed)
	)
	defer tearDown()
	assert.NoError(t, err)

	updatedBooks := seed
	updatedBooks[2].Id = "update2"
	updatedBooks[3].Id = "update3"

	n, err := Upsert(ctx, db, &updatedBooks)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), n)
}

func TestDelete(t *testing.T) {
	// same process for single or bulk (slice of type)
	var (
		ctx          context.Context = context.Background()
		db, tearDown                 = setUpMigrateAndTearDown(test_driver, test_dsn, 1, true, (*Book)(nil))
		err                          = Create(ctx, db, false, &seed)
	)
	defer tearDown()
	assert.NoError(t, err)

	n, err := Delete(ctx, db, &seed[0], nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)

	// ============================================================ multiple

	n, err = Delete(ctx, db, &[]Book{seed[1], seed[2]}, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), n)

	// ============================================================ where

	n, err = Delete(ctx, db, &seed[3], func(q *bun.DeleteQuery) *bun.DeleteQuery {
		q.Where("id = ?", 45)
		return q
	})
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}
