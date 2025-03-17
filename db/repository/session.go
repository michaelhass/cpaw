package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/michaelhass/cpaw/models"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRespository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

type CreateSessionParams struct {
	Token     string
	ExpiresAt time.Time
	UserId    string
}

const createSessionQuery = `
INSERT INTO sessions (token, expires_at, user_id)
VALUES ($1, $2, $3)
RETURNING token, expires_at, user_id;
`

func (sr *SessionRepository) CreateSession(ctx context.Context, arg CreateSessionParams) (models.Session, error) {
	var session models.Session
	expiresAt := arg.ExpiresAt.Unix()
	row := sr.db.QueryRowContext(
		ctx,
		createSessionQuery,
		arg.Token,
		expiresAt,
		arg.UserId,
	)

	err := row.Scan(&session.Token, &session.ExpiresAt, &session.UserId)
	return session, err
}

const getSessionByTokenQuery = `
SELECT (token, expires_at, user_id) FROM sessions
WHERE token = $1;
`

func (sr *SessionRepository) GetSessionByToken(ctx context.Context, sessionToken string) (models.Session, error) {
	var session models.Session
	row := sr.db.QueryRowContext(
		ctx,
		getSessionByTokenQuery,
		sessionToken,
	)
	err := row.Scan(&session.Token, &session.ExpiresAt, &session.UserId)
	if errors.Is(err, sql.ErrNoRows) {
		return session, ErrNotFound
	}
	return session, err
}

const deleteSessionWithTokenQuery = "DELETE FROM sessions WHERE token = $1;"

func (sr *SessionRepository) DeleteSessionWithToken(ctx context.Context, sessionToken string) error {
	_, err := sr.db.ExecContext(ctx, deleteSessionWithTokenQuery, sessionToken)
	return err
}
