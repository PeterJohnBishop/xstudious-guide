package server

import (
	"xstudious-guide/authentication"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"googlemaps.github.io/maps"
)

func AddDynamoDBRoutes(client *dynamodb.Client, r *gin.Engine) {
	r.POST("/register", CreateNewUserReq(client))
	r.POST("/login", AuthUserReq(client))
	r.POST("/refresh-token", authentication.RefreshTokenHandler(client))

	auth := r.Group("/", authentication.AuthMiddleware())
	{
		auth.GET("/users", GetAllUsersReq(client))
		auth.GET("/users/:id", GetUserByIDReq(client))
		auth.PUT("/users", UpdateUserReq(client))
		auth.PUT("/users/password", UpdatePasswordReq(client))
		auth.DELETE("/users/:id", DeleteUserReq(client))
	}
}

func AddS3Routes(s3client *s3.Client, dynamoclient *dynamodb.Client, r *gin.Engine) {
	auth := r.Group("/", authentication.AuthMiddleware())
	{
		auth.POST("/upload", Upload(s3client, dynamoclient))
		auth.GET("/files", GetUserFilesHandler(dynamoclient, s3.NewPresignClient(s3client)))
		auth.GET("/download", Download(s3client))
	}
}

func AddMapRoutes(client *maps.Client, r *gin.Engine) {
	auth := r.Group("/", authentication.AuthMiddleware())
	{
		auth.GET("/geocode", Geocode(client))
		auth.GET("/reverse-geocode", ReverseGeocode(client))
		auth.GET("/directions", GetDirections(client))
	}
}
