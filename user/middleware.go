package user

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// JWTMiddleware checks for a valid JWT in the Authorization header.
func JWTMiddleware(userService *UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString := strings.Replace(authHeader, "Bearer ", "", -1)

			// Parse the token and extract the user ID
			userID, err := userService.ParseToken(tokenString)
			if err == nil {
				// Set the user ID in the context if the token is valid
				c.Set("userID", userID)
				// fmt.Println("userID", userID)
			}
			// Note: No error handling here - we allow the request to continue regardless
		}

		// Proceed to the next middleware/handler
		c.Next()
	}
}

// ParseToken parses a token and returns the user email.
func (us *UserService) ParseToken(tokenString string) (string, error) {
	// Get the secret key from the environment variable
	secretKey := os.Getenv("SECRET_KEY")

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}

	// Extract the claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["_id"].(string)
		if !ok {
			return "", errors.New("invalid token claims")
		}
		return userID, nil
	} else {
		return "", errors.New("invalid token")
	}
}
