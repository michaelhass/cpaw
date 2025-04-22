package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/michaelhass/cpaw/db/repository"
	"github.com/michaelhass/cpaw/hash"
	"github.com/michaelhass/cpaw/models"
)

const (
	DefaultSessionDuration    time.Duration = time.Minute * 30
	DefaultSessionTokenLength int           = 32
	DefaultMinPasswordLength  int           = 6
)

var (
	ErrExpiredSession    = errors.New("Expired Session")
	ErrMinPasswordLength = errors.New("Password should be min. 6 charachters long")
)

type AuthService struct {
	sessions *repository.SessionRepository
	users    *repository.UserRepository
}

func NewAuthService(
	sessions *repository.SessionRepository,
	users *repository.UserRepository,
) *AuthService {
	return &AuthService{sessions: sessions, users: users}
}

type InvalidCredentialsError struct{}

func (e *InvalidCredentialsError) Error() string {
	return "Invalid credentials"
}

type AuthSetupCredentials struct {
	Id       string
	UserName string
	Password string
}

func (as *AuthService) SetUp(ctx context.Context) (AuthSetupCredentials, error) {
	var credentials AuthSetupCredentials

	count, err := as.users.GetUserCount(ctx)
	if err != nil {
		return credentials, err
	}
	if count > 0 {
		return credentials, nil
	}

	tmpPassword := "root"
	createInitialUserParams := repository.CreateUserParams{
		UserName: "root",
		Password: tmpPassword,
		Role:     models.AdminRole,
	}

	user, err := as.users.CreateUser(ctx, createInitialUserParams)

	if err != nil {
		return credentials, err
	}

	credentials.Id = user.Id
	credentials.UserName = user.UserName
	credentials.Password = tmpPassword

	return credentials, err
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

func (as *AuthService) SignOut(ctx context.Context, sessionToken string) error {
	return as.sessions.DeleteSessionWithToken(ctx, sessionToken)
}

func (as *AuthService) GetUserById(ctx context.Context, userId string) (models.User, error) {
	return as.users.GetUserById(ctx, userId)
}

func (as *AuthService) VerifyToken(ctx context.Context, sessionToken string) (models.Session, error) {
	session, err := as.sessions.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		return models.Session{}, err
	}
	if IsSessionExpired(session) {
		return models.Session{}, ErrExpiredSession
	}
	return session, nil
}

type UpdatePasswordParams = repository.UpdateUserPasswordParams

func (as *AuthService) UpdatePassword(ctx context.Context, params UpdatePasswordParams) error {
	if len(params.Password) < DefaultMinPasswordLength {
		return ErrMinPasswordLength
	}
	return as.users.UpdatePassword(ctx, params)
}

type CreateUserParams = repository.CreateUserParams

func (as *AuthService) CreateUser(ctx context.Context, params CreateUserParams) (models.User, error) {
	return as.users.CreateUser(ctx, params)
}

func (as *AuthService) ListUsers(ctx context.Context) ([]models.User, error) {
	return as.users.ListUsers(ctx)
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
