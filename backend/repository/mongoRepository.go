package repository

import (
	"context"
	"friendship/config"
	"friendship/logger"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository interface {
	DB() *mongo.Database
	Collection(name string) *mongo.Collection
}

type mongoRepository struct {
	mongo *mongo.Database
}

func NewMongoRepository(logger logger.Logger, conf *config.Config) MongoRepository {
	logger.Info("Try MongoDB connection")

	db, err := initMongo(conf.Mongo.URI, conf.Mongo.DB, logger)
	if err != nil {
		logger.Error("Failure MongoDB connection", "error", err)
		os.Exit(1)
	}

	logger.Info("Success MongoDB connection")
	return &mongoRepository{mongo: db}
}

func initMongo(uri, dbName string, logger logger.Logger) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	logger.Info("MongoDB connected to", "uri", uri)
	return client.Database(dbName), nil
}

func (m *mongoRepository) DB() *mongo.Database {
	return m.mongo
}

func (m *mongoRepository) Collection(name string) *mongo.Collection {
	return m.mongo.Collection(name)
}
