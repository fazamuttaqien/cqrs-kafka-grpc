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

type UpdateProductCmdHandler interface {
	Handle(ctx context.Context, command *UpdateProductCommand) error
}

type updateProductHandler struct {
	log           logger.Logger
	cfg           *config.Config
	pgRepository  repository.Repository
	kafkaProducer kafka_client.Producer
	tracer        trace.Tracer
}

func NewUpdateProductHandler(
	log logger.Logger,
	cfg *config.Config,
	pgRepository repository.Repository,
	kafkaProducer kafka_client.Producer,
	tracer trace.Tracer,
) *updateProductHandler {
	return &updateProductHandler{
		log:           log,
		cfg:           cfg,
		pgRepository:  pgRepository,
		kafkaProducer: kafkaProducer,
		tracer:        tracer,
	}
}

func (c *updateProductHandler) Handle(ctx context.Context, command *UpdateProductCommand) error {
	ctx, span := c.tracer.Start(ctx, "updateProductHandler.Handle")
	defer span.End()

	dto := &models.Product{
		ProductID:   command.ProductID,
		Name:        command.Name,
		Description: command.Description,
		Price:       command.Price,
	}

	product, err := c.pgRepository.UpdateProduct(ctx, dto)
	if err != nil {
		return err
	}

	msg := &kafka_pb.ProductUpdated{Product: mappers.ProductToGrpcMessage(product)}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Topic:   c.cfg.KafkaTopics.ProductUpdated.TopicName,
		Value:   msgBytes,
		Time:    time.Now().UTC(),
		Headers: tracing.GetKafkaTracingHeadersFromSpanCtx(ctx),
	}

	return c.kafkaProducer.PublishMessage(ctx, message)
}
