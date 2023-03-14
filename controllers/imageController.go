package controllers

import (
	"crypto/rand"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const maxUploadSize = 2 * 1024 * 1024 // 2 mb
var uploadPath = os.TempDir()

func main() {
	r := gin.Default()
	r.GET("/upload", uploadFileHandler())
	r.StaticFS("/files", http.Dir(uploadPath))
	r.Run(":8000")
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func uploadFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// c.Request.ParseMultipartForm(maxUploadSize)
		if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
			fmt.Printf("Could not parse multipart form: %v\n", err)
			c.JSON(400, gin.H{})
			return
		}

		c.Request.FormFile("uploadFile")
		// parse and validate file and post parameters
		file, fileHeader, err := c.Request.FormFile("uploadFile")
		if err != nil {
			c.JSON(400, gin.H{})
			return
		}
		defer file.Close()
		// Get and print out file size
		fileSize := fileHeader.Size
		fmt.Printf("File size (bytes): %v\n", fileSize)
		// validate file size
		if fileSize > maxUploadSize {
			c.JSON(400, gin.H{})
			return
		}
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			c.JSON(400, gin.H{})
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
			c.JSON(400, gin.H{})
			return
		}
		fileName := randToken(12)
		fileEndings, err := mime.ExtensionsByType(detectedFileType)
		if err != nil {
			c.JSON(500, gin.H{})
			return
		}

		newFileName := fileName + fileEndings[0]
		newPath := filepath.Join(uploadPath, newFileName)
		fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			c.JSON(500, gin.H{})
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			c.JSON(500, gin.H{})
			return
		}
		c.JSON(200, gin.H{})

	}
}
