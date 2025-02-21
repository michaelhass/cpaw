package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func NewSqlite(databasePath string) (*sql.DB, error) {
	return sql.Open("sqlite3", databasePath)
}
