package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/michaelhass/cpaw/hash"
	"github.com/michaelhass/cpaw/models"
)

func createTestUserRepository(t *testing.T, name string) (*UserRepository, error) {
	db, err := prepareTestDb(name)
	t.Cleanup(cleanUpTestDb(name, db))
	return NewUserRepository(db), err
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
	t.Run("UpdateName", userRepoTestFunc(testUpdateUserName(repo)))
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

		_, err := repo.GetUserByName(ctx, "non_existing_id")
		if !errors.Is(err, ErrNotFound) {
			t.Error("Wrong error for not found", err)
			return
		}
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

		_, err := repo.GetUserByName(ctx, "non_existing_user")
		if !errors.Is(err, ErrNotFound) {
			t.Error("Wrong error for not found", err)
			return
		}
		expectUsers, err := createTestUsers(ctx, repo, 5)
		if err != nil {
			return
		}

		for _, expectUser := range expectUsers {
			gotUser, err := repo.GetUserByName(ctx, expectUser.UserName)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(expectUser, gotUser) {
				t.Errorf("Users did not match. Expected: %v. Got: %v", expectUser, gotUser)
				return
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
			return
		}

		updatedPw := "updated_pw"
		err = repo.UpdatePassword(ctx, UpdateUserPasswordParams{
			UserId:   initialUser.Id,
			Password: updatedPw,
		})
		if err != nil {
			t.Error(err)
			return
		}

		updatedUser, err := repo.GetUserById(ctx, initialUser.Id)
		if initialUser.PasswordHash == updatedUser.PasswordHash &&
			!hash.VerifyPassword(updatedPw, updatedUser.PasswordHash) {
			t.Error("Unable to to update password:", updatedUser)
		}
	}
}

func testUpdateUserName(repo *UserRepository) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		users, err := createTestUsers(ctx, repo, 2)
		if err != nil {
			t.Error(err)
			return
		}

		updateUser := users[0]
		otherUser := users[1]

		err = repo.UpdateUserName(ctx, UpdateUserNameParams{
			UserName: otherUser.UserName,
			UserId:   updateUser.Id,
		})

		if err == nil {
			t.Error("Expected error because with non unique user name")
			return
		}

		newUserName := "NEW_NAME"
		err = repo.UpdateUserName(ctx, UpdateUserNameParams{
			UserName: newUserName,
			UserId:   updateUser.Id,
		})

		if err != nil {
			t.Error(err)
			return
		}

		updatedUser, err := repo.GetUserById(ctx, updateUser.Id)
		if err != nil {
			t.Error(err)
			return
		}

		if updatedUser.UserName != newUserName {
			t.Error("Did not update user name")
		}
	}
}
