package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/michaelhass/cpaw/hash"
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

const userCountQuery = `
SELECT COUNT(1) FROM users;
`

func (ur *UserRepository) GetUserCount(ctx context.Context) (int, error) {
	var count int
	row := ur.db.QueryRowContext(ctx, userCountQuery)
	err := row.Scan(&count)
	return count, err
}

const createUserQuery = `
INSERT INTO users (id, created_at, user_name, password_hash, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, user_name, password_hash, role;
`

type CreateUserParams struct {
	UserName string
	Password string
	Role     models.Role
}

func (ur *UserRepository) CreateUser(ctx context.Context, arg CreateUserParams) (models.User, error) {
	var user models.User

	uuid, err := uuid.NewRandom()
	if err != nil {
		return user, err
	}

	id := uuid.String()
	createdAt := time.Now().Unix()
	var role models.Role
	if len(arg.Role) > 0 {
		role = arg.Role
	} else {
		role = models.UserRole
	}

	passwordHash, err := hash.NewFromPassword(arg.Password)
	if err != nil {
		return user, err
	}

	row := ur.db.QueryRowContext(
		ctx,
		createUserQuery,
		id, createdAt,
		arg.UserName,
		passwordHash,
		role,
	)
	err = row.Scan(&user.Id, &user.CreatedAt, &user.UserName, &user.PasswordHash, &user.Role)
	return user, err
}

const getUserByIdQuery = `
SELECT id, created_at, user_name, password_hash, role FROM users
WHERE id = $1
LIMIT 1;
`

func (ur *UserRepository) GetUserById(ctx context.Context, id string) (models.User, error) {
	row := ur.db.QueryRowContext(ctx, getUserByIdQuery, id)
	var user models.User
	err := row.Scan(&user.Id, &user.CreatedAt, &user.UserName, &user.PasswordHash, &user.Role)
	return user, err
}

const getUserByNameQuery = `
SELECT id, created_at, user_name, password_hash, role FROM users
WHERE user_name = $1
LIMIT 1;
`

func (ur *UserRepository) GetUserByName(ctx context.Context, name string) (models.User, error) {
	row := ur.db.QueryRowContext(ctx, getUserByNameQuery, name)
	var user models.User
	err := row.Scan(&user.Id, &user.CreatedAt, &user.UserName, &user.PasswordHash, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrNotFound
	}
	return user, err
}

const listUsersQuery = `
SELECT id, created_at, user_name, password_hash, role FROM users
ORDER BY user_name;
`

func (ur *UserRepository) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := ur.db.QueryContext(ctx, listUsersQuery)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Id, &user.CreatedAt, &user.UserName, &user.PasswordHash, &user.Role); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

type UpdateUserPasswordParams struct {
	Id       string
	Password string
}

const updatePasswordQuery = `
UPDATE users
SET password_hash = $1
WHERE id = $2
`

func (ur *UserRepository) UpdatePassword(ctx context.Context, args UpdateUserPasswordParams) error {
	passwordHash, err := hash.NewFromPassword(args.Password)
	if err != nil {
		return err
	}
	_, err = ur.db.ExecContext(ctx, updatePasswordQuery, passwordHash, args.Id)
	return err
}

const deleteUserByIdQuery = "DELETE FROM users WHERE id = $1;"

func (ur *UserRepository) DeleteUserById(ctx context.Context, id string) error {
	_, err := ur.db.ExecContext(ctx, deleteUserByIdQuery, id)
	return err
}

const deleteAllUsersQuery = "DELETE FROM users;"

func (ur *UserRepository) DeleteAll(ctx context.Context) error {
	_, err := ur.db.ExecContext(ctx, deleteAllUsersQuery)
	return err
}
