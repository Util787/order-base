package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Util787/order-base/internal/common"
	"github.com/Util787/order-base/internal/models"
)

type InMemoryStorage struct {
	orders map[string]orderCache
	mu     sync.RWMutex
}

// NewInMemoryStorage must return pointer because of RWMutex in it.
//
// startSize defines the initial capacity of the order cache map.
//
// cleanUpInterval defines the interval for cleaning up expired orders.
func NewInMemoryStorage(ctx context.Context, startSize int, cleanUpInterval time.Duration) *InMemoryStorage {
	strg := &InMemoryStorage{
		orders: make(map[string]orderCache, startSize),
		mu:     sync.RWMutex{},
	}

	ticker := time.NewTicker(cleanUpInterval)
	go func() {
		for range ticker.C {
			strg.cleanUpExpiredOrders()
		}
	}()

	return strg
}

type OrderStorage interface {
	GetAllOrders(ctx context.Context, limit *uint64) ([]models.Order, error)
}

// LoadOrders loads orders from the OrderStorage into InMemoryStorage using orderUID as key.
//
// loadLimit specifies the maximum number of orders to load.
func (i *InMemoryStorage) LoadOrders(ctx context.Context, orderStorage OrderStorage, loadLimit *uint64, ttl *time.Duration) error {
	op := common.GetOperationName()

	orders, err := orderStorage.GetAllOrders(ctx, loadLimit)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, order := range orders {
		i.CacheOrder(ctx, order.OrderUID, order, ttl)
	}
	return nil
}

type orderCache struct {
	order      models.Order
	expiration *uint32 // Unix timestamp, if nil then cache has no expiration time
}

// If ttl is nil then no ttl will be set
func (i *InMemoryStorage) CacheOrder(ctx context.Context, key string, order models.Order, ttl *time.Duration) error {
	op := common.GetOperationName()

	if ctx.Err() != nil {
		return fmt.Errorf("%s: %w", op, ctx.Err())
	}

	var cache orderCache
	cache.order = order

	if ttl != nil {
		expiration := uint32(time.Now().Add(*ttl).Unix())
		cache.expiration = &expiration
	} else {
		cache.expiration = nil
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	i.orders[key] = cache
	return nil
}

func (i *InMemoryStorage) GetOrder(ctx context.Context, key string) (models.Order, error) {
	op := common.GetOperationName()

	if ctx.Err() != nil {
		return models.Order{}, fmt.Errorf("%s: %w", op, ctx.Err())
	}

	i.mu.RLock()
	defer i.mu.RUnlock()

	orderCache, exists := i.orders[key]
	if !exists {
		return models.Order{}, fmt.Errorf("%s: %w", op, models.ErrOrdersNotFound)
	}
	return orderCache.order, nil
}

func (i *InMemoryStorage) cleanUpExpiredOrders() {
	currentTime := uint32(time.Now().Unix())

	i.mu.Lock()
	defer i.mu.Unlock()

	for key, cache := range i.orders {
		if cache.expiration != nil && *cache.expiration < currentTime {
			delete(i.orders, key)
		}
	}
}
