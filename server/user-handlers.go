package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
	"xstudious-guide/amazon"
	"xstudious-guide/authentication"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func ShortUUID() string {
	u := uuid.New()
	return strings.TrimRight(
		strings.NewReplacer("-", "", "_", "", "/", "").Replace(
			base64.URLEncoding.EncodeToString(u[:])),
		"=",
	)
}

func CreateNewUserReq(client *dynamodb.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user amazon.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		id := ShortUUID()

		email := strings.ToLower(user.Email)
		userId := fmt.Sprintf("u_%s", id)

		hashedPassword, err := authentication.HashedPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}

		newUser := map[string]types.AttributeValue{
			"id":       &types.AttributeValueMemberS{Value: userId},
			"name":     &types.AttributeValueMemberS{Value: user.Name},
			"email":    &types.AttributeValueMemberS{Value: email},
			"password": &types.AttributeValueMemberS{Value: hashedPassword},
		}

		if err := amazon.CreateUser(client, "users", newUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		claims := authentication.UserClaims{
			ID:        user.ID,
			Name:      user.Name,
			Email:     email,
			TokenType: "access",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(authentication.AccessTokenTTL).Unix(),
				IssuedAt:  time.Now().Unix(),
				Subject:   user.ID,
			},
		}

		accessToken, err := authentication.NewAccessToken(claims)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access token"})
			return
		}

		refreshClaims := jwt.StandardClaims{
			ExpiresAt: time.Now().Add(authentication.RefreshTokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   user.ID,
		}
		refreshToken, err := authentication.NewRefreshToken(refreshClaims)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create refresh token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":      "User created successfully",
			"user.id":      userId,
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		})
	}
}

func AuthUserReq(client *dynamodb.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		user, err := amazon.GetUserByEmail(client, "users", req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No user found with that email."})
			return
		}

		if !authentication.CheckPasswordHash(req.Password, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Password verification failed"})
			return
		}

		claims := authentication.UserClaims{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			TokenType: "access",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(authentication.AccessTokenTTL).Unix(),
				IssuedAt:  time.Now().Unix(),
				Subject:   user.ID,
			},
		}

		accessToken, err := authentication.NewAccessToken(claims)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access token"})
			return
		}

		refreshClaims := jwt.StandardClaims{
			ExpiresAt: time.Now().Add(authentication.RefreshTokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   user.ID,
		}
		refreshToken, err := authentication.NewRefreshToken(refreshClaims)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create refresh token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Login Success",
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"user":         user,
		})
	}
}

func GetAllUsersReq(client *dynamodb.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			return
		}
		if authentication.ParseAccessToken(token) == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			return
		}

		resp, err := amazon.GetAllUsers(client, "users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
			return
		}

		var users []amazon.User
		for _, item := range resp {
			var user amazon.User
			if err := attributevalue.UnmarshalMap(item, &user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
				return
			}
			users = append(users, user)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Users Found!",
			"users":   users,
		})
	}
}

func GetUserByIDReq(client *dynamodb.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" || authentication.ParseAccessToken(token) == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			return
		}

		resp, err := amazon.GetUserById(client, "users", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		var user amazon.User
		if err := attributevalue.UnmarshalMap(resp, &user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User Found!",
			"user":    user,
		})
	}
}

func UpdateUserReq(client *dynamodb.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" || authentication.ParseAccessToken(token) == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			return
		}

		var user amazon.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := amazon.UpdateUser(client, "users", user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User Updated!"})
	}
}

func UpdatePasswordReq(client *dynamodb.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" || authentication.ParseAccessToken(token) == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			return
		}

		var user amazon.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		hashedPassword, err := authentication.HashedPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = hashedPassword

		if err := amazon.UpdatePassword(client, "users", user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User Password Updated!"})
	}
}

func DeleteUserReq(client *dynamodb.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" || authentication.ParseAccessToken(token) == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			return
		}

		if err := amazon.DeleteUser(client, "users", id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User Deleted!"})
	}
}
