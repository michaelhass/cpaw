package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	DB *sql.DB
}

func NewSqlite(databasePath string) (*Sqlite, error) {
	db, err := sql.Open("sqlite3", databasePath)
	return &Sqlite{DB: db}, err
}

func (s *Sqlite) Close() error {
	return s.DB.Close()
}

const setUpStmt = `
	PRAGMA foreign_keys = ON;
`

func (s *Sqlite) SetUp() error {
	_, err := s.DB.Exec(setUpStmt)
	return err
}

const insertRolesStmt = `
INSERT OR IGNORE INTO roles(name) VALUES("admin");
INSERT OR IGNORE INTO roles(name) VALUES("user");
`

func (s *Sqlite) Seed() error {
	_, err := s.DB.Exec(insertRolesStmt)
	return err
}
