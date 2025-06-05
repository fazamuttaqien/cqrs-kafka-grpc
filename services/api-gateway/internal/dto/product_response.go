package dto

import (
	"time"
	
	pb_reader "github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/proto/product_reader"
)

type ProductResponse struct {
	ProductID   string    `json:"productId"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

func ProductResponseFromGrpc(product *pb_reader.Product) *ProductResponse {
	return &ProductResponse{
		ProductID:   product.GetProductID(),
		Name:        product.GetName(),
		Description: product.GetDescription(),
		Price:       product.GetPrice(),
		CreatedAt:   product.GetCreatedAt().AsTime(),
		UpdatedAt:   product.GetUpdatedAt().AsTime(),
	}
}
