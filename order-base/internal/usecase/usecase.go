package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/Util787/order-base/internal/models"
)

type OrderStorage interface {
	GetOrderById(ctx context.Context, id string) (models.Order, error)
	SaveOrder(ctx context.Context, order models.Order) error
}

type CacheStorage interface {
	GetOrder(ctx context.Context, key string) (models.Order, error)
	CacheOrder(ctx context.Context, key string, order models.Order, ttl *time.Duration) error
}

type OrderUsecase struct {
	log          *slog.Logger
	orderStorage OrderStorage
	cacheStorage CacheStorage
}

func NewOrderUsecase(log *slog.Logger, orderStorage OrderStorage, cacheStorage CacheStorage) OrderUsecase {
	return OrderUsecase{
		log:          log,
		orderStorage: orderStorage,
		cacheStorage: cacheStorage,
	}
}
