// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"templatev27/internal/http/middlewares"
)

// Fiber port-ийн тэмдэглэл: анхны Gin тест нь цэвэр
// AccessLogFormatter(gin.LogFormatterParams) функцийг шалгадаг байсан.
// Fiber-т ижил төстэй formatter hook байхгүй — access log нь middleware
// дотроо гардаг — тиймээс formatter-ийн нэгж тестийг middleware нь доош
// дамжих handler-ийн статус код эсвэл body-г өөрчилдөггүй бөгөөд гинж нь
// амжилт болон алдааны статусуудын аль алинд нь дуустлаа ажилладгийг
// баталгаажуулдаг зан төлөвийн дамжуулах тестээр сольсон.
func TestAccessLogMiddleware_PassesThrough(t *testing.T) {
	app := fiber.New()
	app.Use(middlewares.AccessLogMiddleware())

	app.Get("/ok", func(c fiber.Ctx) error {
		return c.Status(http.StatusOK).SendString("ok")
	})
	app.Get("/boom", func(c fiber.Ctx) error {
		return c.Status(http.StatusInternalServerError).SendString("boom")
	})

	t.Run("2xx passes through unchanged", func(t *testing.T) {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/ok", http.NoBody))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("5xx passes through unchanged", func(t *testing.T) {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/boom", http.NoBody))
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
