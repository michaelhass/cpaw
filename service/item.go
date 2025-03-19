package service

import (
	"context"

	"github.com/michaelhass/cpaw/db/repository"
	"github.com/michaelhass/cpaw/models"
)

type ItemService struct {
	items *repository.ItemRepository
}

func NewItemService(items *repository.ItemRepository) *ItemService {
	return &ItemService{items: items}
}

type CreateItemsParams = repository.CreateItemParams

func (is *ItemService) CreateItem(ctx context.Context, params CreateItemsParams) (models.Item, error) {
	return is.items.CreateItem(ctx, params)
}

func (is *ItemService) ListItemsForUser(ctx context.Context, userId string) ([]models.Item, error) {
	return is.items.ListItemsForUser(ctx, userId)
}

func (is *ItemService) DeleteItemById(ctx context.Context, itemId string) error {
	return is.items.DeleteItemById(ctx, itemId)
}
