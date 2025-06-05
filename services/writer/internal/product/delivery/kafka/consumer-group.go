package kafka

import (
	"context"
	"sync"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/metrics"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/service"

	"github.com/go-playground/validator"
	"github.com/segmentio/kafka-go"
)

const PoolSize = 30

type productMessageProcessor struct {
	log     logger.Logger
	cfg     *config.Config
	v       *validator.Validate
	ps      *service.ProductService
	metrics *metrics.WriterServiceMetrics
}

func NewProductMessageProcessor(
	log logger.Logger,
	cfg *config.Config,
	v *validator.Validate,
	ps *service.ProductService,
	metrics *metrics.WriterServiceMetrics,
) *productMessageProcessor {
	return &productMessageProcessor{
		log:     log,
		cfg:     cfg,
		v:       v,
		ps:      ps,
		metrics: metrics,
	}
}

func (s *productMessageProcessor) ProcessMessages(
	ctx context.Context,
	reader *kafka.Reader,
	wg *sync.WaitGroup,
	workerID int,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		m, err := reader.FetchMessage(ctx)
		if err != nil {
			s.log.Warnf("workerID %d: FetchMessage error: %v", workerID, err)
			continue
		}

		s.logProcessMessage(m, workerID)

		switch m.Topic {
		case s.cfg.KafkaTopics.ProductCreate.TopicName:
			s.processCreateProduct(ctx, reader, m)
		case s.cfg.KafkaTopics.ProductUpdate.TopicName:
			s.processUpdateProduct(ctx, reader, m)
		case s.cfg.KafkaTopics.ProductDelete.TopicName:
			s.processDeleteProduct(ctx, reader, m)
		}
	}
}
