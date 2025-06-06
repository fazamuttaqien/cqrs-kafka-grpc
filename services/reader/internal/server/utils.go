package server

import (
	"context"
	"net/http"
	"time"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/constants"
	kafka_client "github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	stackSize = 1 << 10 // 1 KB
)

func (s *server) connectKafkaBrokers(ctx context.Context) error {
	kafkaConn, err := kafka_client.NewKafkaConnection(ctx, s.config.Kafka)
	if err != nil {
		return errors.Wrap(err, "kafka.NewKafkaCon")
	}

	s.kafkaConnection = kafkaConn

	brokers, err := kafkaConn.Brokers()
	if err != nil {
		return errors.Wrap(err, "kafkaConn.Brokers")
	}

	s.log.Infof("kafka connected to brokers: %+v", brokers)

	return nil
}

func (s *server) getConsumerGroupTopics() []string {
	return []string{
		s.config.KafkaTopics.ProductCreated.TopicName,
		s.config.KafkaTopics.ProductUpdated.TopicName,
		s.config.KafkaTopics.ProductDeleted.TopicName,
	}
}

func (s *server) runHealthCheck(ctx context.Context) {
	health := healthcheck.NewHandler()

	health.AddLivenessCheck(s.config.ServiceName, healthcheck.AsyncWithContext(ctx, func() error {
		return nil
	}, time.Duration(s.config.Probes.CheckIntervalSeconds)*time.Second))

	health.AddReadinessCheck(constants.REDIS, healthcheck.AsyncWithContext(ctx, func() error {
		return s.redisClient.Ping(ctx).Err()
	}, time.Duration(s.config.Probes.CheckIntervalSeconds)*time.Second))

	health.AddReadinessCheck(constants.MONGODB, healthcheck.AsyncWithContext(ctx, func() error {
		return s.mongoClient.Ping(ctx, nil)
	}, time.Duration(s.config.Probes.CheckIntervalSeconds)*time.Second))

	health.AddReadinessCheck(constants.KAFKA, healthcheck.AsyncWithContext(ctx, func() error {
		_, err := s.kafkaConnection.Brokers()
		if err != nil {
			return err
		}
		return nil
	}, time.Duration(s.config.Probes.CheckIntervalSeconds)*time.Second))

	go func() {
		s.log.Infof("Reader microservice Kubernetes probes listening on port: %s", s.config.Probes.Port)
		if err := http.ListenAndServe(s.config.Probes.Port, health); err != nil {
			s.log.WarnMsg("ListenAndServe", err)
		}
	}()
}

func (s *server) runMetrics(cancel context.CancelFunc) {
	metricsServer := fiber.New()
	metricsServer.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	go func() {
		// Convert Prometheus HTTP handler to Fiber handler
		promHandler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
		metricsServer.Get(s.config.Probes.PrometheusPath, func(c *fiber.Ctx) error {
			promHandler(c.Context())
			return nil
		})

		s.log.Infof("Metrics server is running on port: %s", s.config.Probes.PrometheusPort)
		if err := metricsServer.Listen(s.config.Probes.PrometheusPort); err != nil {
			s.log.Errorf("metricsServer.Listen: %v", err)
			cancel()
		}
	}()
}
