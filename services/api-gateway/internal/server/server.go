package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/interceptors"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/client"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/metrics"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/middlewares"
	v1 "github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/products/delivery/http/v1"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/products/service"
	pb_reader "github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/proto/product_reader"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type server struct {
	log   logger.Logger
	cfg   *config.Config
	v     *validator.Validate
	mw    middlewares.MiddlewareManager
	im    interceptors.InterceptorManager
	fiber *fiber.App
	ps    *service.ProductService
	m     *metrics.ApiGatewayMetrics
}

func NewServer(log logger.Logger, cfg *config.Config) *server {
	return &server{log: log, cfg: cfg, fiber: fiber.New(), v: validator.New()}
}

func (s *server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	s.mw = middlewares.NewMiddlewareManager(s.log, s.cfg)
	s.im = interceptors.NewInterceptorManager(s.log)
	s.m = metrics.NewApiGatewayMetrics(s.cfg)

	readerServiceConn, err := client.NewReaderServiceConn(ctx, s.cfg, s.im)
	if err != nil {
		return err
	}
	defer readerServiceConn.Close() // nolint: errcheck
	rsClient := pb_reader.NewReaderServiceClient(readerServiceConn)

	kafkaProducer := kafka.NewProducer(s.log, s.cfg.Kafka.Brokers)
	defer kafkaProducer.Close() // nolint: errcheck

	s.ps = service.NewProductService(s.log, s.cfg, kafkaProducer, rsClient)

	productHandlers := v1.NewProductsHandlers(
		s.fiber.Group(s.cfg.Http.ProductsPath), s.log, s.mw, s.cfg, s.ps, s.v, s.m)
	productHandlers.MapRoutes()

	go func() {
		if err := s.runHttpServer(); err != nil {
			s.log.Errorf(" s.runHttpServer: %v", err)
			cancel()
		}
	}()
	s.log.Infof("API Gateway is listening on PORT: %s", s.cfg.Http.Port)

	s.runMetrics(cancel)
	s.runHealthCheck(ctx)

	if s.cfg.Jaeger.Enable {
		tracer, err := tracing.NewJaegerTracerProvider(s.cfg.Jaeger)
		if err != nil {
			return err
		}
		defer tracer.Shutdown(ctx) // nolint: errcheck
	}

	<-ctx.Done()
	if err := s.fiber.Server().Shutdown(); err != nil {
		s.log.WarnMsg("echo.Server.Shutdown", err)
	}

	return nil
}
