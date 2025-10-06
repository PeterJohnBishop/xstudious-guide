package server

import (
	"net/http"
	"strconv"
	"strings"
	"xstudious-guide/authentication"
	location "xstudious-guide/maps"

	"github.com/gin-gonic/gin"
	"googlemaps.github.io/maps"
)

func GetDirections(client *maps.Client) gin.HandlerFunc {
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

		a := c.Query("origin")
		b := c.Query("destination")
		if a == "" || b == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing origin (a) or destination (b)"})
			return
		}

		route, err := location.GetRoute(client, a, b)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get route"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Route Found!",
			"route":   route,
		})
	}
}

func Geocode(client *maps.Client) gin.HandlerFunc {
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

		// Get query param "address"
		address := c.Query("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing address parameter"})
			return
		}

		result, err := location.Geocode(client, address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to geocode address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Geocoding successful!",
			"result":  result,
		})
	}
}

func ReverseGeocode(client *maps.Client) gin.HandlerFunc {
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

		latStr := c.Query("lat")
		longStr := c.Query("long")
		if latStr == "" || longStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing lat or long parameter"})
			return
		}

		lat64, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to convert latitude to float64"})
			return
		}
		long64, err := strconv.ParseFloat(longStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to convert longitude to float64"})
			return
		}

		result, err := location.ReverseGeocode(client, lat64, long64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reverse geocode coordinates"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Reverse Geocoding successful!",
			"result":  result,
		})
	}
}
