package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

func InitMongoDB() {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_NAME")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("Mongo connection error: %v", err)
	}

	// Пинг для проверки
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Mongo ping error: %v", err)
	}

	MongoClient = client
	MongoDB = client.Database(dbName)
}

func GetMongoDB() *mongo.Database {
	return MongoDB
}

func Database() *mongo.Database {
	return MongoClient.Database("friendship")
}
