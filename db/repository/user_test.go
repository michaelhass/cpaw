package repository

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/michaelhass/cpaw/db"
	"github.com/michaelhass/cpaw/hash"
	"github.com/michaelhass/cpaw/models"
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

func createTestUserRepository(t *testing.T, name string) *UserRepository {
	db, err := prepareTestDb(name)
	t.Cleanup(cleanUpTestDb(name, db))
	if err != nil {
		t.Error(err)
		return nil
	}
	return NewUserRepository(db)
}

func TestUserRepository(t *testing.T) {
	dbName := "UserRepositoryTest_createUser.db"
	repo, err := createTestUserRepository(t, dbName)
	t.Cleanup(cleanUpTestDb(dbName, repo.db))
	if err != nil {
		t.Error(err)
		return
	}

	userRepoTestFunc := func(f func(*testing.T)) func(*testing.T) {
		return func(t *testing.T) {
			t.Cleanup(func() {
				repo.DeleteAll(context.Background())
			})
			f(t)
		}
	}

	t.Run("CreateUser", userRepoTestFunc(testCreateUser(repo)))
	t.Run("GetUserById", userRepoTestFunc(testGetUserById(repo)))
	t.Run("GetUserByName", userRepoTestFunc(testGetUserByName(repo)))
	t.Run("ListUsers", userRepoTestFunc(testListUsers(repo)))
	t.Run("UpdatePassword", userRepoTestFunc(testUpdatePassword(repo)))
}

func createTestUsers(ctx context.Context, repo *UserRepository, count int) ([]models.User, error) {
	users := make([]models.User, count)
	for i := range users {
		var params CreateUserParams
		params.UserName = fmt.Sprintf("TestUser_%d", i)
		params.Password = "test"
		user, err := repo.CreateUser(ctx, params)
		if err != nil {
			return users, err
		}
		users[i] = user
	}
	return users, nil
}

func testCreateUser(repo *UserRepository) func(*testing.T) {
	return func(t *testing.T) {
		params := CreateUserParams{
			UserName: "test_name",
			Password: "test_pw",
			Role:     models.AdminRole,
		}
		user, err := repo.CreateUser(context.Background(), params)
		if err != nil {
			t.Error(err)
			return
		}

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
}

func testGetUserById(repo *UserRepository) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		expectUsers, err := createTestUsers(ctx, repo, 5)
		if err != nil {
			t.Error(err)
			return
		}

		for _, expectUser := range expectUsers {
			gotUser, err := repo.GetUserById(ctx, expectUser.Id)
			if err != nil {
				t.Error("failed to get user", err)
			}
			if !reflect.DeepEqual(expectUser, gotUser) {
				t.Errorf("Users did not match. Expected: %v - got: %v", expectUser, gotUser)
			}
		}
	}
}

func testGetUserByName(repo *UserRepository) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		expectUsers, err := createTestUsers(ctx, repo, 5)
		if err != nil {
			t.Error(err)
			return
		}

		for _, expectUser := range expectUsers {
			gotUser, err := repo.GetUserByName(ctx, expectUser.UserName)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(expectUser, gotUser) {
				t.Errorf("Users did not match. Expected: %v. Got: %v", expectUser, gotUser)
			}
		}
	}
}

func testListUsers(repo *UserRepository) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		expectUsers, err := createTestUsers(ctx, repo, 5)
		if err != nil {
			t.Error(err)
			return
		}

		gotUsers, err := repo.ListUsers(ctx)
		if !reflect.DeepEqual(expectUsers, gotUsers) {
			t.Errorf("Could not retrieve all users correctly. Expected: %v. Got: %v", expectUsers, gotUsers)
		}
	}
}

func testUpdatePassword(repo *UserRepository) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		initialUser, err := repo.CreateUser(ctx, CreateUserParams{
			UserName: "Test_Upadte",
			Password: "initial_pw",
		})
		if err != nil {
			t.Error(err)
		}

		updatedPw := "updated_pw"
		err = repo.UpdatePassword(ctx, UpdateUserPasswordParams{
			Id:       initialUser.Id,
			Password: updatedPw,
		})
		if err != nil {
			t.Error(err)
		}

		updatedUser, err := repo.GetUserById(ctx, initialUser.Id)
		if initialUser.PasswordHash == updatedUser.PasswordHash &&
			!hash.VerifyPassword(updatedPw, updatedUser.PasswordHash) {
			t.Error("Unable to to update password:", updatedUser)
		}
	}
}
