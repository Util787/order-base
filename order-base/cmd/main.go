package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	kafka_subscriber "github.com/Util787/order-base/internal/adapters/kafka-subscriber"
	"github.com/Util787/order-base/internal/adapters/rest"
	"github.com/Util787/order-base/internal/config"
	"github.com/Util787/order-base/internal/infra/storage"
	"github.com/Util787/order-base/internal/logger/slogpretty"
	"github.com/Util787/order-base/internal/usecase"
)

// kafka vars
const (
	numFetchers    = 2
	numHandlers    = 2
	messageChanBuf = 100
)

// in-memory storage vars
var (
	loadLimit uint64 = 100
)

func main() {
	cfg := config.MustLoadConfig()

	// for now I think using logger only in adapters and usecase layers will be enough
	log := setupLogger(cfg.Env)

	// storages
	postgreStorage := storage.MustInitPostgres(context.Background(), cfg.PostgresConfig)

	inMemoryStorage := storage.NewInMemoryStorage(context.Background(), 100)
	inMemoryStorage.LoadOrders(context.Background(), &postgreStorage, &loadLimit)

	// usecases
	orderUsecase := usecase.NewOrderUsecase(log, &postgreStorage, &inMemoryStorage)

	// kafka
	kafkaSub := kafka_subscriber.NewKafkaSubscriber(log, cfg.KafkaConfig, &orderUsecase, messageChanBuf)

	// rest
	serv := rest.NewHTTPServer(log, cfg.Env, cfg.HTTPServerConfig, &orderUsecase)

	// start
	go kafkaSub.Subscribe(context.Background(), numFetchers, numHandlers)

	go func() {
		log.Info("HTTP server start", slog.String("host", cfg.HTTPServerConfig.Host), slog.Int("port", cfg.HTTPServerConfig.Port))
		if err := serv.Run(); err != nil {
			log.Error("HTTP server error", slog.String("error", err.Error()))
		}
	}()

	//graceful shutdown
	shutDownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	log.Info("Shutting down gracefully...")

	log.Info("Shutting down server")
	if err := serv.Shutdown(shutDownCtx); err != nil {
		log.Error("HTTP server shutdown error", slog.String("error", err.Error()))
	}

	log.Info("Shutting down kafka subscriber")
	if err := kafkaSub.Shutdown(); err != nil {
		log.Error("Kafka subscriber shutdown error", slog.String("error", err.Error()))
	}

	log.Info("Shutting down postgres")
	postgreStorage.Shutdown()

	log.Info("Shutdown complete")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case config.EnvLocal:
		log = slogpretty.NewPrettyLogger(os.Stdout, slog.LevelDebug)
	case config.EnvDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case config.EnvProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
