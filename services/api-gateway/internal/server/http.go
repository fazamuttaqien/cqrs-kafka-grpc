package server

import (
	"strings"
	"time"

	// "github.com/fazamuttaqien/cqrs-kafka-grpc/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
)

const (
	maxHeaderBytes = 1 << 20
	stackSize      = 1 << 10         // 1 KB
	bodyLimit      = 2 * 1024 * 1024 // 2MB
	readTimeout    = 15 * time.Second
	writeTimeout   = 15 * time.Second
	gzipLevel      = 5
)

func (s *server) runHttpServer() error {
	// Create new Fiber instance
	s.fiber = fiber.New(fiber.Config{
		ReadTimeout:           readTimeout,
		WriteTimeout:          writeTimeout,
		BodyLimit:             bodyLimit,
		DisableStartupMessage: true,
	})

	s.mapRoutes()

	return s.fiber.Listen(s.config.Http.Port)
}

func (s *server) mapRoutes() {
	// Setup Swagger
	// docs.SwaggerInfo.Version = "1.0"
	// docs.SwaggerInfo.Title = "API Gateway"
	// docs.SwaggerInfo.Description = "API Gateway CQRS microservices."
	// docs.SwaggerInfo.BasePath = "/api/v1"

	// Swagger route
	s.fiber.Get("/swagger/*", swagger.HandlerDefault)

	// Middlewares
	s.fiber.Use(s.middlewareManager.RequestLoggerMiddleware)
	s.fiber.Use(recover.New(recover.Config{
		EnableStackTrace: false,
	}))
	s.fiber.Use(requestid.New())
	s.fiber.Use(compress.New(compress.Config{
		Level: compress.Level(gzipLevel),
		Next: func(c *fiber.Ctx) bool {
			return strings.Contains(c.Path(), "swagger")
		},
	}))
}
