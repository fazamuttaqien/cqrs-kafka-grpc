package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/interceptors"
	kafka_client "github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/postgres"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/metrics"
	kafka_consumer "github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/delivery/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/repository"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/service"

	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type server struct {
	log                logger.Logger
	config             *config.Config
	validate           *validator.Validate
	kafkaConnection    *kafka.Conn
	productService     *service.ProductService
	interceptorManager interceptors.InterceptorManager
	pgxPool            *pgxpool.Pool
	metrics            *metrics.WriterServiceMetrics
}

func NewServer(log logger.Logger, cfg *config.Config) *server {
	return &server{log: log, config: cfg, validate: validator.New()}
}

func (s *server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	tracer, err := tracing.NewOTLPTracerProvider(s.config.OTLP)
	if err != nil {
		return err
	}
	defer tracer.Shutdown(ctx) // nolint: errcheck

	s.interceptorManager = interceptors.NewInterceptorManager(s.log)
	s.metrics = metrics.NewWriterServiceMetrics(s.config)

	pgxPool, err := postgres.NewPgxConnection(s.config.Postgres)
	if err != nil {
		return errors.Wrap(err, "postgresql.NewPgxConnection")
	}
	s.pgxPool = pgxPool
	s.log.Infof("postgres connected: %v", pgxPool.Stat().TotalConns())
	defer pgxPool.Close()

	kafkaProducer := kafka_client.NewProducer(s.log, s.config.Kafka.Brokers)
	defer kafkaProducer.Close() // nolint: errcheck

	productRepository := repository.NewProductRepository(s.log, s.config, pgxPool, tracer.GetTracer())
	s.productService = service.NewProductService(s.log, s.config, productRepository, kafkaProducer, tracer.GetTracer())
	productMessageProcessor := kafka_consumer.NewProductMessageProcessor(s.log, s.config, s.validate, s.productService, s.metrics)

	s.log.Info("Starting Writer Kafka consumers")
	cg := kafka_client.NewConsumerGroup(s.config.Kafka.Brokers, s.config.Kafka.GroupID, s.log)
	go cg.ConsumeTopic(ctx, s.getConsumerGroupTopics(), kafka_consumer.PoolSize, productMessageProcessor.ProcessMessages)

	closeGrpcServer, grpcServer, err := s.newWriterGrpcServer()
	if err != nil {
		return errors.Wrap(err, "NewScmGrpcServer")
	}
	defer closeGrpcServer() // nolint: errcheck

	if err := s.connectKafkaBrokers(ctx); err != nil {
		return errors.Wrap(err, "s.connectKafkaBrokers")
	}
	defer s.kafkaConnection.Close() // nolint: errcheck

	if s.config.Kafka.InitTopics {
		s.initKafkaTopics(ctx)
	}

	s.runHealthCheck(ctx)
	s.runMetrics(cancel)

	<-ctx.Done()
	grpcServer.GracefulStop()

	return nil
}
