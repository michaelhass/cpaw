package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/michaelhass/cpaw/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

const createUserQuery = `
INSERT INTO users (id, created_at, user_name, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING (id, created_at, user_name, password_hash)
`

type CreateUserParams struct {
	UserName     string
	PasswordHash string
}

func (ur *UserRepository) CreateUser(ctx context.Context, arg CreateUserParams) (models.User, error) {
	var user models.User

	uuid, err := uuid.NewRandom()
	if err != nil {
		return user, err
	}

	id := uuid.String()
	createdAt := time.Now().Unix()

	row := ur.db.QueryRowContext(ctx, createUserQuery, id, createdAt, arg.UserName, arg.PasswordHash)
	err = row.Scan(&user.Id, &user.CreatedAt, &user.UserName, &user.PasswordHash)
	return user, err
}

const getUserByIdQuery = `
SELECT id, created_at, user_name, password_hash FROM users
WHERE id = $1
LIMIT 1;
`

func (ur *UserRepository) GetUserById(ctx context.Context, id string) (models.User, error) {
	row := ur.db.QueryRowContext(ctx, getUserByIdQuery, id)
	var user models.User
	err := row.Scan(&user.Id, &user.CreatedAt, &user.UserName, &user.PasswordHash)
	return user, err
}

const getAllUsersQuery = `
SELECT id, created_at, user_name, password_hash FROM users
ORDER BY user_name
`

func (ur *UserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	rows, err := ur.db.QueryContext(ctx, getAllUsersQuery)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Id, &user.CreatedAt, &user.UserName, &user.PasswordHash); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

const deleteUserByIdQuery = `
DELETE FROM users
WHERE id = $1
`

func (ur *UserRepository) DeleteUserById(ctx context.Context, id string) error {
	_, err := ur.db.ExecContext(ctx, deleteUserByIdQuery, id)
	return err
}
