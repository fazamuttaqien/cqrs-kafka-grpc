package main

import (
	"flag"
	"log"
	
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/config"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/services/api-gateway/internal/server"
)

// @contact.name Faza Muttaqien
// @contact.url https://github.com/fazamuttaqien
// @contact.email fazamttqn@gmail.com
func main() {
	flag.Parse()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName("ApiGateway")

	s := server.NewServer(appLogger, cfg)
	appLogger.Fatal(s.Run())
}
