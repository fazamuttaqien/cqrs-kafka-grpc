package service

import (
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/commands"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/queries"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/repository"
	"go.opentelemetry.io/otel/trace"
)

type ProductService struct {
	Commands *commands.ProductCommands
	Queries  *queries.ProductQueries
}

func NewProductService(
	log logger.Logger,
	cfg *config.Config,
	mongoRepository repository.Repository,
	redisCache repository.CacheRepository,
	tracer trace.Tracer,
) *ProductService {

	createProductHandler := commands.NewCreateProductHandler(log, cfg, mongoRepository, redisCache, tracer)
	deleteProductCmdHandler := commands.NewDeleteProductCmdHandler(log, cfg, mongoRepository, redisCache, tracer)
	updateProductCmdHandler := commands.NewUpdateProductCmdHandler(log, cfg, mongoRepository, redisCache, tracer)

	getProductByIdHandler := queries.NewGetProductByIdHandler(log, cfg, mongoRepository, redisCache, tracer)
	searchProductHandler := queries.NewSearchProductHandler(log, cfg, mongoRepository, redisCache)

	productCommands := commands.NewProductCommands(createProductHandler, updateProductCmdHandler, deleteProductCmdHandler)
	productQueries := queries.NewProductQueries(getProductByIdHandler, searchProductHandler)

	return &ProductService{Commands: productCommands, Queries: productQueries}
}
