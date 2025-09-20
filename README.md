# Order Base

This project is a microservice for managing orders, consuming messages from Kafka and saving them in postgres, it also provides a REST API for order retrieval and simple web UI.

## Features

- **REST API**: Provides an endpoint to retrieve order information by ID.
- **Kafka Consumer**: Subscribes to a Kafka topic to process and save orders.
- **PostgreSQL Persistence**: All orders data is stored in a PostgreSQL database.
- **In-Memory Cache with TTL**: Stores recently accessed orders in memory (with TTL) for faster retrieval.

## Quick start 🚀

### Requirements 📦
-   [Docker](https://docs.docker.com/get-docker/)
-   [Go 1.24.3+](https://golang.org/doc/install) (only if you want to run order-base or kafka-client manually)

### 1. Clone Repository 📂
```bash
git clone https://github.com/Util787/order-base
cd order-base/order-base
```

### 2. Configure `.env` ⚙️
Create a `.env` file and configure according to your environment (`prod`, `dev`, or `local`) for example:
```env
ENV=local
SHUTDOWN_TIMEOUT=3s

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB_NAME=postgres
POSTGRES_USER=postgres
POSTGRES_PASSWORD=111
POSTGRES_MAX_CONNS=10
POSTGRES_CONN_MAX_LIFETIME=1h
POSTGRES_CONN_MAX_IDLE_TIME=30s

HTTP_SERVER_HOST=0.0.0.0
HTTP_SERVER_PORT=8080
HTTP_SERVER_READ_HEADER_TIMEOUT=5s
HTTP_SERVER_WRITE_TIMEOUT=10s
HTTP_SERVER_READ_TIMEOUT=10s

KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=orders
KAFKA_GROUP_ID=orders-consumer-group
KAFKA_MAX_WAIT=5s

BITNAMI_VERSION=3.6
POSTGRES_VERSION=17
ORDER_BASE_PORT=8080
ZOOKEEPER_PORT=2181
KAFKA_PORT=9095
```

### 3. Run with Docker Compose 🐳
Now you can run the entire project using Docker Compose (Postgres migrations will be applied automatically).

```bash
docker compose up --build
```

### 4. UI 💻
UI for receiving orders is available on:
```
http://localhost:ORDER_BASE_PORT/order-base
```
Replace `ORDER_BASE_PORT` with the actual port from your `.env`

## Testing with Kafka Client 🛠️

You can use provided `kafka-client(for_tests)` to send test orders to the Kafka topic

1.  In new terminal (you can split terminals in VSCode if you forgot) navigate to the `kafka-client(for_tests)` directory:
    ```bash
    cd "kafka-client(for_tests)"
    ```

2.  Run the client, specifying the Kafka broker port (default is 9092) and topic (default is 'orders'):
    ```bash
    go run client.go -p KAFKA_PORT -t KAFKA_TOPIC
    ```
    Replace `KAFKA_PORT` and `KAFKA_TOPIC` with the actual values from your `.env`

    The client will then prompt you to enter the number of orders and items per order to send. These orders will be published to the Kafka topic, and the `order-base` service will consume and process them

## Manual run

### Before running order-base locally without docker don't forget to run this command in `order-base` dir:

```bash
go mod tidy
``` 

### Order-base also works with yaml config (yaml won't work with docker-compose though):
1. Add path to your yaml in `CONFIG_PATH` env var:

```env
CONFIG_PATH=
```

2. Yaml example:

```yaml
env:
shutdown-timeout:

postgres:
  host:
  port: 
  db-name:
  user:
  password:

  max-conns:
  conn-max-lifetime:
  conn-max-idle-time:

http-server:
  host:
  port:
  read-header-timeout:
  write-timeout:
  read-timeout:

kafka:
  brokers:
  -
  topic:
  group-id:
  max-wait:
```

### TO DO
- Add pgx mapping
- Add sharding in InMemoryStorage
- Add DLQ for kafka
