package queries

import (
	"context"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/models"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/repository"
)

type SearchProductHandler interface {
	Handle(ctx context.Context, query *SearchProductQuery) (*models.ProductsList, error)
}

type searchProductHandler struct {
	log             logger.Logger
	cfg             *config.Config
	mongoRepository repository.Repository
	redisCache      repository.CacheRepository
}

func NewSearchProductHandler(
	log logger.Logger, cfg *config.Config,
	mongoRepository repository.Repository, redisCache repository.CacheRepository,
) *searchProductHandler {
	return &searchProductHandler{
		log: log, cfg: cfg,
		mongoRepository: mongoRepository, redisCache: redisCache,
	}
}

func (s *searchProductHandler) Handle(ctx context.Context, query *SearchProductQuery) (*models.ProductsList, error) {
	return s.mongoRepository.Search(ctx, query.Text, query.Pagination)
}
