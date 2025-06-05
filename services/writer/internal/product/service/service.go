package service

import (
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/commands"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/queries"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/repository"
	
	"go.opentelemetry.io/otel/trace"
)

type ProductService struct {
	Commands *commands.ProductCommands
	Queries  *queries.ProductQueries
}

func NewProductService(
	log logger.Logger,
	cfg *config.Config,
	pgRepository repository.Repository,
	kafkaProducer kafka.Producer,
	tracer trace.Tracer,
) *ProductService {

	updateProductHandler := commands.NewUpdateProductHandler(log, cfg, pgRepository, kafkaProducer, tracer)
	createProductHandler := commands.NewCreateProductHandler(log, cfg, pgRepository, kafkaProducer, tracer)
	deleteProductHandler := commands.NewDeleteProductHandler(log, cfg, pgRepository, kafkaProducer, tracer)

	getProductByIdHandler := queries.NewGetProductByIdHandler(log, cfg, pgRepository)

	productCommands := commands.NewProductCommands(createProductHandler, updateProductHandler, deleteProductHandler)
	productQueries := queries.NewProductQueries(getProductByIdHandler)

	return &ProductService{Commands: productCommands, Queries: productQueries}
}
