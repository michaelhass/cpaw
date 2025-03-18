package db

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

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

func NewSqlite(opts ...SqliteOption) (*Sqlite, error) {
	conf := &SqliteConfig{
		dbName: "cpaw",
	}

	for _, opt := range opts {
		opt(conf)
	}

	db, err := sql.Open("sqlite3", conf.dbPath)
	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return nil, err
	}

	config := &sqlite3.Config{}
	driver, err := sqlite3.WithInstance(db, config)
	if err != nil {
		return nil, err
	}
	migration, err := migrate.NewWithInstance("ifs", sourceDriver, conf.dbName, driver)

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
	if err := s.MigrateUp(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	if _, err := s.DB.Exec(setUpStmt); err != nil {
		return err
	}
	return s.seed()
}

func (s *Sqlite) seed() error {
	return nil
}
