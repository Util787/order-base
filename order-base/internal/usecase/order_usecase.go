package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Util787/order-base/internal/common"
	"github.com/Util787/order-base/internal/models"
)

func (u *OrderUsecase) GetOrderById(ctx context.Context, id string) (models.Order, error) {
	op := common.GetOperationName()
	log := common.LogOpAndId(ctx, op, u.log)

	order, err := u.cacheStorage.GetOrder(ctx, id)
	if err == nil {
		log.Debug("found in cache", slog.String("order_id", id))
		return order, nil
	}
	log.Debug("failed to found in cache, fetching from storage", slog.String("order_id", id), slog.String("error", err.Error()))

	order, err = u.orderStorage.GetOrderById(ctx, id)
	if err != nil {
		return models.Order{}, fmt.Errorf("%s: %w", op, err)
	}
	return order, nil
}

func (u *OrderUsecase) SaveOrder(ctx context.Context, order models.Order) error {
	op := common.GetOperationName()
	log := common.LogOpAndId(ctx, op, u.log)

	if err := u.orderStorage.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := u.cacheStorage.CacheOrder(ctx, order.OrderUID, order, &common.DefaultTTL); err != nil {
		log.Warn("failed to cache order", slog.String("order_id", order.OrderUID), slog.String("error", err.Error()))
	}

	return nil
}
