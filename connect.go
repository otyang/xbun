package xbun

import (
	"database/sql"
	"fmt"
	"runtime"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/schema"
)

// NewDBConnection establishes a connection to the database.
func NewDBConnection(driver DBDriver, dataSourceName string, poolMax int, printQueries bool) (*bun.DB, error) {
	if !driver.IsValid() {
		return nil, fmt.Errorf("unknown database driver %s", driver.String())
	}

	conn, err := sql.Open(driver.String(), dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	db := bun.NewDB(conn, driver.GetDialect(), bun.WithDiscardUnknownColumns())

	// Verify successful connection
	if db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	// Configure query logging
	if printQueries {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	// Set connection pool settings
	db.SetMaxOpenConns(poolMax * runtime.GOMAXPROCS(0))
	db.SetMaxIdleConns(poolMax * runtime.GOMAXPROCS(0))

	return db, nil
}

// DBDriver represents the supported database drivers.
type DBDriver string

const (
	DriverPostgresql DBDriver = "pg"
	DriverSqlite     DBDriver = sqliteshim.ShimName
)

func (d DBDriver) String() string {
	return string(d)
}

// IsValid checks if the provided driver is supported.
func (d DBDriver) IsValid() bool {
	switch d {
	case DriverSqlite, DriverPostgresql:
		return true
	}
	return false
}

// GetDialect returns the corresponding dialect for the driver.
func (d DBDriver) GetDialect() schema.Dialect {
	switch d {
	case DriverSqlite:
		return sqlitedialect.New()
	case DriverPostgresql:
		return pgdialect.New()
	}

	panic(fmt.Errorf("unknown db driver dialect: %s", d.String()))
}
