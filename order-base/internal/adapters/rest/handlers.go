package rest

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Util787/order-base/internal/models"
	"github.com/gin-gonic/gin"
)

type OrderUsecase interface {
	GetOrderById(ctx context.Context, id string) (models.Order, error)
}

type Handler struct {
	log          *slog.Logger
	orderUsecase OrderUsecase
}

func (h *Handler) getOrderById(c *gin.Context) {
	log := h.log.With(slog.String("op", c.GetString("op")), slog.String("request_id", c.GetString("request_id")))

	orderUID := c.Param("order_id")
	log.Debug("Recieved order_id", slog.String("order_id", orderUID))

	order, err := h.orderUsecase.GetOrderById(c.Request.Context(), orderUID)
	if err != nil {
		if errors.Is(err, models.ErrOrdersNotFound) {
			newErrorResponse(c, log, http.StatusNotFound, "order not found", err)
			return
		}
		newErrorResponse(c, log, http.StatusInternalServerError, "failed to get order", err)
		return
	}

	c.JSON(http.StatusOK, order)
}
