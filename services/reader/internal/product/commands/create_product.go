package commands

import (
	"context"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/models"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/repository"

	"go.opentelemetry.io/otel/trace"
)

type CreateProductCmdHandler interface {
	Handle(ctx context.Context, command *CreateProductCommand) error
}

type createProductHandler struct {
	log             logger.Logger
	cfg             *config.Config
	mongoRepository repository.Repository
	redisCache      repository.CacheRepository
	tracer          trace.Tracer
}

func NewCreateProductHandler(
	log logger.Logger,
	cfg *config.Config,
	mongoRepository repository.Repository,
	redisCache repository.CacheRepository,
	tracer trace.Tracer,
) *createProductHandler {
	return &createProductHandler{
		log: log, cfg: cfg,
		mongoRepository: mongoRepository, redisCache: redisCache,
		tracer: tracer,
	}
}

func (c *createProductHandler) Handle(ctx context.Context, command *CreateProductCommand) error {
	ctx, span := c.tracer.Start(ctx, "createProductHandler.Handle")
	defer span.End()

	product := &models.Product{
		ProductID:   command.ProductID,
		Name:        command.Name,
		Description: command.Description,
		Price:       command.Price,
		CreatedAt:   command.CreatedAt,
		UpdatedAt:   command.UpdatedAt,
	}

	created, err := c.mongoRepository.CreateProduct(ctx, product)
	if err != nil {
		return err
	}

	c.redisCache.PutProduct(ctx, created.ProductID, created)
	return nil
}
