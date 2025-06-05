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

type UpdateProductCmdHandler interface {
	Handle(ctx context.Context, command *UpdateProductCommand) error
}

type updateProductCmdHandler struct {
	log           logger.Logger
	cfg           *config.Config
	kafkaProducer kafka_client.Producer
	tracer        trace.Tracer
}

func NewUpdateProductHandler(
	log logger.Logger, cfg *config.Config,
	kafkaProducer kafka_client.Producer,
	tracer trace.Tracer,
) *updateProductCmdHandler {
	return &updateProductCmdHandler{
		log: log, cfg: cfg,
		kafkaProducer: kafkaProducer,
		tracer:        tracer,
	}
}

func (c *updateProductCmdHandler) Handle(ctx context.Context, command *UpdateProductCommand) error {
	ctx, span := c.tracer.Start(ctx, "updateProductCmdHandler.Handle")
	defer span.End()

	dto := &kafka_messages.ProductUpdate{
		ProductID:   command.UpdateDto.ProductID.String(),
		Name:        command.UpdateDto.Name,
		Description: command.UpdateDto.Description,
		Price:       command.UpdateDto.Price,
	}

	dtoBytes, err := proto.Marshal(dto)
	if err != nil {
		return err
	}

	return c.kafkaProducer.PublishMessage(ctx, kafka.Message{
		Topic:   c.cfg.KafkaTopics.ProductUpdate.TopicName,
		Value:   dtoBytes,
		Time:    time.Now().UTC(),
		Headers: tracing.GetKafkaTracingHeadersFromSpanCtx(ctx),
	})
}
