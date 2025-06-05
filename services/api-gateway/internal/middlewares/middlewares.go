package middlewares

import (
	"strings"
	"time"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/config"
	"github.com/gofiber/fiber/v2"
)

type MiddlewareManager interface {
	RequestLoggerMiddleware(ctx *fiber.Ctx) error
}

type middlewareManager struct {
	log logger.Logger
	cfg *config.Config
}

func NewMiddlewareManager(
	log logger.Logger,
	cfg *config.Config,
) *middlewareManager {
	return &middlewareManager{
		log: log,
		cfg: cfg,
	}
}

func (mw *middlewareManager) RequestLoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Call next handler
		err := c.Next()

		// Get response status and size
		status := c.Response().StatusCode()
		size := len(c.Response().Body())
		s := time.Since(start)

		if !mw.checkIgnoredURI(c.Path(), mw.cfg.Http.IgnoreLogUrls) {
			mw.log.HttpMiddlewareAccessLogger(
				c.Method(),
				c.Path(),
				status,
				int64(size),
				s,
			)
		}

		return err
	}
}

func (mw *middlewareManager) checkIgnoredURI(requestURI string, uriList []string) bool {
	for _, s := range uriList {
		if strings.Contains(requestURI, s) {
			return true
		}
	}
	return false
}
