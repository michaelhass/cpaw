package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/michaelhass/cpaw/models"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

type CreateItemParams struct {
	Content string
	UserId  string
}

const createItemQuery = `
INSERT INTO items (id, created_at, content, user_id)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, content, user_id;
`

func (ir *ItemRepository) CreateItem(ctx context.Context, arg CreateItemParams) (models.Item, error) {
	var item models.Item

	uuid, err := uuid.NewRandom()
	if err != nil {
		return item, err
	}

	id := uuid.String()
	createdAt := time.Now().Unix()

	row := ir.db.QueryRowContext(
		ctx,
		createItemQuery,
		id,
		createdAt,
		arg.Content,
		arg.UserId,
	)

	err = row.Scan(&item.Id, &item.CreatedAt, &item.Content, &item.UserId)
	return item, err
}

const getItemByIdQuery = "SELECT id, created_at, content, user_id FROM items WHERE id = $1;"

func (ir *ItemRepository) GetItemById(ctx context.Context, itemId string) (models.Item, error) {
	row := ir.db.QueryRowContext(ctx, getItemByIdQuery, itemId)
	var item models.Item
	err := row.Scan(&item.Id, &item.CreatedAt, &item.Content, &item.UserId)
	if errors.Is(err, sql.ErrNoRows) {
		return item, ErrNotFound
	}
	return item, err
}

const getItemForUserWuery = "SELECT id, created_at, content, user_id FROM items WHERE id = $1 AND user_id = $2;"

type GetItemForUserParams struct {
	ItemId string
	UserId string
}

func (ir *ItemRepository) GetItemForUser(ctx context.Context, arg GetItemForUserParams) (models.Item, error) {
	row := ir.db.QueryRowContext(ctx, getItemByIdQuery, arg.ItemId, arg.UserId)
	var item models.Item
	err := row.Scan(&item.Id, &item.CreatedAt, &item.Content, &item.UserId)
	if errors.Is(err, sql.ErrNoRows) {
		return item, ErrNotFound
	}
	return item, err
}

const listItemsForUserQuery = `
SELECT * FROM items
WHERE user_id = $1
ORDER BY created_at DESC
`

func (ir *ItemRepository) ListItemsForUser(ctx context.Context, userId string) ([]models.Item, error) {
	items := []models.Item{}

	rows, err := ir.db.QueryContext(ctx, listItemsForUserQuery, userId)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.Id, &item.CreatedAt, &item.Content, &item.UserId); err != nil {
			return items, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

const deleteItemForUserQuery = "DELETE FROM items WHERE id = $1 AND user_id = $2;"

type DeleteUserItemParams struct {
	ItemId string
	UserId string
}

func (ir *ItemRepository) DeleteItemForUser(ctx context.Context, arg DeleteUserItemParams) error {
	_, err := ir.db.ExecContext(ctx, deleteItemForUserQuery, arg.ItemId, arg.UserId)
	return err
}
