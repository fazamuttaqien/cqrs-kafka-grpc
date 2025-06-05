package commands

import (
	"context"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/models"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/repository"
	"go.opentelemetry.io/otel/trace"
)

type UpdateProductCmdHandler interface {
	Handle(ctx context.Context, command *UpdateProductCommand) error
}

type updateProductCmdHandler struct {
	log             logger.Logger
	cfg             *config.Config
	mongoRepository repository.Repository
	redisCache      repository.CacheRepository
	tracer          trace.Tracer
}

func NewUpdateProductCmdHandler(
	log logger.Logger, cfg *config.Config,
	mongoRepository repository.Repository, redisCache repository.CacheRepository,
	tracer trace.Tracer,
) *updateProductCmdHandler {
	return &updateProductCmdHandler{
		log: log, cfg: cfg,
		mongoRepository: mongoRepository, redisCache: redisCache,
		tracer: tracer,
	}
}

func (c *updateProductCmdHandler) Handle(ctx context.Context, command *UpdateProductCommand) error {
	ctx, span := c.tracer.Start(ctx, "updateProductCmdHandler.Handle")
	defer span.End()

	product := &models.Product{
		ProductID:   command.ProductID,
		Name:        command.Name,
		Description: command.Description,
		Price:       command.Price,
		UpdatedAt:   command.UpdatedAt,
	}

	updated, err := c.mongoRepository.UpdateProduct(ctx, product)
	if err != nil {
		return err
	}

	c.redisCache.PutProduct(ctx, updated.ProductID, updated)
	return nil
}
