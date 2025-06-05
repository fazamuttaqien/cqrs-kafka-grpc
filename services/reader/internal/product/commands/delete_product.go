package commands

import (
	"context"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/repository"
	"go.opentelemetry.io/otel/trace"
)

type DeleteProductCmdHandler interface {
	Handle(ctx context.Context, command *DeleteProductCommand) error
}

type deleteProductCmdHandler struct {
	log             logger.Logger
	cfg             *config.Config
	mongoRepository repository.Repository
	redisCache      repository.CacheRepository
	tracer          trace.Tracer
}

func NewDeleteProductCmdHandler(
	log logger.Logger, cfg *config.Config,
	mongoRepository repository.Repository,
	redisCache repository.CacheRepository,
	tracer trace.Tracer,
) *deleteProductCmdHandler {
	return &deleteProductCmdHandler{
		log: log, cfg: cfg,
		mongoRepository: mongoRepository, redisCache: redisCache,
		tracer: tracer,
	}
}

func (c *deleteProductCmdHandler) Handle(ctx context.Context, command *DeleteProductCommand) error {
	ctx, span := c.tracer.Start(ctx, "deleteProductCmdHandler.Handle")
	defer span.End()

	if err := c.mongoRepository.DeleteProduct(ctx, command.ProductID); err != nil {
		return err
	}

	c.redisCache.DelProduct(ctx, command.ProductID.String())
	return nil
}
