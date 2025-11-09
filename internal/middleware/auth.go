package middleware

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

var jwtSecret = []byte("marketai-secret-key-change-in-production") // Set via InitAuth
var apiKey = ""                                                     // Set via InitAuth

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTProtected protects routes with JWT authentication
func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization header required",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired token",
			})
		}

		// Extract claims and store in context
		if claims, ok := token.Claims.(*Claims); ok {
			c.Locals("username", claims.Username)
		}

		return c.Next()
	}
}

// GenerateToken generates a JWT token for a user
func GenerateToken(username string) (string, error) {
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token")
		return "", err
	}

	return tokenString, nil
}

// InitAuth initializes authentication with config values
func InitAuth(jwtSecretKey, apiKeyValue string) {
	if jwtSecretKey != "" {
		jwtSecret = []byte(jwtSecretKey)
	}
	apiKey = apiKeyValue
}

// APIKeyOrJWTProtected allows either API key or JWT token
func APIKeyOrJWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First try API key
		apiKeyHeader := c.Get("X-API-Key")
		if apiKeyHeader == "" {
			// Also check Authorization header with "ApiKey <key>" format
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "ApiKey" {
					apiKeyHeader = parts[1]
				}
			}
		}

		if apiKeyHeader != "" {
			// Validate API key
			if apiKey != "" && apiKeyHeader != apiKey {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error":   "Invalid API key",
				})
			}
			c.Locals("api_key", apiKeyHeader)
			return c.Next()
		}

		// Fall back to JWT
		return JWTProtected()(c)
	}
}

