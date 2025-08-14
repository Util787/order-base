package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Util787/order-base/internal/adapters/rest"
	"github.com/Util787/order-base/internal/config"
	"github.com/Util787/order-base/internal/infra/storage"
	"github.com/Util787/order-base/internal/logger/slogpretty"
	"github.com/Util787/order-base/internal/usecase"
)

func main() {
	cfg := config.MustLoadConfig()

	// for now I think using logger only in adapters and usecase layers will be enough
	log := setupLogger(cfg.Env)

	// storages
	postgreStorage := storage.MustInitPostgres(context.Background(), cfg.PostgresConfig)

	inMemoryStorage := storage.NewInMemoryStorage(context.Background(), 100)
	var loadLimit uint64 = 1
	inMemoryStorage.LoadOrders(context.Background(), &postgreStorage, &loadLimit)

	// usecases
	orderUsecase := usecase.NewOrderUsecase(log, &postgreStorage, &inMemoryStorage)

	//rest
	serv := rest.NewHTTPServer(cfg.Env, cfg.HTTPServerConfig, log, &orderUsecase)
	log.Info("Starting HTTP server", slog.String("host", cfg.HTTPServerConfig.Host), slog.Int("port", cfg.HTTPServerConfig.Port))
	serv.Run()
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
