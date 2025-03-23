package repository

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"

	"github.com/michaelhass/cpaw/db"
)

const (
	dbTestDir string = "../../tmp/tests/"
)

func dbTestPath(name string) string {
	return fmt.Sprintf("%s%s.db", dbTestDir, name)
}

func cleanUpTestDb(name string, db *sql.DB) func() {
	return func() {
		if db != nil {
			_ = db.Close()
		}
		_ = os.Remove(dbTestPath(name))
	}
}

func prepareTestDb(name string) (*sql.DB, error) {
	os.MkdirAll(dbTestDir, fs.ModePerm)
	sqlite, err := db.NewSqlite(
		db.WithDbName(name),
		db.WithDbPath(dbTestPath(name)),
	)
	if err != nil {
		return nil, err
	}
	err = sqlite.SetUp()
	if err != nil {
		return nil, err
	}
	return sqlite.DB, nil
}
