package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}
}

// DBinstance func
func MinioInstance() *minio.Client {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	MINIO_URL, exists := os.LookupEnv("MINIO_URL")
	if !exists {
		log.Fatal("MONGODB_URL not defined in .env file")
	}

	MINIO_ACCESS, exists := os.LookupEnv("MINIO_ACCESS")
	if !exists {
		log.Fatal("MONGODB_URL not defined in .env file")
	}

	MINIO_SECRET, exists := os.LookupEnv("MINIO_SECRET")
	if !exists {
		log.Fatal("MONGODB_URL not defined in .env file")
	}

	var err error
	minioClient, err := minio.New(MINIO_URL, &minio.Options{
		Creds:  credentials.NewStaticV4(MINIO_ACCESS, MINIO_SECRET, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return minioClient
}

// Client Database instance
var MinioClient *minio.Client = MinioInstance()
