package xbun

import (
	"context"
	"log"

	"github.com/uptrace/bun"
)

func SetUpMigrateAndTearDown(driver string, dataSourceName string, poolMax int, printQueries bool, modelsPtr ...any,
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
