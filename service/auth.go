package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/michaelhass/cpaw/db/repository"
	"github.com/michaelhass/cpaw/hash"
	"github.com/michaelhass/cpaw/models"
)

const (
	DefaultSessionDuration    time.Duration = time.Minute * 15
	DefaultSessionTokenLength int           = 32
	DefaultMinPasswordLength  int           = 6
	DefaultCleanUpInterval    time.Duration = time.Minute * 1
)

var (
	ErrExpiredSession    = errors.New("Expired Session")
	ErrMinPasswordLength = errors.New("Password should be min. 6 characters long")
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

func (as *AuthService) SetUp(
	ctx context.Context,
	createInitialUser func() CreateUserParams,
) (models.User, error) {
	count, err := as.users.GetUserCount(ctx)
	if err != nil {
		return models.User{}, err
	}
	if count > 0 {
		return models.User{}, nil
	}
	params := createInitialUser()
	if len(params.UserName) == 0 {
		return models.User{}, errors.New("Empty user name")
	}
	if len(params.Password) < DefaultMinPasswordLength {
		return models.User{}, ErrMinPasswordLength
	}
	params.Role = models.AdminRole
	initialUser, err := as.users.CreateUser(ctx, params)
	return initialUser, err
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
	if !as.IsValidPassword(params.Password) {
		return ErrMinPasswordLength
	}
	return as.users.UpdatePassword(ctx, params)
}

func (as *AuthService) IsValidPassword(password string) bool {
	return len(password) >= DefaultMinPasswordLength
}

type UpdateUserNameParams = repository.UpdateUserNameParams

func (as *AuthService) UpdateUserName(ctx context.Context, params UpdateUserNameParams) error {
	return as.users.UpdateUserName(ctx, params)
}

type CreateUserParams = repository.CreateUserParams

func (as *AuthService) CreateUser(ctx context.Context, params CreateUserParams) (models.User, error) {
	return as.users.CreateUser(ctx, params)
}

func (as *AuthService) ListUsers(ctx context.Context) ([]models.User, error) {
	return as.users.ListUsers(ctx)
}

func (as *AuthService) GetUserById(ctx context.Context, userId string) (models.User, error) {
	return as.users.GetUserById(ctx, userId)
}

func (as *AuthService) DeleteUserById(ctx context.Context, userId string) error {
	return as.users.DeleteUserById(ctx, userId)
}

func (as *AuthService) RunPeriodicCleanUpTask(parentContext context.Context) context.CancelFunc {
	ticker := time.NewTicker(DefaultCleanUpInterval)
	ctx, cancel := context.WithCancel(parentContext)

	log.Println("Starting AuthService clean up task")
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := as.sessions.DeleteExpired(ctx); err != nil {
					log.Println("Error deleting expired sessions", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	return cancel
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
