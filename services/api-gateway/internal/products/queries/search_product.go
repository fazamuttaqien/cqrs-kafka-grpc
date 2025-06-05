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

type SearchProductHandler interface {
	Handle(ctx context.Context, query *SearchProductQuery) (*dto.ProductsListResponse, error)
}

type searchProductHandler struct {
	log      logger.Logger
	cfg      *config.Config
	rsClient pb_reader.ReaderServiceClient
	tracer   trace.Tracer
}

func NewSearchProductHandler(
	log logger.Logger, cfg *config.Config,
	rsClient pb_reader.ReaderServiceClient,
	tracer trace.Tracer,
) *searchProductHandler {
	return &searchProductHandler{
		log: log, cfg: cfg,
		rsClient: rsClient,
		tracer:   tracer,
	}
}

func (s *searchProductHandler) Handle(ctx context.Context, query *SearchProductQuery) (*dto.ProductsListResponse, error) {
	ctx, span := s.tracer.Start(ctx, "searchProductHandler.Handle")
	defer span.End()

	ctx = tracing.InjectTextMapCarrierToGrpcMetaData(ctx)
	res, err := s.rsClient.SearchProduct(ctx, &pb_reader.SearchReq{
		Search: query.Text,
		Page:   int64(query.Pagination.GetPage()),
		Size:   int64(query.Pagination.GetSize()),
	})
	if err != nil {
		return nil, err
	}

	return dto.ProductsListResponseFromGrpc(res), nil
}
