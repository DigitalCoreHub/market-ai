package handlers

import (
	"github.com/1batu/market-ai/internal/config"
	"github.com/1batu/market-ai/internal/middleware"
	"github.com/1batu/market-ai/internal/models"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	apiKey string
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		apiKey: cfg.Auth.APIKey,
	}
}

type LoginRequest struct {
	APIKey string `json:"api_key" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Type  string `json:"type"`
}

// Login handles API key authentication and returns JWT token
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Validate API key
	if h.apiKey == "" || req.APIKey != h.apiKey {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Response{
			Success: false,
			Message: "Invalid API key",
		})
	}

	// Generate JWT token for API key
	token, err := middleware.GenerateToken("api_key_user")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.Response{
			Success: false,
			Message: "Failed to generate token",
		})
	}

	return c.JSON(models.Response{
		Success: true,
		Data: LoginResponse{
			Token: token,
			Type:  "Bearer",
		},
	})
}

