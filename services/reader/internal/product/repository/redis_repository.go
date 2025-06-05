package repository

import (
	"context"
	"encoding/json"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/models"
	"go.opentelemetry.io/otel/trace"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const (
	redisProductPrefixKey = "reader:product"
)

type redisRepository struct {
	log         logger.Logger
	cfg         *config.Config
	redisClient redis.UniversalClient
	tracer      trace.Tracer
}

func NewRedisRepository(
	log logger.Logger,
	cfg *config.Config,
	redisClient redis.UniversalClient,
	tracer trace.Tracer,
) *redisRepository {
	return &redisRepository{log: log, cfg: cfg, redisClient: redisClient, tracer: tracer}
}

func (r *redisRepository) PutProduct(ctx context.Context, key string, product *models.Product) {
	ctx, span := r.tracer.Start(ctx, "redisRepository.PutProduct")
	defer span.End()

	productBytes, err := json.Marshal(product)
	if err != nil {
		r.log.WarnMsg("json.Marshal", err)
		return
	}

	if err := r.redisClient.HSetNX(ctx, r.getRedisProductPrefixKey(), key, productBytes).Err(); err != nil {
		r.log.WarnMsg("redisClient.HSetNX", err)
		return
	}
	r.log.Debugf("HSetNX prefix: %s, key: %s", r.getRedisProductPrefixKey(), key)
}

func (r *redisRepository) GetProduct(ctx context.Context, key string) (*models.Product, error) {
	ctx, span := r.tracer.Start(ctx, "redisRepository.GetProduct")
	defer span.End()

	productBytes, err := r.redisClient.HGet(ctx, r.getRedisProductPrefixKey(), key).Bytes()
	if err != nil {
		if err != redis.Nil {
			r.log.WarnMsg("redisClient.HGet", err)
		}
		return nil, errors.Wrap(err, "redisClient.HGet")
	}

	var product models.Product
	if err := json.Unmarshal(productBytes, &product); err != nil {
		return nil, err
	}

	r.log.Debugf("HGet prefix: %s, key: %s", r.getRedisProductPrefixKey(), key)
	return &product, nil
}

func (r *redisRepository) DelProduct(ctx context.Context, key string) {
	if err := r.redisClient.HDel(ctx, r.getRedisProductPrefixKey(), key).Err(); err != nil {
		r.log.WarnMsg("redisClient.HDel", err)
		return
	}
	r.log.Debugf("HDel prefix: %s, key: %s", r.getRedisProductPrefixKey(), key)
}

func (r *redisRepository) DelAllProducts(ctx context.Context) {
	if err := r.redisClient.Del(ctx, r.getRedisProductPrefixKey()).Err(); err != nil {
		r.log.WarnMsg("redisClient.HDel", err)
		return
	}
	r.log.Debugf("Del key: %s", r.getRedisProductPrefixKey())
}

func (r *redisRepository) getRedisProductPrefixKey() string {
	if r.cfg.ServiceSettings.RedisProductPrefixKey != "" {
		return r.cfg.ServiceSettings.RedisProductPrefixKey
	}

	return redisProductPrefixKey
}
