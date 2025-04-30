package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/michaelhass/cpaw/models"
)

func createTestSessionRepository(t *testing.T, name string) (*SessionRepository, error) {
	db, err := prepareTestDb(name)
	t.Cleanup(cleanUpTestDb(name, db))
	return NewSessionRespository(db), err
}

func TestSessionRepository(t *testing.T) {
	dbName := "SessionRepositoryTest_createUser.db"
	sessionRepo, err := createTestSessionRepository(t, dbName)
	userRepo := NewUserRepository(sessionRepo.db)
	t.Cleanup(cleanUpTestDb(dbName, sessionRepo.db))
	if err != nil {
		t.Error(err)
		return
	}

	sessionRepoTestFunc := func(f func(*testing.T, models.User)) func(*testing.T) {
		return func(t *testing.T) {
			t.Cleanup(func() {
				userRepo.DeleteAll(context.Background())
			})
			testUser, err := userRepo.CreateUser(context.Background(), CreateUserParams{
				UserName: "Some Name",
				Password: "pw",
			})
			if err != nil {
				t.Error(err)
				return
			}
			f(t, testUser)
		}
	}

	t.Run("CreateSession", sessionRepoTestFunc(testCreateSession(sessionRepo)))
	t.Run("GetSessionByToken", sessionRepoTestFunc(testGetSessionByToken(sessionRepo)))
	t.Run("DeleteSession", sessionRepoTestFunc(testDeleteSession(sessionRepo)))
	t.Run("DeleteExpiredSessions", sessionRepoTestFunc(testDeleteExpiredSessions(sessionRepo)))
}

func testCreateSession(repo *SessionRepository) func(*testing.T, models.User) {
	return func(t *testing.T, testUser models.User) {
		ctx := context.Background()

		params := CreateSessionParams{
			Token:     "token123",
			ExpiresAt: time.Now().Add(time.Minute * 15),
			UserId:    "",
		}
		notFoundSession, err := repo.CreateSession(ctx, params)
		if err == nil || len(notFoundSession.Token) > 0 {
			t.Error("Expected error for missing user id")
			return
		}

		params.UserId = testUser.Id
		session, err := repo.CreateSession(ctx, params)
		if err != nil {
			t.Error(err)
			return
		}
		if session.Token != params.Token {
			t.Errorf("'Token' not stored correctly. Expected: %s. Got: %s.", params.Token, session.Token)
			return
		}
		if session.ExpiresAt != params.ExpiresAt.Unix() {
			t.Errorf("'ExpiredAt' not stored correctly. Expected: %s. Got: %s.", params.Token, session.Token)
			return
		}
		if session.UserId != params.UserId {
			t.Errorf("'UserId' not stored correctly. Expected: %s. Got: %s.", params.Token, session.Token)
			return
		}
	}
}

func testGetSessionByToken(repo *SessionRepository) func(*testing.T, models.User) {
	return func(t *testing.T, testUser models.User) {
		ctx := context.Background()

		expectParams := CreateSessionParams{
			Token:     "token123",
			ExpiresAt: time.Now().Add(time.Minute * 15),
			UserId:    testUser.Id,
		}

		for i := range 5 {
			copy := expectParams
			copy.Token = fmt.Sprintf("other_%d", i)
			_, _ = repo.CreateSession(ctx, copy)
		}

		_, err := repo.GetSessionByToken(ctx, expectParams.Token)
		if !errors.Is(err, ErrNotFound) {
			t.Error("Expected 'ErrNotFound'.Got: ", err)
			return
		}

		_, _ = repo.CreateSession(ctx, expectParams)
		session, err := repo.GetSessionByToken(ctx, expectParams.Token)
		if err != nil {
			t.Error(err)
			return
		}
		if session.Token != expectParams.Token {
			t.Errorf("'Token' not correct. Expected: %s. Got: %s.", expectParams.Token, session.Token)
			return
		}
		if session.ExpiresAt != expectParams.ExpiresAt.Unix() {
			t.Errorf("'ExpiredAt' not correct. Expected: %s. Got: %s.", expectParams.Token, session.Token)
			return
		}
		if session.UserId != expectParams.UserId {
			t.Errorf("'UserId' not corrected. Expected: %s. Got: %s.", expectParams.Token, session.Token)
			return
		}
	}
}

func testDeleteSession(repo *SessionRepository) func(*testing.T, models.User) {
	return func(t *testing.T, testUser models.User) {
		ctx := context.Background()

		paramsOne := CreateSessionParams{
			Token:     "token1",
			ExpiresAt: time.Now().Add(time.Minute * 15),
			UserId:    testUser.Id,
		}

		paramsTwo := CreateSessionParams{
			Token:     "token2",
			ExpiresAt: time.Now().Add(time.Minute * 15),
			UserId:    testUser.Id,
		}

		_, _ = repo.CreateSession(ctx, paramsOne)
		_, _ = repo.CreateSession(ctx, paramsTwo)

		err := repo.DeleteSessionWithToken(ctx, paramsOne.Token)
		if err != nil {
			t.Error(err)
		}

		_, err = repo.GetSessionByToken(ctx, paramsOne.Token)
		if !errors.Is(err, ErrNotFound) {
			t.Error("Session not deleted", err)
		}

		_, err = repo.GetSessionByToken(ctx, paramsTwo.Token)
		if errors.Is(err, ErrNotFound) {
			t.Error("Session should not have been deleted")
		}
	}
}

func testDeleteExpiredSessions(repo *SessionRepository) func(*testing.T, models.User) {
	return func(t *testing.T, testUser models.User) {
		ctx := context.Background()

		paramsOne := CreateSessionParams{
			Token:     "token1",
			ExpiresAt: time.Now().Add(time.Minute * -1),
			UserId:    testUser.Id,
		}

		paramsTwo := CreateSessionParams{
			Token:     "token2",
			ExpiresAt: time.Now().Add(time.Minute * 15),
			UserId:    testUser.Id,
		}

		paramsThree := CreateSessionParams{
			Token:     "token3",
			ExpiresAt: time.Now().Add(time.Second * -10),
			UserId:    testUser.Id,
		}

		_, _ = repo.CreateSession(ctx, paramsOne)
		_, _ = repo.CreateSession(ctx, paramsTwo)
		_, _ = repo.CreateSession(ctx, paramsThree)

		err := repo.DeleteExpired(ctx)
		if err != nil {
			t.Error(err)
		}

		_, err = repo.GetSessionByToken(ctx, paramsOne.Token)
		if !errors.Is(err, ErrNotFound) {
			t.Error("Session not deleted", err)
		}

		_, err = repo.GetSessionByToken(ctx, paramsTwo.Token)
		if errors.Is(err, ErrNotFound) {
			t.Error("Session should not have been deleted")
		}

		_, err = repo.GetSessionByToken(ctx, paramsThree.Token)
		if !errors.Is(err, ErrNotFound) {
			t.Error("Session not deleted", err)
		}
	}
}
