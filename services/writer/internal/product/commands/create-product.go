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
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/models"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/repository"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/mappers"
)

type CreateProductCmdHandler interface {
	Handle(ctx context.Context, command *CreateProductCommand) error
}

type createProductHandler struct {
	log           logger.Logger
	cfg           *config.Config
	pgRepository  repository.Repository
	kafkaProducer kafka_client.Producer
	tracer        trace.Tracer
}

func NewCreateProductHandler(
	log logger.Logger,
	cfg *config.Config,
	pgRepository repository.Repository,
	kafkaProducer kafka_client.Producer,
	tracer trace.Tracer,
) *createProductHandler {
	return &createProductHandler{
		log:           log,
		cfg:           cfg,
		pgRepository:  pgRepository,
		kafkaProducer: kafkaProducer,
		tracer:        tracer,
	}
}

func (c *createProductHandler) Handle(
	ctx context.Context,
	command *CreateProductCommand,
) error {
	ctx, span := c.tracer.Start(ctx, "CreateProductCmdHandler.Handle")
	defer span.End()

	dto := &models.Product{
		ProductID:   command.ProductID,
		Name:        command.Name,
		Price:       command.Price,
		Description: command.Description,
	}

	product, err := c.pgRepository.CreateProduct(ctx, dto)
	if err != nil {
		c.log.Error("failed to create product", "error", err)
		return err
	}

	msg := &kafka_pb.ProductCreated{Product: mappers.ProductToGrpcMessage(product)}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		c.log.Error("failed to marshal product created message", "error", err)
		return err
	}

	message := kafka.Message{
		Topic:   c.cfg.KafkaTopics.ProductCreated.TopicName,
		Value:   msgBytes,
		Time:    time.Now().UTC(),
		Headers: tracing.GetKafkaTracingHeadersFromSpanCtx(ctx),
	}

	return c.kafkaProducer.PublishMessage(ctx, message)
}
