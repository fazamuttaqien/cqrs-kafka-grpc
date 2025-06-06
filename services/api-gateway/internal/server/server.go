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
	log                logger.Logger
	config             *config.Config
	validate           *validator.Validate
	middlewareManager  middlewares.MiddlewareManager
	interceptorManager interceptors.InterceptorManager
	fiber              *fiber.App
	productService     *service.ProductService
	metrics            *metrics.ApiGatewayMetrics
}

func NewServer(log logger.Logger, cfg *config.Config) *server {
	return &server{log: log, config: cfg, fiber: fiber.New(), validate: validator.New()}
}

func (s *server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	tracer, err := tracing.NewOTLPTracerProvider(s.config.OTLP)
	if err != nil {
		return err
	}
	defer tracer.Shutdown(ctx) // nolint: errcheck

	s.middlewareManager = middlewares.NewMiddlewareManager(s.log, s.config)
	s.interceptorManager = interceptors.NewInterceptorManager(s.log)
	s.metrics = metrics.NewApiGatewayMetrics(s.config)

	readerServiceConnection, err := client.NewReaderServiceConn(ctx, s.config, s.interceptorManager)
	if err != nil {
		return err
	}
	defer readerServiceConnection.Close() // nolint: errcheck
	readerServiceClient := pb_reader.NewReaderServiceClient(readerServiceConnection)

	kafkaProducer := kafka.NewProducer(s.log, s.config.Kafka.Brokers)
	defer kafkaProducer.Close() // nolint: errcheck

	s.productService = service.NewProductService(
		s.log, s.config, kafkaProducer, readerServiceClient, tracer.GetTracer())

	productHandlers := v1.NewProductsHandlers(
		s.fiber.Group(s.config.Http.ProductsPath).(*fiber.Group),
		s.log, s.middlewareManager, s.config, s.productService, s.validate, s.metrics, tracer.GetTracer())
	productHandlers.MapRoutes()

	go func() {
		if err := s.runHttpServer(); err != nil {
			s.log.Errorf(" s.runHttpServer: %v", err)
			cancel()
		}
	}()
	s.log.Infof("API Gateway is listening on PORT: %s", s.config.Http.Port)

	s.runMetrics(cancel)
	s.runHealthCheck(ctx)

	<-ctx.Done()
	if err := s.fiber.Server().Shutdown(); err != nil {
		s.log.WarnMsg("fiber.Server.Shutdown", err)
	}

	return nil
}
