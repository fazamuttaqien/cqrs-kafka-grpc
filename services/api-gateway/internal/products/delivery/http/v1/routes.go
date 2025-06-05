package v1

import "github.com/gofiber/fiber/v2"

func (h *productsHandlers) MapRoutes() {
	h.group.Post("/create", h.CreateProduct())
	h.group.Get("/:id", h.GetProductByID())
	h.group.Get("/search", h.SearchProduct())
	h.group.Put("/:id", h.UpdateProduct())
	h.group.Delete("/:id", h.DeleteProduct())
	h.group.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"Status": "OK",
		})
	})
}
