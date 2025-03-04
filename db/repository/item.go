package repository

import (
	"context"
	"database/sql"
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
RETURNING (id, created_at, content, user_id);
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

const listItemsForUserQuery = `
SELECT * FROM items
WHERE user_id = $1
ORDER BY created_at DESC
`

func (ir *ItemRepository) ListItemsForUser(ctx context.Context, userId string) ([]models.Item, error) {
	rows, err := ir.db.QueryContext(ctx, listItemsForUserQuery, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.Id, &item.CreatedAt, &item.Content, &item.UserId); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const deleteItemByIdQuery = "DELETE FROM items WHERE id = $1;"

func (ir *ItemRepository) DeleteItemById(ctx context.Context, itemID string) error {
	_, err := ir.db.ExecContext(ctx, deleteItemByIdQuery, itemID)
	return err
}
