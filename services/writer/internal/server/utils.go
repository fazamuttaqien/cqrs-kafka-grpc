package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/constants"
	kafka_client "github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
)

const stackSize = 1 << 10 // 1KB

func (s *server) connectKafkaBrokers(ctx context.Context) error {
	kafkaConn, err := kafka_client.NewKafkaConnection(ctx, s.config.Kafka)
	if err != nil {
		return errors.Wrap(err, "kafka.NewKafkaConnection")
	}

	s.kafkaConnection = kafkaConn

	brokers, err := kafkaConn.Brokers()
	if err != nil {
		return errors.Wrap(err, "kafkaConnection.Brokers")
	}

	s.log.Infof("Kafka connected to brokers: %+v", brokers)

	return nil
}

func (s *server) initKafkaTopics(ctx context.Context) {
	controller, err := s.kafkaConnection.Controller()
	if err != nil {
		s.log.WarnMsg("kafkaConnection.Controller", err)
		return
	}

	controllerURI := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	s.log.Infof("kafka controller uri: %s", controllerURI)

	conn, err := kafka.DialContext(ctx, "tcp", controllerURI)
	if err != nil {
		s.log.WarnMsg("initKafkaTopics.DialContext", err)
		return
	}
	defer conn.Close() // nolint: errcheck

	s.log.Infof("established new kafka controller connection: %s", controllerURI)

	productCreateTopic := kafka.TopicConfig{
		Topic:             s.config.KafkaTopics.ProductCreate.TopicName,
		NumPartitions:     s.config.KafkaTopics.ProductCreate.Partitions,
		ReplicationFactor: s.config.KafkaTopics.ProductCreate.ReplicationFactor,
	}

	productCreatedTopic := kafka.TopicConfig{
		Topic:             s.config.KafkaTopics.ProductCreated.TopicName,
		NumPartitions:     s.config.KafkaTopics.ProductCreated.Partitions,
		ReplicationFactor: s.config.KafkaTopics.ProductCreated.ReplicationFactor,
	}

	productUpdateTopic := kafka.TopicConfig{
		Topic:             s.config.KafkaTopics.ProductUpdate.TopicName,
		NumPartitions:     s.config.KafkaTopics.ProductUpdate.Partitions,
		ReplicationFactor: s.config.KafkaTopics.ProductUpdate.ReplicationFactor,
	}

	productUpdatedTopic := kafka.TopicConfig{
		Topic:             s.config.KafkaTopics.ProductUpdated.TopicName,
		NumPartitions:     s.config.KafkaTopics.ProductUpdated.Partitions,
		ReplicationFactor: s.config.KafkaTopics.ProductUpdated.ReplicationFactor,
	}

	productDeleteTopic := kafka.TopicConfig{
		Topic:             s.config.KafkaTopics.ProductDelete.TopicName,
		NumPartitions:     s.config.KafkaTopics.ProductDelete.Partitions,
		ReplicationFactor: s.config.KafkaTopics.ProductDelete.ReplicationFactor,
	}

	productDeletedTopic := kafka.TopicConfig{
		Topic:             s.config.KafkaTopics.ProductDeleted.TopicName,
		NumPartitions:     s.config.KafkaTopics.ProductDeleted.Partitions,
		ReplicationFactor: s.config.KafkaTopics.ProductDeleted.ReplicationFactor,
	}

	if err := conn.CreateTopics(
		productCreateTopic,
		productUpdateTopic,
		productCreatedTopic,
		productUpdatedTopic,
		productDeleteTopic,
		productDeletedTopic,
	); err != nil {
		s.log.WarnMsg("kafkaConn.CreateTopics", err)
		return
	}

	s.log.Infof("Kafka topics created or already exists: %+v", []kafka.TopicConfig{productCreateTopic, productUpdateTopic, productCreatedTopic, productUpdatedTopic, productDeleteTopic, productDeletedTopic})
}

func (s *server) getConsumerGroupTopics() []string {
	return []string{
		s.config.KafkaTopics.ProductCreate.TopicName,
		s.config.KafkaTopics.ProductUpdate.TopicName,
		s.config.KafkaTopics.ProductDelete.TopicName,
	}
}

func (s *server) runHealthCheck(ctx context.Context) {
	health := healthcheck.NewHandler()

	health.AddLivenessCheck(s.config.ServiceName, healthcheck.AsyncWithContext(ctx, func() error {
		return nil
	}, time.Duration(s.config.Probes.CheckIntervalSeconds)*time.Second))

	health.AddReadinessCheck(constants.POSTGRES, healthcheck.AsyncWithContext(ctx, func() error {
		return s.pgxPool.Ping(ctx)
	}, time.Duration(s.config.Probes.CheckIntervalSeconds)*time.Second))

	health.AddReadinessCheck(constants.KAFKA, healthcheck.AsyncWithContext(ctx, func() error {
		_, err := s.kafkaConnection.Brokers()
		if err != nil {
			return err
		}
		return nil
	}, time.Duration(s.config.Probes.CheckIntervalSeconds)*time.Second))

	go func() {
		s.log.Infof("Writer microservice Kubernetes probes listening on port: %s", s.config.Probes.Port)
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
