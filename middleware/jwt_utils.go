package middleware

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// JWT secret key (from environment) - lazy loaded
var jwtSecret []byte
var jwtSecretLoaded bool

// Initialize checks and configurations
func init() {
	// We'll load the JWT secret when first needed
}

// GetJWTSecret returns the JWT secret key
func GetJWTSecret() []byte {
	if !jwtSecretLoaded {
		jwtSecret = []byte(os.Getenv("JWT_SECRET"))
		jwtSecretLoaded = true
		
		// Check if JWT_SECRET is set
		if len(jwtSecret) == 0 {
			fmt.Println("Warning: JWT_SECRET environment variable is not set")
		}
	}
	return jwtSecret
}

// ParseJWT validates the token string and returns the claims if valid
func ParseJWT(tokenStr string) (jwt.MapClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
} 