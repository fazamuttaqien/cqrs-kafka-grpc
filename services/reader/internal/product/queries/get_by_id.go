package queries

import (
	"context"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/models"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/repository"
	"go.opentelemetry.io/otel/trace"
)

type GetProductByIdHandler interface {
	Handle(ctx context.Context, query *GetProductByIdQuery) (*models.Product, error)
}

type getProductByIdHandler struct {
	log             logger.Logger
	cfg             *config.Config
	mongoRepository repository.Repository
	redisCache      repository.CacheRepository
	tracer          trace.Tracer
}

func NewGetProductByIdHandler(
	log logger.Logger, cfg *config.Config,
	mongoRepository repository.Repository, redisCache repository.CacheRepository,
	tracer trace.Tracer,
) *getProductByIdHandler {
	return &getProductByIdHandler{
		log: log, cfg: cfg,
		mongoRepository: mongoRepository, redisCache: redisCache,
		tracer: tracer,
	}
}

func (q *getProductByIdHandler) Handle(ctx context.Context, query *GetProductByIdQuery) (*models.Product, error) {
	ctx, span := q.tracer.Start(ctx, "getProductByIdHandler.Handle")
	defer span.End()

	if product, err := q.redisCache.GetProduct(ctx, query.ProductID.String()); err == nil && product != nil {
		return product, nil
	}

	product, err := q.mongoRepository.GetProductById(ctx, query.ProductID)
	if err != nil {
		return nil, err
	}

	q.redisCache.PutProduct(ctx, product.ProductID, product)
	return product, nil
}
