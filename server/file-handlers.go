package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"xstudious-guide/amazon"
	"xstudious-guide/authentication"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

func Upload(client *s3.Client, dynamo *dynamodb.Client) gin.HandlerFunc {
	presigner := s3.NewPresignClient(client)

	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}
		claims := authentication.ParseAccessToken(token)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			return
		}

		id := ShortUUID()
		fileID := fmt.Sprintf("f_%s", id)
		userID := claims.ID

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve file"})
			return
		}
		defer file.Close()

		fileKey, presignedURL, err := amazon.UploadFile(client, presigner, header.Filename, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		userFile := amazon.UserFile{
			UserID:   userID,
			FileID:   fileID,
			FileKey:  fileKey,
			Uploaded: time.Now().Unix(),
		}

		if err := amazon.SaveUserFile(dynamo, "files", userFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "File uploaded successfully",
			"fileId":       fileID,
			"fileKey":      fileKey,
			"presignedURL": presignedURL,
		})
	}
}

func GetUserFilesHandler(dynamo *dynamodb.Client, presigner *s3.PresignClient) gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}
		claims := authentication.ParseAccessToken(token)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			return
		}

		userID := claims.ID

		files, err := amazon.GetUserFiles(dynamo, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user files"})
			return
		}

		type FileResponse struct {
			FileID       string `json:"fileId"`
			FileKey      string `json:"fileKey"`
			PresignedURL string `json:"presignedURL"`
			Uploaded     int64  `json:"uploaded"`
		}

		var response []FileResponse
		bucketName := os.Getenv("AWS_BUCKET")

		for _, f := range files {
			presignedReq, err := presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(f.FileKey),
			}, s3.WithPresignExpires(15*time.Minute))
			if err != nil {
				continue // skip files with errors
			}

			response = append(response, FileResponse{
				FileID:       f.FileID,
				FileKey:      f.FileKey,
				PresignedURL: presignedReq.URL,
				Uploaded:     f.Uploaded,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"userId": userID,
			"files":  response,
		})
	}
}

func Download(client *s3.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method != http.MethodGet {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}
		claims := authentication.ParseAccessToken(token)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			return
		}

		filename := c.Query("filename")
		if filename == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing filename parameter"})
			return
		}

		url, err := amazon.DownloadFile(client, filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate download URL"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"downloadUrl": url,
		})
	}
}
