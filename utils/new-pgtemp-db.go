package utils

import (
	"github.com/Apollo-group-io/pgtest"
)

func startDb(dbRootDir string, persistent bool) (*pgtest.PG, error) {
	if persistent {
		// start the database in the temporary directory
		return pgtest.New().DataDir(dbRootDir).Persistent().EnableFSync().Start()
	}
	return pgtest.New().DataDir(dbRootDir).Start()
}

func StartPgTempDb(dbRootDir string, persistent bool) (*pgtest.PG, error) {

	// start the database in the temporary directory
	db, err := startDb(dbRootDir, persistent)
	if err != nil {
		return nil, err
	}
	// run a query to block until the database is ready
	// instead of sleeping for unknown time.
	db.DB.Query("SELECT 1")
	return db, nil
}
