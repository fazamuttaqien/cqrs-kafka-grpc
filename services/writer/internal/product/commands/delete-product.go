package commands

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	kafka_client "github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"
	kafka_pb "github.com/fazamuttaqien/cqrs-kafka-grpc/proto/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/repository"
)

type DeleteProductCmdHandler interface {
	Handle(ctx context.Context, command *DeleteProductCommand) error
}

type deleteProductHandler struct {
	log           logger.Logger
	cfg           *config.Config
	pgRepository  repository.Repository
	kafkaProducer kafka_client.Producer
	tracer        trace.Tracer
}

func NewDeleteProductHandler(
	log logger.Logger,
	cfg *config.Config,
	pgRepository repository.Repository,
	kafkaProducer kafka_client.Producer,
	tracer trace.Tracer,
) *deleteProductHandler {
	return &deleteProductHandler{
		log:           log,
		cfg:           cfg,
		pgRepository:  pgRepository,
		kafkaProducer: kafkaProducer,
		tracer:        tracer,
	}
}

func (c *deleteProductHandler) Handle(ctx context.Context, command *DeleteProductCommand) error {
	ctx, span := c.tracer.Start(ctx, "CreateProductCmdHandler.Handle")
	defer span.End()

	if err := c.pgRepository.DeleteProductByID(ctx, command.ProductID); err != nil {
		return err
	}

	msg := &kafka_pb.ProductDeleted{ProductID: command.ProductID.String()}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Topic:   c.cfg.KafkaTopics.ProductDeleted.TopicName,
		Value:   msgBytes,
		Time:    time.Now().UTC(),
		Headers: tracing.GetKafkaTracingHeadersFromSpanCtx(ctx),
	}

	return c.kafkaProducer.PublishMessage(ctx, message)
}
