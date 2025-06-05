package commands

import (
	"context"
	"time"

	kafka_client "github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"
	kafka_messages "github.com/fazamuttaqien/cqrs-kafka-grpc/proto/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/config"
	"go.opentelemetry.io/otel/trace"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type DeleteProductCmdHandler interface {
	Handle(ctx context.Context, command *DeleteProductCommand) error
}

type deleteProductHandler struct {
	log           logger.Logger
	cfg           *config.Config
	kafkaProducer kafka_client.Producer
	tracer        trace.Tracer
}

func NewDeleteProductHandler(
	log logger.Logger, cfg *config.Config,
	kafkaProducer kafka_client.Producer,
	tracer trace.Tracer,
) *deleteProductHandler {
	return &deleteProductHandler{
		log: log, cfg: cfg,
		kafkaProducer: kafkaProducer,
		tracer:        tracer,
	}
}

func (c *deleteProductHandler) Handle(ctx context.Context, command *DeleteProductCommand) error {
	ctx, span := c.tracer.Start(ctx, "deleteProductHandler.Handle")
	defer span.End()

	dto := &kafka_messages.ProductDelete{ProductID: command.ProductID.String()}

	dtoBytes, err := proto.Marshal(dto)
	if err != nil {
		return err
	}

	return c.kafkaProducer.PublishMessage(ctx, kafka.Message{
		Topic:   c.cfg.KafkaTopics.ProductDelete.TopicName,
		Value:   dtoBytes,
		Time:    time.Now().UTC(),
		Headers: tracing.GetKafkaTracingHeadersFromSpanCtx(ctx),
	})
}
