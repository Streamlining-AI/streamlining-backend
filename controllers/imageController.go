package controllers

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/Streamlining-AI/streamlining-backend/database"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

const maxUploadSize = 2 * 1024 * 1024 // 2 mb

var minioClient *minio.Client = database.MinioClient

func RandToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func UploadFileHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		// c.Request.ParseMultipartForm(maxUploadSize)
		if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
			fmt.Printf("Could not parse multipart form: %v\n", err)
			c.JSON(400, gin.H{"message": "Could not parse multipart form: %v"})
			return
		}

		c.Request.FormFile("uploadFile")
		// parse and validate file and post parameters
		file, fileHeader, err := c.Request.FormFile("uploadFile")
		if err != nil {
			c.JSON(400, gin.H{"message": "Could not get upload header."})
			return
		}
		defer file.Close()
		// Get and print out file size
		fileSize := fileHeader.Size
		fmt.Printf("File size (bytes): %v\n", fileSize)
		// validate file size
		if fileSize > maxUploadSize {
			c.JSON(400, gin.H{"message": "Image file size > 2mb."})
			return
		}
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			c.JSON(400, gin.H{"message": "Could not read file."})
			return
		}
		buffer := bytes.NewBuffer(fileBytes)

		// check file type, detectcontenttype only needs the first 512 bytes
		detectedFileType := http.DetectContentType(fileBytes)
		switch detectedFileType {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
		case "application/pdf":
			break
		default:
			c.JSON(400, gin.H{"message": "File is not image."})
			return
		}
		fileName := RandToken(12)
		fileEndings, err := mime.ExtensionsByType(detectedFileType)
		if err != nil {
			c.JSON(500, gin.H{"message": "Failed to upload."})
			return
		}

		// Check if the bucket already exists
		exists, err := minioClient.BucketExists(context.Background(), "mybucket")
		if err != nil {
			fmt.Println(err)
			return
		}

		// If the bucket doesn't exist, create it
		if !exists {
			err = minioClient.MakeBucket(context.Background(), "mybucket", minio.MakeBucketOptions{})
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Bucket '%s' created successfully.\n", "mybucket")
		} else {
			fmt.Printf("Bucket '%s' already exists.\n", "mybucket")
		}

		newFileName := fileName + fileEndings[0]

		_, err = minioClient.PutObject(context.Background(), "mybucket", newFileName, buffer, fileSize, minio.PutObjectOptions{
			ContentType: detectedFileType,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"image_url": "/files/" + newFileName})

	}
}

func GetFile() gin.HandlerFunc {

	return func(c *gin.Context) {
		filename := c.Param("filename")

		// Get the object from MinIO.
		object, err := minioClient.GetObject(context.Background(), "mybucket", filename, minio.GetObjectOptions{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Read the object into a byte array.
		objectBytes, err := ioutil.ReadAll(object)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Set the content type for the response based on the file extension.
		contentType := mime.TypeByExtension(filepath.Ext(filename))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		c.Header("Content-Type", contentType)

		// Set the Content-Disposition header to force the browser to download the file.
		// c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		// Write the object bytes to the response.
		c.Data(http.StatusOK, contentType, objectBytes)
	}

}
