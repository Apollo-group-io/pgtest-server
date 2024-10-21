package utils

import (
	"fmt"

	"github.com/rubenv/pgtest"
)

func StartPGTestDB(dbRootDir string, enableFsync bool) (*pgtest.PG, error) {
	config := pgtest.New().DataDir(dbRootDir)
	if enableFsync {
		config = config.EnableFSync()
	}
	db, err := config.Start()
	if err != nil {
		return nil, fmt.Errorf("error starting pgtest db: %s", err)
	}
	db.DB.Query("SELECT 1")
	return db, nil
}
