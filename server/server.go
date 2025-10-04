package server

import (
	"log"
	"xstudious-guide/amazon"

	"github.com/gin-gonic/gin"
)

func InitServer() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// connect DynamoDB
	dynamoClient := amazon.ConnectDB()
	AddDynamoDBRoutes(dynamoClient, router)

	// connect S3
	s3Client := amazon.ConnectS3()
	AddS3Routes(s3Client, dynamoClient, router)

	log.Println("Server listening on :8080")
	router.Run(":8080")
}
