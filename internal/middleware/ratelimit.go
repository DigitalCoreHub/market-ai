package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimit creates a rate limiting middleware
func RateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        60,                // 60 requests
		Expiration: 1 * time.Minute,   // per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Rate limit by IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Rate limit exceeded. Please try again later.",
			})
		},
	})
}

// StrictRateLimit creates a stricter rate limiting middleware (for sensitive endpoints)
func StrictRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,                // 10 requests
		Expiration: 1 * time.Minute,   // per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Rate limit exceeded. Please try again later.",
			})
		},
	})
}

