package db

import (
	"database/sql"

	_ "github.com/mattn/go-slite3"
)

type SqliteConnection struct {
	*sql.DB
}

func ConnectSqlite(databasePath string) (*SqliteConnection, error) {
	db, err := sql.Open("sqlite3", databasePath)
	return &SqliteConnection{DB: db}, err
}
