package repository

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"testing"

	"github.com/michaelhass/cpaw/db"
	"github.com/michaelhass/cpaw/hash"
	"github.com/michaelhass/cpaw/models"
	"github.com/stretchr/testify/assert"
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

func createTestUserRepository(name string) (*UserRepository, error) {
	db, err := prepareTestDb(name)
	if err != nil {
		return nil, err
	}
	return NewUserRepository(db), nil
}

func TestCreateUser(t *testing.T) {
	dbName := "UserRepositoryTest_createUser.db"
	repo, err := createTestUserRepository(dbName)
	t.Cleanup(cleanUpTestDb(dbName, repo.db))
	assert.NoError(t, err, "Error creating test DB")

	params := CreateUserParams{
		UserName: "test_name",
		Password: "test_pw",
		Role:     models.AdminRole,
	}

	user, err := repo.CreateUser(context.Background(), params)
	assert.NoError(t, err, "Unable to create user")
	if len(user.Id) == 0 {
		t.Error("User id not created")
	}
	if user.UserName != params.UserName {
		t.Error("Wrong UserName")
	}
	if user.PasswordHash == params.Password ||
		!hash.VerifyPassword(params.Password, user.PasswordHash) {
		t.Error("Password not hashed correctly")
	}
	if user.Role != params.Role {
		t.Error("User role not stored correctly")
	}
}
