package rest

import (
	"github.com/Util787/order-base/internal/config"
	"github.com/gin-gonic/gin"
)

func (h *Handler) InitRoutes(env string) *gin.Engine {
	if env == config.EnvProd {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	if env != config.EnvProd {
		router.Use(gin.Logger())
	}

	router.StaticFile("/order-base", "./ui/index.html")

	v1 := router.Group("/api/v1")
	v1.Use(NewBasicMiddleware(h.log))

	{
		orders := v1.Group("/orders")
		{
			orders.GET("/:order_id", h.getOrderById)
		}
	}
	return router
}
