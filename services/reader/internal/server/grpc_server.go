package server

import (
	readerGrpc "github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/internal/product/delivery/grpc"
	pb_reader "github.com/fazamuttaqien/cqrs-kafka-grpc/services/reader/proto/product_reader"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"net"
	"time"
)

const (
	maxConnectionIdle = 5
	gRPCTimeout       = 15
	maxConnectionAge  = 5
	gRPCTime          = 10
)

func (s *server) newReaderGrpcServer() (func() error, *grpc.Server, error) {
	l, err := net.Listen("tcp", s.config.GRPC.Port)
	if err != nil {
		return nil, nil, errors.Wrap(err, "net.Listen")
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: maxConnectionIdle * time.Minute,
			Timeout:           gRPCTimeout * time.Second,
			MaxConnectionAge:  maxConnectionAge * time.Minute,
			Time:              gRPCTime * time.Minute,
		}),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(),
			s.interceptorManager.Logger,
		),
		),
	)

	readerGrpcService := readerGrpc.NewReaderGrpcService(s.log, s.config, s.validate, s.productService, s.metrics)
	pb_reader.RegisterReaderServiceServer(grpcServer, readerGrpcService)
	grpc_prometheus.Register(grpcServer)

	if s.config.GRPC.Development {
		reflection.Register(grpcServer)
	}

	go func() {
		s.log.Infof("Reader gRPC server is listening on port: %s", s.config.GRPC.Port)
		s.log.Fatal(grpcServer.Serve(l))
	}()

	return l.Close, grpcServer, nil
}
