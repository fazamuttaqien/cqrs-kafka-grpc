package service

import (
	kafka_client "github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/products/commands"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/products/queries"
	pb_reader "github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/proto/product_reader"
	
	"go.opentelemetry.io/otel/trace"
)

type ProductService struct {
	Commands *commands.ProductCommands
	Queries  *queries.ProductQueries
}

func NewProductService(
	log logger.Logger, cfg *config.Config,
	kafkaProducer kafka_client.Producer, rsClient pb_reader.ReaderServiceClient,
	tracer trace.Tracer,
) *ProductService {

	createProductHandler := commands.NewCreateProductHandler(log, cfg, kafkaProducer, tracer)
	updateProductHandler := commands.NewUpdateProductHandler(log, cfg, kafkaProducer, tracer)
	deleteProductHandler := commands.NewDeleteProductHandler(log, cfg, kafkaProducer, tracer)

	getProductByIdHandler := queries.NewGetProductByIdHandler(log, cfg, rsClient, tracer)
	searchProductHandler := queries.NewSearchProductHandler(log, cfg, rsClient, tracer)

	productCommands := commands.NewProductCommands(createProductHandler, updateProductHandler, deleteProductHandler)
	productQueries := queries.NewProductQueries(getProductByIdHandler, searchProductHandler)

	return &ProductService{Commands: productCommands, Queries: productQueries}
}
