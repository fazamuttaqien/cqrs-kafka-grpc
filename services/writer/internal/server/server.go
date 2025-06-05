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
	log             logger.Logger
	cfg             *config.Config
	v               *validator.Validate
	kafkaConnection *kafka.Conn
	ps              *service.ProductService
	im              interceptors.InterceptorManager
	pgxPool         *pgxpool.Pool
	metrics         *metrics.WriterServiceMetrics
}

func NewServer(log logger.Logger, cfg *config.Config) *server {
	return &server{log: log, cfg: cfg, v: validator.New()}
}

func (s *server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	s.im = interceptors.NewInterceptorManager(s.log)
	s.metrics = metrics.NewWriterServiceMetrics(s.cfg)

	pgxPool, err := postgres.NewPgx(s.cfg.Postgresql)
	if err != nil {
		return errors.Wrap(err, "postgresql.NewPgxConn")
	}
	s.pgxPool = pgxPool
	s.log.Infof("postgres connected: %v", pgxPool.Stat().TotalConns())
	defer pgxPool.Close()

	kafkaProducer := kafka_client.NewProducer(s.log, s.cfg.Kafka.Brokers)
	defer kafkaProducer.Close() // nolint: errcheck

	productRepository := repository.NewProductRepository(s.log, s.cfg, pgxPool)
	s.ps = service.NewProductService(s.log, s.cfg, productRepository, kafkaProducer)
	productMessageProcessor := kafka_consumer.NewProductMessageProcessor(s.log, s.cfg, s.v, s.ps, s.metrics)

	s.log.Info("Starting Writer Kafka consumers")
	cg := kafka_client.NewConsumerGroup(s.cfg.Kafka.Brokers, s.cfg.Kafka.GroupID, s.log)
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

	if s.cfg.Kafka.InitTopics {
		s.initKafkaTopics(ctx)
	}

	s.runHealthCheck(ctx)
	s.runMetrics(cancel)

	if s.cfg.Jaeger.Enable {
		tracer, err := tracing.NewJaegerTracerProvider(s.cfg.Jaeger)
		if err != nil {
			return err
		}
		defer tracer.Shutdown(ctx) // nolint: errcheck
	}

	<-ctx.Done()
	grpcServer.GracefulStop()

	return nil
}
