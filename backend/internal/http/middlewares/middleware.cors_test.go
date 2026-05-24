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

func TestCORSMiddleware(t *testing.T) {
	router := fiber.New()
	router.Use(middlewares.CORSMiddleware())
	router.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("test")
	})

	t.Run("Test 1 | CORS Headers Are Set", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Origin", "http://localhost:3000")
		resp, err := router.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	})
	t.Run("Test 2 | Preflight OPTIONS Returns 204", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		resp, err := router.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
		assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Authorization")
	})
	t.Run("Test 3 | Extra Headers Are Not Blocked", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("X-Custom-Header", "something")
		resp, err := router.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
