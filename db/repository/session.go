package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/michaelhass/cpaw/models"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRespository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

type CreateSessionParams struct {
	ExpiresAt time.Time
	UserId    string
}

const createSessionQuery = `
INSERT INTO sessions (id, expires_at, user_id)
VALUES ($1, $2, $3)
RETURNING (id, expires_at, user_id);
`

func (sr *SessionRepository) CreateSession(ctx context.Context, arg CreateSessionParams) (models.Session, error) {
	var session models.Session

	uuid, err := uuid.NewRandom()
	if err != nil {
		return session, err
	}
	sessionId := uuid.String()
	expiresAt := arg.ExpiresAt.Unix()

	row := sr.db.QueryRowContext(
		ctx,
		createSessionQuery,
		sessionId,
		expiresAt,
		arg.UserId,
	)

	err = row.Scan(&session.Id, &session.ExpiresAt, session.UserId)
	return session, err
}

const getSessionForUserQuery = `
SELECT (id, expires_at, user_id) FROM sessions
WHERE user_id = $1;
`

func (sr *SessionRepository) GetSessionForUser(ctx context.Context, userId string) (models.Session, error) {
	var session models.Session
	row := sr.db.QueryRowContext(
		ctx,
		getSessionForUserQuery,
		userId,
	)
	err := row.Scan(&session.Id, &session.ExpiresAt, session.UserId)
	return session, err
}

const deleteSessionForUserQuery = "DELETE FROM sessions WHERE id = $1;"

func (sr *SessionRepository) DeleteSessionForUser(ctx context.Context, userId string) error {
	_, err := sr.db.ExecContext(ctx, deleteSessionForUserQuery, userId)
	return err
}
