package controllers

import (
	"crypto/rand"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	helper "github.com/Streamlining-AI/streamlining-backend/helpers"
	"github.com/gin-gonic/gin"
)

const maxUploadSize = 2 * 1024 * 1024 // 2 mb

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

		uploadPath, err := helper.CreateAndGetDir("data/images")

		if err != nil {
			c.JSON(500, gin.H{"message": "Failed to upload."})
			return
		}
		newFileName := fileName + fileEndings[0]
		newPath := filepath.Join(uploadPath, newFileName)
		fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			c.JSON(500, gin.H{"message": "Failed to upload."})
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			c.JSON(500, gin.H{"message": "Failed to upload."})
			return
		}
		c.JSON(200, gin.H{"image_url": "/files/" + newFileName})

	}
}
