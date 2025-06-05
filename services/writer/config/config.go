package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/constants"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/postgres"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/probes"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Writer microservice microservice config path")
}

type Config struct {
	ServiceName string           `mapstructure:"serviceName"`
	Logger      *logger.Config   `mapstructure:"logger"`
	KafkaTopics KafkaTopics      `mapstructure:"kafkaTopics"`
	GRPC        GRPC             `mapstructure:"grpc"`
	Postgresql  *postgres.Config `mapstructure:"postgres"`
	Kafka       *kafka.Config    `mapstructure:"kafka"`
	Probes      probes.Config    `mapstructure:"probes"`
	Jaeger      *tracing.Config  `mapstructure:"jaeger"`
}

type GRPC struct {
	Port        string `mapstructure:"port"`
	Development bool   `mapstructure:"development"`
}

type KafkaTopics struct {
	ProductCreate  kafka.TopicConfig `mapstructure:"productCreate"`
	ProductCreated kafka.TopicConfig `mapstructure:"productCreated"`
	ProductUpdate  kafka.TopicConfig `mapstructure:"productUpdate"`
	ProductUpdated kafka.TopicConfig `mapstructure:"productUpdated"`
	ProductDelete  kafka.TopicConfig `mapstructure:"productDelete"`
	ProductDeleted kafka.TopicConfig `mapstructure:"productDeleted"`
}

func InitConfig() (*Config, error) {
	if configPath == "" {
		configPathFromEnv := os.Getenv(constants.ConfigPath)
		if configPathFromEnv != "" {
			configPath = configPathFromEnv
		} else {
			getwd, err := os.Getwd()
			if err != nil {
				return nil, errors.Wrap(err, "os.Getwd")
			}
			configPath = fmt.Sprintf("%s/writer_service/config/config.yaml", getwd)
		}
	}

	cfg := &Config{}

	viper.SetConfigType(constants.Yaml)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "viper.ReadInConfig")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}

	grpcPort := os.Getenv(constants.GrpcPort)
	if grpcPort != "" {
		cfg.GRPC.Port = grpcPort
	}

	postgresHost := os.Getenv(constants.PostgresqlHost)
	if postgresHost != "" {
		cfg.Postgresql.Host = postgresHost
	}
	postgresPort := os.Getenv(constants.PostgresqlPort)
	if postgresPort != "" {
		cfg.Postgresql.Port = postgresPort
	}
	jaegerAddr := os.Getenv(constants.JaegerHostPort)
	if jaegerAddr != "" {
		cfg.Jaeger.HostPort = jaegerAddr
	}
	kafkaBrokers := os.Getenv(constants.KafkaBrokers)
	if kafkaBrokers != "" {
		cfg.Kafka.Brokers = []string{kafkaBrokers}
	}

	return cfg, nil
}
