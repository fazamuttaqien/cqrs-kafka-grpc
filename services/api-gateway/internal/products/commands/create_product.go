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

type CreateProductCmdHandler interface {
	Handle(ctx context.Context, command *CreateProductCommand) error
}

type createProductHandler struct {
	log           logger.Logger
	cfg           *config.Config
	kafkaProducer kafka_client.Producer
	tracer        trace.Tracer
}

func NewCreateProductHandler(
	log logger.Logger, cfg *config.Config,
	kafkaProducer kafka_client.Producer,
	tracer trace.Tracer,
) *createProductHandler {
	return &createProductHandler{
		log: log, cfg: cfg,
		kafkaProducer: kafkaProducer, tracer: tracer,
	}
}

func (c *createProductHandler) Handle(ctx context.Context, command *CreateProductCommand) error {
	ctx, span := c.tracer.Start(ctx, "createProductHandler.Handle")
	defer span.End()

	dto := &kafka_messages.ProductCreate{
		ProductID:   command.CreateDto.ProductID.String(),
		Name:        command.CreateDto.Name,
		Description: command.CreateDto.Description,
		Price:       command.CreateDto.Price,
	}

	dtoBytes, err := proto.Marshal(dto)
	if err != nil {
		return err
	}

	return c.kafkaProducer.PublishMessage(ctx, kafka.Message{
		Topic:   c.cfg.KafkaTopics.ProductCreate.TopicName,
		Value:   dtoBytes,
		Time:    time.Now().UTC(),
		Headers: tracing.GetKafkaTracingHeadersFromSpanCtx(ctx),
	})
}
