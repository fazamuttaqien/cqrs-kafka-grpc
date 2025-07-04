package queries

import (
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/utils"
	"github.com/satori/go.uuid"
)

type ProductQueries struct {
	GetProductById GetProductByIdHandler
	SearchProduct  SearchProductHandler
}

func NewProductQueries(getProductById GetProductByIdHandler, searchProduct SearchProductHandler) *ProductQueries {
	return &ProductQueries{GetProductById: getProductById, SearchProduct: searchProduct}
}

type GetProductByIdQuery struct {
	ProductID uuid.UUID `json:"productId" validate:"required,gte=0,lte=255"`
}

func NewGetProductByIdQuery(productID uuid.UUID) *GetProductByIdQuery {
	return &GetProductByIdQuery{ProductID: productID}
}

type SearchProductQuery struct {
	Text       string            `json:"text"`
	Pagination *utils.Pagination `json:"pagination"`
}

func NewSearchProductQuery(text string, pagination *utils.Pagination) *SearchProductQuery {
	return &SearchProductQuery{Text: text, Pagination: pagination}
}
