package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/michaelhass/cpaw/db/repository"
	"github.com/michaelhass/cpaw/hash"
	"github.com/michaelhass/cpaw/models"
)

const (
	DefaultSessionDuration    time.Duration = time.Minute * 30
	DefaultSessionTokenLength int           = 32
)

type AuthService struct {
	sessions *repository.SessionRepository
	users    *repository.UserRepository
}

type InvalidCredentialsError struct{}

func (e *InvalidCredentialsError) Error() string {
	return "Invalid credentials"
}

type AuthSignInResult struct {
	User    models.User
	Session models.Session
}

func (as *AuthService) SignIn(ctx context.Context, userName string, password string) (AuthSignInResult, error) {
	var result AuthSignInResult

	user, err := as.users.GetUserByName(ctx, userName)
	if err != nil {
		return result, &InvalidCredentialsError{}
	}

	isMatch := hash.VerifyPassword(password, user.PasswordHash)
	if !isMatch {
		return result, &InvalidCredentialsError{}
	}

	token, err := generateSessionToken(DefaultSessionTokenLength)
	if err != nil {
		return result, err
	}

	session, err := as.sessions.CreateSession(ctx, repository.CreateSessionParams{
		Token:     token,
		ExpiresAt: newSessionExpirationTime(),
		UserId:    user.Id,
	})

	if err != nil {
		return result, err
	}

	result.Session = session
	result.User = user

	return result, nil
}

func (as *AuthService) SignOut(ctx context.Context, sessionId string) error {
	return as.sessions.DeleteSessionById(ctx, sessionId)
}

func generateSessionToken(length int) (string, error) {
	randomValues := make([]byte, length)
	if _, err := rand.Read(randomValues); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(randomValues), nil
}

func newSessionExpirationTime() time.Time {
	return time.Now().Add(DefaultSessionDuration)
}

func IsSessionExpired(session models.Session) bool {
	return time.Now().Unix() > session.ExpiresAt
}
