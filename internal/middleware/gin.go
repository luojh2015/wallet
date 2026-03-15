package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luojh/wallet/pkg/logger"
	"go.uber.org/zap"
)

type GinMiddleware struct {
	logger logger.Logger
}

func NewGinMiddleware(logger logger.Logger) *GinMiddleware {
	return &GinMiddleware{logger: logger}
}

func (l *GinMiddleware) Apply(r gin.IRoutes) {
	r.Use(gin.Recovery())

	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		l.logger.Info("HTTP Request",
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	})
}
