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

func (is *ItemService) GetItemById(ctx context.Context, itemId string) (models.Item, error) {
	return is.items.GetItemById(ctx, itemId)
}

type GetItemForUserParams = repository.GetItemForUserParams

func (is *ItemService) GetItemForUser(ctx context.Context, params GetItemForUserParams) (models.Item, error) {
	return is.items.GetItemForUser(ctx, params)
}

func (is *ItemService) ListItemsForUser(ctx context.Context, userId string) ([]models.Item, error) {
	return is.items.ListItemsForUser(ctx, userId)
}

type DeleteUserItemParams = repository.DeleteUserItemParams

func (is *ItemService) DeleteItemForUser(ctx context.Context, params DeleteUserItemParams) error {
	return is.items.DeleteItemForUser(ctx, params)
}
