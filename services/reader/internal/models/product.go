package models

import (
	"time"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/utils"
	pb_reader "github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/proto/product_reader"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Product struct {
	ProductID   string    `json:"productId" bson:"_id,omitempty"`
	Name        string    `json:"name,omitempty" bson:"name,omitempty" validate:"required,min=3,max=250"`
	Description string    `json:"description,omitempty" bson:"description,omitempty" validate:"required,min=3,max=500"`
	Price       float64   `json:"price,omitempty" bson:"price,omitempty" validate:"required"`
	CreatedAt   time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

// ProductsList products list response with pagination
type ProductsList struct {
	TotalCount int64      `json:"totalCount" bson:"totalCount"`
	TotalPages int64      `json:"totalPages" bson:"totalPages"`
	Page       int64      `json:"page" bson:"page"`
	Size       int64      `json:"size" bson:"size"`
	HasMore    bool       `json:"hasMore" bson:"hasMore"`
	Products   []*Product `json:"products" bson:"products"`
}

func NewProductListWithPagination(products []*Product, count int64, pagination *utils.Pagination) *ProductsList {
	return &ProductsList{
		TotalCount: count,
		TotalPages: int64(pagination.GetTotalPages(int(count))),
		Page:       int64(pagination.GetPage()),
		Size:       int64(pagination.GetSize()),
		HasMore:    pagination.GetHasMore(int(count)),
		Products:   products,
	}
}

func ProductToGrpcMessage(product *Product) *pb_reader.Product {
	return &pb_reader.Product{
		ProductID:   product.ProductID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
	}
}

func ProductListToGrpc(products *ProductsList) *pb_reader.SearchResp {
	list := make([]*pb_reader.Product, 0, len(products.Products))
	for _, product := range products.Products {
		list = append(list, ProductToGrpcMessage(product))
	}

	return &pb_reader.SearchResp{
		TotalCount: products.TotalCount,
		TotalPages: products.TotalPages,
		Page:       products.Page,
		Size:       products.Size,
		HasMore:    products.HasMore,
		Products:   list,
	}
}
