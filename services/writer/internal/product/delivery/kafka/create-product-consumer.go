package kafka

import (
	"context"
	"time"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"
	kafka_pb "github.com/fazamuttaqien/cqrs-kafka-grpc/proto/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/commands"

	"github.com/avast/retry-go"
	"github.com/satori/go.uuid"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

const (
	retryAttempts = 3
	retryDelay    = 300 * time.Millisecond
)

var (
	retryOptions = []retry.Option{retry.Attempts(retryAttempts), retry.Delay(retryDelay), retry.DelayType(retry.BackOffDelay)}
)

func (s *productMessageProcessor) processCreateProduct(
	ctx context.Context,
	r *kafka.Reader,
	m kafka.Message,
) {
	s.metrics.CreateProductKafkaMessages.Inc()

	ctx, span := tracing.StartKafkaConsumerTracerSpan(ctx, m.Headers, "productMessageProcessor.processCreateProduct")
	defer span.End()

	var msg kafka_pb.ProductCreate
	if err := proto.Unmarshal(m.Value, &msg); err != nil {
		s.log.WarnMsg("proto.Unmarshal", err)
		s.commitErrMessage(ctx, r, m)
		return
	}

	proUUID, err := uuid.FromString(msg.GetProductID())
	if err != nil {
		s.log.WarnMsg("proto.Unmarshal", err)
		s.commitErrMessage(ctx, r, m)
		return
	}

	command := commands.NewCreateProductCommand(proUUID, msg.GetName(), msg.GetDescription(), msg.GetPrice())
	if err := s.v.StructCtx(ctx, command); err != nil {
		s.log.WarnMsg("validate", err)
		s.commitErrMessage(ctx, r, m)
		return
	}

	if err := retry.Do(func() error {
		return s.ps.Commands.CreateProduct.Handle(ctx, command)
	}, append(retryOptions, retry.Context(ctx))...); err != nil {
		s.log.WarnMsg("CreateProduct.Handle", err)
		s.metrics.ErrorKafkaMessages.Inc()
		return
	}

	s.commitMessage(ctx, r, m)
}
