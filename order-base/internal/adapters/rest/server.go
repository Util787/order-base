package rest

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Util787/order-base/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func NewHTTPServer(env string, config config.HTTPServerConfig, log *slog.Logger, orderUsecase OrderUsecase) Server {
	handler := Handler{
		log:          log,
		orderUsecase: orderUsecase,
	}

	httpServer := &http.Server{
		Addr:              config.Host + ":" + strconv.Itoa(config.Port),
		Handler:           handler.InitRoutes(env),
		MaxHeaderBytes:    1 << 20, // 1 MB
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		ReadTimeout:       config.ReadTimeout,
	}

	return Server{
		httpServer: httpServer,
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
