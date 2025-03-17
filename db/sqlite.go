package db

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	DB        *sql.DB
	driver    database.Driver
	migration *migrate.Migrate
}

type SqliteConfig struct {
	dbPath              string
	dbName              string
	migrationsSourceUrl string
}

type SqliteOption func(*SqliteConfig)

func WithDbPath(path string) SqliteOption {
	return func(conf *SqliteConfig) {
		conf.dbPath = path
	}
}

func WithDbName(name string) SqliteOption {
	return func(conf *SqliteConfig) {
		conf.dbName = name
	}
}

func WithMigrationSource(sourceUrl string) SqliteOption {
	return func(conf *SqliteConfig) {
		conf.migrationsSourceUrl = sourceUrl
	}
}

func NewSqlite(opts ...SqliteOption) (*Sqlite, error) {
	conf := &SqliteConfig{
		dbName:              "cpaw",
		migrationsSourceUrl: "file://./db/migrations",
	}
	for _, opt := range opts {
		opt(conf)
	}

	db, err := sql.Open("sqlite3", conf.dbPath)
	config := &sqlite3.Config{}
	driver, err := sqlite3.WithInstance(db, config)
	if err != nil {
		return nil, err
	}
	migration, err := migrate.NewWithDatabaseInstance(
		conf.migrationsSourceUrl,
		conf.dbName,
		driver,
	)
	return &Sqlite{
		DB:        db,
		driver:    driver,
		migration: migration,
	}, err
}

func (s *Sqlite) MigrateUp() error {
	return s.migration.Up()
}

func (s *Sqlite) MigrateDown() error {
	return s.migration.Up()
}

func (s *Sqlite) Close() error {
	return s.driver.Close()
}

const setUpStmt = `
	PRAGMA foreign_keys = ON;
`

func (s *Sqlite) SetUp() error {
	_, err := s.DB.Exec(setUpStmt)
	return err
}

func (s *Sqlite) Seed() error {
	return nil
}
