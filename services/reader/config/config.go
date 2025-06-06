package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/constants"
	kafka_client "github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/mongodb"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/postgres"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/probes"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/redis"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Reader microservice config path")
}

type Config struct {
	ServiceName      string               `mapstructure:"serviceName"`
	Logger           *logger.Config       `mapstructure:"logger"`
	KafkaTopics      KafkaTopics          `mapstructure:"kafkaTopics"`
	GRPC             GRPC                 `mapstructure:"grpc"`
	Postgres       *postgres.Config     `mapstructure:"postgres"`
	Kafka            *kafka_client.Config `mapstructure:"kafka"`
	Mongo            *mongodb.Config      `mapstructure:"mongo"`
	Redis            *redis.Config        `mapstructure:"redis"`
	MongoCollections MongoCollections     `mapstructure:"mongoCollections"`
	Probes           probes.Config        `mapstructure:"probes"`
	ServiceSettings  ServiceSettings      `mapstructure:"serviceSettings"`
	OTLP             *tracing.Config      `mapstructure:"otlp"`
}

type GRPC struct {
	Port        string `mapstructure:"port"`
	Development bool   `mapstructure:"development"`
}

type MongoCollections struct {
	Products string `mapstructure:"products"`
}

type KafkaTopics struct {
	ProductCreated kafka_client.TopicConfig `mapstructure:"productCreated"`
	ProductUpdated kafka_client.TopicConfig `mapstructure:"productUpdated"`
	ProductDeleted kafka_client.TopicConfig `mapstructure:"productDeleted"`
}

type ServiceSettings struct {
	RedisProductPrefixKey string `mapstructure:"redisProductPrefixKey"`
}

func InitConfig() (*Config, error) {
	if configPath == "" {
		configPathFromEnv := os.Getenv(constants.CONFIG_PATH)
		if configPathFromEnv != "" {
			configPath = configPathFromEnv
		} else {
			getwd, err := os.Getwd()
			if err != nil {
				return nil, errors.Wrap(err, "os.Getwd")
			}
			configPath = fmt.Sprintf("%s/reader_service/config/config.yaml", getwd)
		}
	}

	cfg := &Config{}

	viper.SetConfigType(constants.YAML)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "viper.ReadInConfig")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}

	grpcPort := os.Getenv(constants.GRPC_PORT)
	if grpcPort != "" {
		cfg.GRPC.Port = grpcPort
	}
	postgresHost := os.Getenv(constants.POSTGRES_HOST)
	if postgresHost != "" {
		cfg.Postgres.Host = postgresHost
	}
	postgresPort := os.Getenv(constants.POSTGRES_PORT)
	if postgresPort != "" {
		cfg.Postgres.Port = postgresPort
	}
	mongoURI := os.Getenv(constants.MONGO_URI)
	if mongoURI != "" {
		//cfg.Mongo.URI = "mongodb://host.docker.internal:27017"
		cfg.Mongo.URI = mongoURI
	}
	redisAddr := os.Getenv(constants.REDIS_ADDR)
	if redisAddr != "" {
		cfg.Redis.Addr = redisAddr
	}

	kafkaBrokers := os.Getenv(constants.KAFKA_BROKERS)
	if kafkaBrokers != "" {
		cfg.Kafka.Brokers = []string{kafkaBrokers}
	}

	return cfg, nil
}
