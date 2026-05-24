// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

func Root(c fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "v1 online...",
	})
}
