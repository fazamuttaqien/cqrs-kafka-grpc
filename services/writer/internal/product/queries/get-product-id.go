package queries

import (
	"context"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/models"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/repository"
)

type GetProductByIdHandler interface {
	Handle(ctx context.Context, query *GetProductByIdQuery) (*models.Product, error)
}

type getProductByIdHandler struct {
	log          logger.Logger
	cfg          *config.Config
	pgRepository repository.Repository
}

func NewGetProductByIdHandler(
	log logger.Logger,
	cfg *config.Config,
	pgRepository repository.Repository) *getProductByIdHandler {
	return &getProductByIdHandler{
		log:          log,
		cfg:          cfg,
		pgRepository: pgRepository,
	}
}

func (q *getProductByIdHandler) Handle(ctx context.Context, query *GetProductByIdQuery) (*models.Product, error) {
	return q.pgRepository.GetProductById(ctx, query.ProductID)
}
