package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/Util787/order-base/internal/common"
	"github.com/Util787/order-base/internal/config"
	"github.com/Util787/order-base/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Use to set orders slice capacity
const ordersDefaultCap uint8 = 100

type PostgresStorage struct {
	pgxPool *pgxpool.Pool
}

func MustInitPostgres(ctx context.Context, cfg config.PostgresConfig, log *slog.Logger) PostgresStorage {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DbName,
	)

	pgxConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		panic(fmt.Errorf("failed to parse postgres connection string: %w", err))
	}

	// Pool configuration
	pgxConfig.MaxConns = int32(cfg.MaxConns)
	pgxConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	pgxConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create postgres connection pool: %w", err))
	}

	err = pool.Ping(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to ping postgres: %w", err))
	}

	return PostgresStorage{
		pgxPool: pool,
	}
}

func (p *PostgresStorage) Shutdown() {
	p.pgxPool.Close()
}

var orderQueryBase sq.SelectBuilder = sq.StatementBuilder.
	PlaceholderFormat(sq.Dollar).
	Select(
		"orders.order_uid",
		"orders.track_number",
		"orders.entry",
		"orders.locale",
		"orders.internal_signature",
		"orders.customer_id",
		"orders.delivery_service",
		"orders.shardkey",
		"orders.sm_id",
		"orders.date_created",
		"orders.oof_shard",

		"deliveries.delivery_uid",
		"deliveries.name",
		"deliveries.phone",
		"deliveries.zip",
		"deliveries.city",
		"deliveries.address",
		"deliveries.region",
		"deliveries.email",

		"payments.transaction",
		"payments.request_id",
		"payments.currency",
		"payments.provider",
		"payments.amount",
		"payments.payment_dt",
		"payments.bank",
		"payments.delivery_cost",
		"payments.goods_total",
		"payments.custom_fee",
		// items aggregated in json(its better than array_agg)
		"COALESCE(json_agg(items) FILTER (WHERE items.chrt_id IS NOT NULL), '[]') AS items",
	).
	From("orders").
	Join("deliveries ON orders.delivery_uid = deliveries.delivery_uid").
	Join("payments ON orders.payment_transaction = payments.transaction").
	// I dont think left joins are necessary here because orders with no items are pointless? But it might be useful to specify error. Can change to inner join to filter orders with no items
	LeftJoin("order_items ON orders.order_uid = order_items.order_uid").
	LeftJoin("items ON order_items.chrt_id = items.chrt_id").
	GroupBy(
		"orders.order_uid",
		"deliveries.delivery_uid",
		"payments.transaction",
	)

// This method should be called to cache the most recently-created orders but it may be useful somewhere else in the future
//
// if limit is nil then no limit is applied
func (p *PostgresStorage) GetAllOrders(ctx context.Context, limit *uint64) ([]models.Order, error) {
	op := common.GetOperationName()

	var orders []models.Order

	queryBuilder := orderQueryBase.
		OrderBy("orders.date_created DESC")

	if limit != nil {
		queryBuilder = queryBuilder.Limit(*limit)
		orders = make([]models.Order, 0, *limit)
	} else {
		orders = make([]models.Order, 0, ordersDefaultCap)
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	conn, err := p.pgxPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to acquire connection: %w", op, err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var ord models.Order
		var itemsJSON []byte

		err := rows.Scan(
			&ord.OrderUID,
			&ord.TrackNumber,
			&ord.Entry,
			&ord.Locale,
			&ord.InternalSignature,
			&ord.CustomerID,
			&ord.DeliveryService,
			&ord.Shardkey,
			&ord.SmID,
			&ord.DateCreated,
			&ord.OofShard,

			&ord.Delivery.DeliveryUID,
			&ord.Delivery.Name,
			&ord.Delivery.Phone,
			&ord.Delivery.Zip,
			&ord.Delivery.City,
			&ord.Delivery.Address,
			&ord.Delivery.Region,
			&ord.Delivery.Email,

			&ord.Payment.Transaction,
			&ord.Payment.RequestID,
			&ord.Payment.Currency,
			&ord.Payment.Provider,
			&ord.Payment.Amount,
			&ord.Payment.PaymentDt,
			&ord.Payment.Bank,
			&ord.Payment.DeliveryCost,
			&ord.Payment.GoodsTotal,
			&ord.Payment.CustomFee,

			&itemsJSON,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) { // just in case someone will add filter logic
				return nil, fmt.Errorf("%s: %w", op, models.ErrOrdersNotFound)
			}
			return nil, fmt.Errorf("%s: failed to scan rows: %w", op, err)
		}

		if err := json.Unmarshal(itemsJSON, &ord.Items); err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal items JSON: %w", op, err)
		}

		orders = append(orders, ord)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: rows err: %w", op, rows.Err())
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("%s: %w", op, models.ErrOrdersNotFound)
	}

	return orders, nil
}

func (p *PostgresStorage) GetOrderById(ctx context.Context, id string) (models.Order, error) {
	op := common.GetOperationName()

	queryBuilder := orderQueryBase.
		Where(sq.Eq{"orders.order_uid": id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return models.Order{}, fmt.Errorf("%s: failed to build query: %w", op, err)
	}

	conn, err := p.pgxPool.Acquire(ctx)
	if err != nil {
		return models.Order{}, fmt.Errorf("%s: failed to acquire connection: %w", op, err)
	}
	defer conn.Release()

	row := conn.QueryRow(ctx, query, args...)

	var ord models.Order
	var itemsJSON []byte

	err = row.Scan(
		&ord.OrderUID,
		&ord.TrackNumber,
		&ord.Entry,
		&ord.Locale,
		&ord.InternalSignature,
		&ord.CustomerID,
		&ord.DeliveryService,
		&ord.Shardkey,
		&ord.SmID,
		&ord.DateCreated,
		&ord.OofShard,

		&ord.Delivery.DeliveryUID,
		&ord.Delivery.Name,
		&ord.Delivery.Phone,
		&ord.Delivery.Zip,
		&ord.Delivery.City,
		&ord.Delivery.Address,
		&ord.Delivery.Region,
		&ord.Delivery.Email,

		&ord.Payment.Transaction,
		&ord.Payment.RequestID,
		&ord.Payment.Currency,
		&ord.Payment.Provider,
		&ord.Payment.Amount,
		&ord.Payment.PaymentDt,
		&ord.Payment.Bank,
		&ord.Payment.DeliveryCost,
		&ord.Payment.GoodsTotal,
		&ord.Payment.CustomFee,

		&itemsJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, fmt.Errorf("%s: %w", op, models.ErrOrdersNotFound)
		}
		return models.Order{}, fmt.Errorf("%s: failed to scan row: %w", op, err)
	}

	if err := json.Unmarshal(itemsJSON, &ord.Items); err != nil {
		return models.Order{}, fmt.Errorf("%s: failed to unmarshal items JSON: %w", op, err)
	}

	return ord, nil
}

func (p *PostgresStorage) SaveOrder(ctx context.Context, order models.Order) error {
	op := common.GetOperationName()

	conn, err := p.pgxPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to acquire connection: %w", op, err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
	INSERT INTO deliveries (delivery_uid, name, phone, zip, city, address, region, email)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, order.Delivery.DeliveryUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("%s: failed to insert delivery: %w", op, err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO payments (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to insert payment: %w", op, err)
	}

	_, err = tx.Exec(ctx, `
	INSERT INTO orders (order_uid, track_number, entry, delivery_uid, payment_transaction, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Delivery.DeliveryUID,
		order.Payment.Transaction,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard)
	if err != nil {
		return fmt.Errorf("%s: failed to insert order: %w", op, err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
		INSERT INTO items (
			chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return fmt.Errorf("%s: failed to insert item: %w", op, err)
		}
		_, err = tx.Exec(ctx, `
		INSERT INTO order_items (order_uid, chrt_id)
		VALUES ($1, $2)
		`,
			order.OrderUID,
			item.ChrtID,
		)
		if err != nil {
			return fmt.Errorf("%s: failed to insert order_item: %w", op, err)
		}
	}

	return tx.Commit(ctx)
}
