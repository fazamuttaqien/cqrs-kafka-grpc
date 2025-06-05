package products

import "github.com/gofiber/fiber/v2"

type HttpDelivery interface {
	CreateProduct() fiber.Handler
	UpdateProduct() fiber.Handler
	DeleteProduct() fiber.Handler

	GetProductByID() fiber.Handler
	SearchProduct() fiber.Handler
}
