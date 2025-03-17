package repository

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/michaelhass/cpaw/db"
)

const (
	dbTestDir                string = "../../tmp/tests/"
	dbTestMigrationSourceUrl string = "file://../migrations"
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
		db.WithMigrationSource(dbTestMigrationSourceUrl),
	)
	if err != nil {
		return nil, err
	}
	err = sqlite.SetUp()
	if err != nil {
		log.Println("bla")
		return nil, err
	}
	return sqlite.DB, nil
}
