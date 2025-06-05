package mongodb

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const (
	connectTimeout  = 30 * time.Second
	maxConnIdleTime = 3 * time.Minute
	minPoolSize     = 20
	maxPoolSize     = 300
)

type Config struct {
	URI      string `mapstructure:"uri"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Db       string `mapstructure:"db"`
}

func NewMongoConnection(ctx context.Context, cfg *Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	client, err := mongo.Connect(
		options.Client().ApplyURI(cfg.URI).
			SetAuth(options.Credential{Username: cfg.User, Password: cfg.Password}).
			SetConnectTimeout(connectTimeout).
			SetMaxConnIdleTime(maxConnIdleTime).
			SetMinPoolSize(minPoolSize).
			SetMaxPoolSize(maxPoolSize))
	if err != nil {
		// If the connection fails, we return nil and the error
		log.Fatalln("Failed to connect to MongoDB:", err)
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		// If the ping fails, we return nil and the error
		log.Fatalln("Failed to ping MongoDB:", err)
		return nil, err
	}

	return client, nil
}
