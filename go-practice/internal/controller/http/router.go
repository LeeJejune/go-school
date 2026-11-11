// Package http exposes the public HTTP API through Gin.
package http

import (
	"log/slog"
	"net/http"

	"github.com/LeeJejune/go-school/go-practice/internal/config"
	"github.com/LeeJejune/go-school/go-practice/internal/gateway"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	clients   *gateway.Clients
	dashboard *gateway.Dashboard
}

func NewHandler(clients *gateway.Clients, dashboard *gateway.Dashboard) *Handler {
	return &Handler{clients: clients, dashboard: dashboard}
}

func NewRouter(cfg config.Config, logger *slog.Logger, h *Handler) *gin.Engine {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.HandleMethodNotAllowed = true
	router.Use(requestScope(cfg.RequestTimeout, logger), recoveryMiddleware(logger))
	router.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok\n") })
	v1 := router.Group("/v1")
	{
		v1.GET("/resources", h.ListResources)
		v1.PUT("/resources/:id", h.UpsertResource)
		v1.POST("/lists", h.CreateList)
		v1.GET("/lists/:id", h.GetList)
		v1.PATCH("/lists/:id/items/:resourceID", h.SetRead)
		v1.GET("/dashboard/:id", h.GetDashboard)
	}
	router.NoRoute(func(c *gin.Context) { writeError(c, status.Error(codes.NotFound, "route not found")) })
	router.NoMethod(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, errorResponse{Code: int32(codes.Unimplemented), Message: "method not allowed", Details: []any{}})
	})
	return router
}
