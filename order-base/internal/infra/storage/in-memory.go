package storage

import (
	"context"
	"fmt"

	"github.com/Util787/order-base/internal/common"
	"github.com/Util787/order-base/internal/models"
)

type InMemoryStorage struct {
	orders map[string]models.Order
}

func NewInMemoryStorage(ctx context.Context, startSize int) InMemoryStorage {
	return InMemoryStorage{
		orders: make(map[string]models.Order, startSize),
	}
}

type OrderStorage interface {
	GetAllOrders(ctx context.Context, limit *uint64) ([]models.Order, error)
}

// LoadOrders loads orders from the OrderStorage into InMemoryStorage using orderUID as key.
//
// loadLimit specifies the maximum number of orders to load.
func (s *InMemoryStorage) LoadOrders(ctx context.Context, orderStorage OrderStorage, loadLimit *uint64) error {
	op := common.GetOperationName()

	orders, err := orderStorage.GetAllOrders(ctx, loadLimit)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, order := range orders {
		s.CacheOrder(ctx, order.OrderUID, order)
	}
	return nil
}

func (s *InMemoryStorage) CacheOrder(ctx context.Context, key string, order models.Order) error {
	op := common.GetOperationName()

	if ctx.Err() != nil {
		return fmt.Errorf("%s: %w", op, ctx.Err())
	}

	s.orders[key] = order
	return nil
}

func (s *InMemoryStorage) GetOrder(ctx context.Context, key string) (models.Order, error) {
	op := common.GetOperationName()

	if ctx.Err() != nil {
		return models.Order{}, fmt.Errorf("%s: %w", op, ctx.Err())
	}

	order, exists := s.orders[key]
	if !exists {
		return models.Order{}, fmt.Errorf("%s: %w", op, models.ErrOrdersNotFound)
	}
	return order, nil
}
