// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

func RootHandler(c fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"status":  true,
		"message": "welcome to an amazing api",
	})
}
