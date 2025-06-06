package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/interceptors"
	kafka_client "github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/mongodb"
	redis_client "github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/redis"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/metrics"
	reader_kafka "github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/delivery/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/repository"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/service"

	"github.com/go-playground/validator"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type server struct {
	log                logger.Logger
	config             *config.Config
	validate           *validator.Validate
	kafkaConnection    *kafka.Conn
	interceptorManager interceptors.InterceptorManager
	mongoClient        *mongo.Client
	redisClient        redis.UniversalClient
	productService     *service.ProductService
	metrics            *metrics.ReaderServiceMetrics
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
	s.metrics = metrics.NewReaderServiceMetrics(s.config)

	mongoConnection, err := mongodb.NewMongoConnection(ctx, s.config.Mongo)
	if err != nil {
		return errors.Wrap(err, "NewMongoDBConn")
	}
	s.mongoClient = mongoConnection
	defer mongoConnection.Disconnect(ctx) // nolint: errcheck
	s.log.Infof("Mongo connected: %v", mongoConnection.NumberSessionsInProgress())

	s.redisClient = redis_client.NewUniversalRedisClient(s.config.Redis)
	defer s.redisClient.Close() // nolint: errcheck
	s.log.Infof("Redis connected: %+v", s.redisClient.PoolStats())

	mongoRepository := repository.NewMongoRepository(s.log, s.config, s.mongoClient, tracer.GetTracer())
	redisCache := repository.NewRedisRepository(s.log, s.config, s.redisClient, tracer.GetTracer())

	s.productService = service.NewProductService(
		s.log, s.config, mongoRepository, redisCache, tracer.GetTracer())

	readerMessageProcessor := reader_kafka.NewReaderMessageProcessor(
		s.log, s.config, s.validate, s.productService, s.metrics)

	s.log.Info("Starting Reader Kafka consumers")
	cg := kafka_client.NewConsumerGroup(s.config.Kafka.Brokers, s.config.Kafka.GroupID, s.log)
	go cg.ConsumeTopic(ctx, s.getConsumerGroupTopics(), reader_kafka.PoolSize, readerMessageProcessor.ProcessMessages)

	if err := s.connectKafkaBrokers(ctx); err != nil {
		return errors.Wrap(err, "s.connectKafkaBrokers")
	}
	defer s.kafkaConnection.Close() // nolint: errcheck

	s.runHealthCheck(ctx)
	s.runMetrics(cancel)

	closeGrpcServer, grpcServer, err := s.newReaderGrpcServer()
	if err != nil {
		return errors.Wrap(err, "NewScmGrpcServer")
	}
	defer closeGrpcServer() // nolint: errcheck

	<-ctx.Done()
	grpcServer.GracefulStop()
	return nil
}
