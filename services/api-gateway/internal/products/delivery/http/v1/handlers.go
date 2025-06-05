package v1

import (
	"net/http"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/constants"
	httpErrors "github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/http/errors"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/utils"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/dto"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/metrics"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/middlewares"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/products/commands"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/products/queries"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/products/service"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-playground/validator"
	"github.com/satori/go.uuid"
)

type productsHandlers struct {
	group   *fiber.Group
	log     logger.Logger
	mw      middlewares.MiddlewareManager
	cfg     *config.Config
	ps      *service.ProductService
	v       *validator.Validate
	metrics *metrics.ApiGatewayMetrics
	tracer  trace.Tracer
}

func NewProductsHandlers(
	group *fiber.Group,
	log logger.Logger,
	mw middlewares.MiddlewareManager,
	cfg *config.Config,
	ps *service.ProductService,
	v *validator.Validate,
	metrics *metrics.ApiGatewayMetrics,
	tracer trace.Tracer,
) *productsHandlers {
	return &productsHandlers{
		group: group, log: log, mw: mw, cfg: cfg, ps: ps, v: v, metrics: metrics, tracer: tracer}
}

// CreateProduct
// @Tags Products
// @Summary Create product
// @Description Create new product item
// @Accept json
// @Produce json
// @Success 201 {object} dto.CreateProductResponseDto
// @Router /products [post]
func (h *productsHandlers) CreateProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h.metrics.CreateProductHttpRequests.Inc()

		ctx, span := tracing.StartHttpServerTracerSpan(c, "productsHandlers.CreateProduct")
		defer span.End()

		createDto := &dto.CreateProductDto{}
		if err := c.BodyParser(createDto); err != nil {
			h.log.WarnMsg("Bind", err)
			h.traceErr(span, err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		createDto.ProductID = uuid.NewV4()
		if err := h.v.StructCtx(ctx, createDto); err != nil {
			h.log.WarnMsg("validate", err)
			h.traceErr(span, err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		if err := h.ps.Commands.CreateProduct.Handle(ctx, commands.NewCreateProductCommand(createDto)); err != nil {
			h.log.WarnMsg("CreateProduct", err)
			h.metrics.ErrorHttpRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.metrics.SuccessHttpRequests.Inc()
		return c.Status(fiber.StatusCreated).JSON(dto.CreateProductResponseDto{ProductID: createDto.ProductID})
	}
}

// GetProductByID
// @Tags Products
// @Summary Get product
// @Description Get product by id
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} dto.ProductResponse
// @Router /products/{id} [get]
func (h *productsHandlers) GetProductByID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h.metrics.GetProductByIdHttpRequests.Inc()

		ctx, span := tracing.StartHttpServerTracerSpan(c, "productsHandlers.GetProductByID")
		defer span.End()

		productUUID, err := uuid.FromString(c.Params(constants.ID))
		if err != nil {
			h.log.WarnMsg("uuid.FromString", err)
			h.traceErr(span, err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		query := queries.NewGetProductByIdQuery(productUUID)
		response, err := h.ps.Queries.GetProductById.Handle(ctx, query)
		if err != nil {
			h.log.WarnMsg("GetProductById", err)
			h.metrics.ErrorHttpRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.metrics.SuccessHttpRequests.Inc()
		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// SearchProduct
// @Tags Products
// @Summary Search product
// @Description Get product by name with pagination
// @Accept json
// @Produce json
// @Param search query string false "search text"
// @Param page query string false "page number"
// @Param size query string false "number of elements"
// @Success 200 {object} dto.ProductsListResponse
// @Router /products/search [get]
func (h *productsHandlers) SearchProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h.metrics.SearchProductHttpRequests.Inc()

		ctx, span := tracing.StartHttpServerTracerSpan(c, "productsHandlers.SearchProduct")
		defer span.End()

		pq := utils.NewPaginationFromQueryParams(c.QueryParam(constants.Size), c.QueryParam(constants.Page))

		query := queries.NewSearchProductQuery(c.QueryParam(constants.Search), pq)
		response, err := h.ps.Queries.SearchProduct.Handle(ctx, query)
		if err != nil {
			h.log.WarnMsg("SearchProduct", err)
			h.metrics.ErrorHttpRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.metrics.SuccessHttpRequests.Inc()
		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// UpdateProduct
// @Tags Products
// @Summary Update product
// @Description Update existing product
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} dto.UpdateProductDto
// @Router /products/{id} [put]
func (h *productsHandlers) UpdateProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h.metrics.UpdateProductHttpRequests.Inc()

		ctx, span := tracing.StartHttpServerTracerSpan(c, "productsHandlers.UpdateProduct")
		defer span.End()

		productUUID, err := uuid.FromString(c.Params(constants.ID))
		if err != nil {
			h.log.WarnMsg("uuid.FromString", err)
			h.traceErr(span, err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		updateDto := &dto.UpdateProductDto{ProductID: productUUID}
		if err := c.BodyParser(updateDto); err != nil {
			h.log.WarnMsg("Bind", err)
			h.traceErr(span, err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		if err := h.v.StructCtx(ctx, updateDto); err != nil {
			h.log.WarnMsg("validate", err)
			h.traceErr(span, err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		if err := h.ps.Commands.UpdateProduct.Handle(ctx, commands.NewUpdateProductCommand(updateDto)); err != nil {
			h.log.WarnMsg("UpdateProduct", err)
			h.metrics.ErrorHttpRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.metrics.SuccessHttpRequests.Inc()
		return c.Status(fiber.StatusOK).JSON(updateDto)
	}
}

// DeleteProduct
// @Tags Products
// @Summary Delete product
// @Description Delete existing product
// @Accept json
// @Produce json
// @Success 200 ""
// @Param id path string true "Product ID"
// @Router /products/{id} [delete]
func (h *productsHandlers) DeleteProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h.metrics.DeleteProductHttpRequests.Inc()

		ctx, span := tracing.StartHttpServerTracerSpan(c, "productsHandlers.DeleteProduct")
		defer span.End()

		productUUID, err := uuid.FromString(c.Params(constants.ID))
		if err != nil {
			h.log.WarnMsg("uuid.FromString", err)
			h.traceErr(span, err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		if err := h.ps.Commands.DeleteProduct.Handle(ctx, commands.NewDeleteProductCommand(productUUID)); err != nil {
			h.log.WarnMsg("DeleteProduct", err)
			h.metrics.ErrorHttpRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.metrics.SuccessHttpRequests.Inc()
		return c.SendStatus(http.StatusOK)
	}
}

func (h *productsHandlers) traceErr(span opentracing.Span, err error) {
	span.SetTag("error", true)
	span.LogKV("error_code", err.Error())
	h.metrics.ErrorHttpRequests.Inc()
}
