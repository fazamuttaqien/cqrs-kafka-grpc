package queries

import (
	"context"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/dto"
	pb_reader "github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/proto/product_reader"
	
	"go.opentelemetry.io/otel/trace"
)

type GetProductByIdHandler interface {
	Handle(ctx context.Context, query *GetProductByIdQuery) (*dto.ProductResponse, error)
}

type getProductByIdHandler struct {
	log      logger.Logger
	cfg      *config.Config
	rsClient pb_reader.ReaderServiceClient
	tracer   trace.Tracer
}

func NewGetProductByIdHandler(
	log logger.Logger, cfg *config.Config,
	rsClient pb_reader.ReaderServiceClient,
	tracer trace.Tracer,
) *getProductByIdHandler {
	return &getProductByIdHandler{
		log: log, cfg: cfg,
		rsClient: rsClient,
		tracer:   tracer,
	}
}

func (q *getProductByIdHandler) Handle(ctx context.Context, query *GetProductByIdQuery) (*dto.ProductResponse, error) {
	ctx, span := q.tracer.Start(ctx, "getProductByIdHandler.Handle")
	defer span.End()

	ctx = tracing.InjectTextMapCarrierToGrpcMetaData(ctx)
	res, err := q.rsClient.GetProductById(ctx, &pb_reader.GetProductByIdReq{ProductID: query.ProductID.String()})
	if err != nil {
		return nil, err
	}

	return dto.ProductResponseFromGrpc(res.GetProduct()), nil
}
