package middlewares

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
			c.Abort()
			return
		}
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Token has expired"})
				c.Abort()
				return
			}
		}
		var userID string
		if hasuraClaims, ok := claims["https://hasura.io/jwt/claims"].(map[string]interface{}); ok {
			if id, ok := hasuraClaims["x-hasura-user-id"].(string); ok {
				userID = id
			}
		} else if id, ok := claims["X-Hasura-User-Id"].(string); ok {
			userID = id
		}

		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User ID not found in token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)

		c.Next()
	}
}
