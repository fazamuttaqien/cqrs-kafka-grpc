package repository

import (
	"context"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/models"
	"go.opentelemetry.io/otel/trace"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

type productRepository struct {
	log    logger.Logger
	cfg    *config.Config
	pgx    *pgxpool.Pool
	tracer trace.Tracer
}

func NewProductRepository(log logger.Logger, cfg *config.Config, pgx *pgxpool.Pool, tracer trace.Tracer) *productRepository {
	return &productRepository{log: log, cfg: cfg, pgx: pgx, tracer: tracer}
}

func (p *productRepository) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	ctx, span := p.tracer.Start(ctx, "productRepository.CreateProduct")
	defer span.End()

	var created models.Product
	if err := p.pgx.QueryRow(ctx, createProductQuery, &product.ProductID, &product.Name, &product.Description, &product.Price).Scan(
		&created.ProductID,
		&created.Name,
		&created.Description,
		&created.Price,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return nil, errors.Wrap(err, "db.QueryRow")
	}

	return &created, nil
}

func (p *productRepository) UpdateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	ctx, span := p.tracer.Start(ctx, "productRepository.UpdateProduct")
	defer span.End()

	var prod models.Product
	if err := p.pgx.QueryRow(
		ctx,
		updateProductQuery,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.ProductID,
	).Scan(&prod.ProductID, &prod.Name, &prod.Description, &prod.Price, &prod.CreatedAt, &prod.UpdatedAt); err != nil {
		return nil, errors.Wrap(err, "Scan")
	}

	return &prod, nil
}

func (p *productRepository) GetProductById(ctx context.Context, uuid uuid.UUID) (*models.Product, error) {
	ctx, span := p.tracer.Start(ctx, "productRepository.GetProductById")
	defer span.End()

	var product models.Product
	if err := p.pgx.QueryRow(ctx, getProductByIdQuery, uuid).Scan(
		&product.ProductID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.CreatedAt,
		&product.UpdatedAt,
	); err != nil {
		return nil, errors.Wrap(err, "Scan")
	}

	return &product, nil
}

func (p *productRepository) DeleteProductByID(ctx context.Context, uuid uuid.UUID) error {
	ctx, span := p.tracer.Start(ctx, "productRepository.DeleteProductByID")
	defer span.End()

	_, err := p.pgx.Exec(ctx, deleteProductByIdQuery, uuid)
	if err != nil {
		return errors.Wrap(err, "Exec")
	}

	return nil
}
