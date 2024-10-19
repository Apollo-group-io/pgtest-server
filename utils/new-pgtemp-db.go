package utils

import (
	"fmt"

	"github.com/Apollo-group-io/pgtest"
)

func StartTempDb(dbRootDir string) (*pgtest.PG, error) {
	db, err := pgtest.New().DataDir(dbRootDir).Start()
	if err != nil {
		return nil, fmt.Errorf("error starting temp db: %s", err)
	}
	db.DB.Query("SELECT 1")
	return db, nil
}

func StartTemplateDB(dbRootDir string) (*pgtest.PG, error) {
	// start the database in the temporary directory
	db, err := pgtest.New().DataDir(dbRootDir).Persistent().EnableFSync().Start()
	if err != nil {
		return nil, fmt.Errorf("error starting template db: %s", err)
	}
	db.DB.Query("SELECT 1")
	return db, nil
}
