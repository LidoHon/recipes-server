package middlewares

import (
	"net/http"
	"strings"

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
			return []byte("your-secret-key"), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
			c.Abort()
			return
		}

		// Extract x-hasura-user-id from the token
		hasuraClaims, ok := claims["https://hasura.io/jwt/claims"].(map[string]interface{})
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid Hasura claims"})
			c.Abort()
			return
		}

		userID, ok := hasuraClaims["x-hasura-user-id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User ID not found in token"})
			c.Abort()
			return
		}

		// Set user_id in the context
		c.Set("user_id", userID)
		c.Next()
	}
}
