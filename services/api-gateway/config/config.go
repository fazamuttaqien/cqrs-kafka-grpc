package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/constants"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/kafka"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/logger"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/probes"
	"github.com/fazamuttaqien/cqrs-kafka-grpc/pkg/tracing"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "API Gateway microservice config path")
}

type Config struct {
	ServiceName string          `mapstructure:"serviceName"`
	Logger      *logger.Config  `mapstructure:"logger"`
	KafkaTopics KafkaTopics     `mapstructure:"kafkaTopics"`
	Http        Http            `mapstructure:"http"`
	Grpc        Grpc            `mapstructure:"grpc"`
	Kafka       *kafka.Config   `mapstructure:"kafka"`
	Probes      probes.Config   `mapstructure:"probes"`
	OTLP        *tracing.Config `mapstructure:"otlp"`
}

type Http struct {
	Port                string   `mapstructure:"port"`
	Development         bool     `mapstructure:"development"`
	BasePath            string   `mapstructure:"basePath"`
	ProductsPath        string   `mapstructure:"productsPath"`
	DebugHeaders        bool     `mapstructure:"debugHeaders"`
	HttpClientDebug     bool     `mapstructure:"httpClientDebug"`
	DebugErrorsResponse bool     `mapstructure:"debugErrorsResponse"`
	IgnoreLogUrls       []string `mapstructure:"ignoreLogUrls"`
}

type Grpc struct {
	ReaderServicePort string `mapstructure:"readerServicePort"`
}

type KafkaTopics struct {
	ProductCreate kafka.TopicConfig `mapstructure:"productCreate"`
	ProductUpdate kafka.TopicConfig `mapstructure:"productUpdate"`
	ProductDelete kafka.TopicConfig `mapstructure:"productDelete"`
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
			configPath = fmt.Sprintf("%s/api_gateway_service/config/config.yaml", getwd)
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

	httpPort := os.Getenv(constants.HTTP_PORT)
	if httpPort != "" {
		cfg.Http.Port = httpPort
	}

	kafkaBrokers := os.Getenv(constants.KAFKA_BROKERS)
	if kafkaBrokers != "" {
		cfg.Kafka.Brokers = []string{kafkaBrokers}
	}

	otlpEndpoint := os.Getenv(constants.OTLP_ENDPOINT)
	if otlpEndpoint != "" {
		cfg.OTLP.Endpoint = otlpEndpoint
	}

	readerServicePort := os.Getenv(constants.READER_SERVICE_PORT)
	if readerServicePort != "" {
		cfg.Grpc.ReaderServicePort = readerServicePort
	}

	return cfg, nil
}
