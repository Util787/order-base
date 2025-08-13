CREATE TABLE IF NOT EXISTS deliveries (
    delivery_uid VARCHAR(50) PRIMARY KEY, -- Im not sure if I could use name as pk so I created this field
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(16) NOT NULL,
    zip VARCHAR(20) NOT NULL,
    city VARCHAR(100) NOT NULL,
    address VARCHAR(255) NOT NULL,
    region VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
    transaction VARCHAR(50) PRIMARY KEY,
    request_id VARCHAR(100) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    provider VARCHAR(100) NOT NULL,
    amount INTEGER NOT NULL,
    payment_dt INTEGER NOT NULL,
    bank VARCHAR(100) NOT NULL,
    delivery_cost INTEGER NOT NULL,
    goods_total INTEGER NOT NULL,
    custom_fee INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(50) PRIMARY KEY,
    track_number VARCHAR(100) NOT NULL,
    entry VARCHAR(100) NOT NULL,
    delivery_uid VARCHAR(50) UNIQUE REFERENCES deliveries(delivery_uid),
    payment_transaction VARCHAR(50) UNIQUE REFERENCES payments(transaction),
    locale VARCHAR(10) NOT NULL,
    internal_signature VARCHAR(100) NOT NULL,
    customer_id VARCHAR(100) NOT NULL,
    delivery_service VARCHAR(100) NOT NULL,
    shardkey VARCHAR(100) NOT NULL,
    sm_id INTEGER NOT NULL,
    date_created TIMESTAMPTZ NOT NULL,
    oof_shard VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS items (
    chrt_id BIGINT PRIMARY KEY,
    track_number VARCHAR(100) NOT NULL,
    price INTEGER NOT NULL,
    rid VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sale INTEGER NOT NULL,
    size VARCHAR(50) NOT NULL,
    total_price INTEGER NOT NULL,
    nm_id INTEGER NOT NULL,
    brand VARCHAR(100) NOT NULL,
    status INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS order_items (
    order_uid VARCHAR(50) REFERENCES orders(order_uid),
    chrt_id BIGINT REFERENCES items(chrt_id),
    PRIMARY KEY (order_uid, chrt_id)
);
