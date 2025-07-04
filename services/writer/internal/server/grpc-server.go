package server

import (
	"net"
	"time"

	delivery_grpc "github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/internal/product/delivery/grpc"
	pb_writer "github.com/fazamuttaqien/cqrs-kafka-grpc/services/writer/proto/product_writer"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

const (
	maxConnectionIdle = 5
	gRPCTimeout       = 15
	maxConnectionAge  = 5
	gRPCTime          = 10
)

func (s *server) newWriterGrpcServer() (func() error, *grpc.Server, error) {
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

	writerGrpcWriter := delivery_grpc.NewWriterGrpcService(s.log, s.config, s.validate, s.productService, s.metrics)
	pb_writer.RegisterWriterServiceServer(grpcServer, writerGrpcWriter)
	grpc_prometheus.Register(grpcServer)

	if s.config.GRPC.Development {
		reflection.Register(grpcServer)
	}

	go func() {
		s.log.Infof("Writer gRPC server is listening on port: %s", s.config.GRPC.Port)
		s.log.Fatal(grpcServer.Serve(l))
	}()

	return l.Close, grpcServer, nil
}
