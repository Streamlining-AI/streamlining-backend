package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}
}

// DBinstance func
func DBinstance() *mongo.Client {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	MONGODB_URL, exists := os.LookupEnv("MONGODB_URL")
	if !exists {
		log.Fatal("MONGODB_URL not defined in .env file")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(MONGODB_URL))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	return client
}

// Client Database instance
var Client *mongo.Client = DBinstance()

// OpenCollection is a  function makes a connection with a collection in the database
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	CLUSTER_DB, exists := os.LookupEnv("CLUSTER_DB")
	if !exists {
		log.Fatal("CLUSTER_DB not defined in .env file")
	}
	var collection *mongo.Collection = client.Database(CLUSTER_DB).Collection(collectionName)

	return collection
}
