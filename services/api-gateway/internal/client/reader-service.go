package client

import (
	"context"
	"time"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/interceptors"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/config"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	backoffLinear  = 100 * time.Millisecond
	backoffRetries = 3
)

func NewReaderServiceConn(ctx context.Context, cfg *config.Config, im interceptors.InterceptorManager) (*grpc.ClientConn, error) {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(backoffLinear)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted),
		grpc_retry.WithMax(backoffRetries),
	}

	readerServiceConn, err := grpc.NewClient(
		cfg.Grpc.ReaderServicePort,
		grpc.WithUnaryInterceptor(im.ClientRequestLoggerInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "grpc.DialContext")
	}

	return readerServiceConn, nil
}
