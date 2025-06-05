package repository

import (
	"context"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/utils"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/models"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/otel/trace"
)

type mongoRepository struct {
	log    logger.Logger
	cfg    *config.Config
	mg     *mongo.Client
	tracer trace.Tracer
}

func NewMongoRepository(
	log logger.Logger,
	cfg *config.Config,
	mg *mongo.Client,
	tracer trace.Tracer,
) *mongoRepository {
	return &mongoRepository{log: log, cfg: cfg, mg: mg, tracer: tracer}
}

func (p *mongoRepository) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	ctx, span := p.tracer.Start(ctx, "mongoRepository.CreateProduct")
	defer span.End()

	collection := p.mg.Database(p.cfg.Mongo.Db).Collection(p.cfg.MongoCollections.Products)

	_, err := collection.InsertOne(ctx, product)
	if err != nil {
		p.traceErr(span, err)
		return nil, errors.Wrap(err, "InsertOne")
	}

	return product, nil
}

func (p *mongoRepository) UpdateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	ctx, span := p.tracer.Start(ctx, "mongoRepository.UpdateProduct")
	defer span.End()

	collection := p.mg.Database(p.cfg.Mongo.Db).Collection(p.cfg.MongoCollections.Products)

	ops := options.FindOneAndUpdate()
	ops.SetReturnDocument(options.After)
	ops.SetUpsert(true)

	var updated models.Product
	if err := collection.FindOneAndUpdate(ctx, bson.M{"_id": product.ProductID}, bson.M{"$set": product}, ops).Decode(&updated); err != nil {
		p.traceErr(span, err)
		return nil, errors.Wrap(err, "Decode")
	}

	return &updated, nil
}

func (p *mongoRepository) GetProductById(ctx context.Context, uuid uuid.UUID) (*models.Product, error) {
	ctx, span := p.tracer.Start(ctx, "mongoRepository.GetProductById")
	defer span.End()

	collection := p.mg.Database(p.cfg.Mongo.Db).Collection(p.cfg.MongoCollections.Products)

	var product models.Product
	if err := collection.FindOne(ctx, bson.M{"_id": uuid.String()}).Decode(&product); err != nil {
		p.traceErr(span, err)
		return nil, errors.Wrap(err, "Decode")
	}

	return &product, nil
}

func (p *mongoRepository) DeleteProduct(ctx context.Context, uuid uuid.UUID) error {
	ctx, span := p.tracer.Start(ctx, "mongoRepository.DeleteProduct")
	defer span.End()

	collection := p.mg.Database(p.cfg.Mongo.Db).Collection(p.cfg.MongoCollections.Products)

	return collection.FindOneAndDelete(ctx, bson.M{"_id": uuid.String()}).Err()
}

func (p *mongoRepository) Search(ctx context.Context, search string, pagination *utils.Pagination) (*models.ProductsList, error) {
	ctx, span := p.tracer.Start(ctx, "mongoRepository.Search")
	defer span.End()

	collection := p.mg.Database(p.cfg.Mongo.Db).Collection(p.cfg.MongoCollections.Products)

	filter := bson.D{
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "name", Value: bson.Regex{Pattern: search, Options: "gi"}}},
			bson.D{{Key: "description", Value: bson.Regex{Pattern: search, Options: "gi"}}},
		}},
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		p.traceErr(span, err)
		return nil, errors.Wrap(err, "CountDocuments")
	}
	if count == 0 {
		return &models.ProductsList{Products: make([]*models.Product, 0)}, nil
	}

	limit := int64(pagination.GetLimit())
	skip := int64(pagination.GetOffset())
	cursor, err := collection.Find(ctx, filter, &options.FindOptions{
		Skip: &skip,
		Limit: &limit,
	})
	if err != nil {
		p.traceErr(span, err)
		return nil, errors.Wrap(err, "Find")
	}
	defer cursor.Close(ctx) // nolint: errcheck

	products := make([]*models.Product, 0, pagination.GetSize())

	for cursor.Next(ctx) {
		var prod models.Product
		if err := cursor.Decode(&prod); err != nil {
			p.traceErr(span, err)
			return nil, errors.Wrap(err, "Find")
		}
		products = append(products, &prod)
	}

	if err := cursor.Err(); err != nil {
		span.SetTag("error", true)
		span.LogKV("error_code", err.Error())
		return nil, errors.Wrap(err, "cursor.Err")
	}

	return models.NewProductListWithPagination(products, count, pagination), nil
}

func (p *mongoRepository) traceErr(span trace.Span, err error) {
	span.SetTag("error", true)
	span.LogKV("error_code", err.Error())
}
