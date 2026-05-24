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
	"templatev27/internal/config"
	"templatev27/internal/constants"
	"templatev27/internal/http/middlewares"
)

func newSecRouter() *fiber.App {
	r := fiber.New()
	r.Use(middlewares.SecurityHeadersMiddleware())
	r.Get("/ping", func(c fiber.Ctx) error { return c.SendString("ok") })
	return r
}

func TestSecurityHeaders_DevDefaults(t *testing.T) {
	config.AppConfig.Environment = constants.EnvironmentDevelopment
	r := newSecRouter()
	resp, err := r.Test(httptest.NewRequest(http.MethodGet, "/ping", http.NoBody))
	require.NoError(t, err)

	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"))
	assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "default-src 'none'")
	// HSTS-г development-д илгээх ЁСГҮЙ — localhost browser-уудад энгийн
	// HTTP-г татгалзахыг заах болно.
	assert.Empty(t, resp.Header.Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_ProductionAddsHSTS(t *testing.T) {
	config.AppConfig.Environment = constants.EnvironmentProduction
	t.Cleanup(func() { config.AppConfig.Environment = constants.EnvironmentDevelopment })

	r := newSecRouter()
	resp, err := r.Test(httptest.NewRequest(http.MethodGet, "/ping", http.NoBody))
	require.NoError(t, err)

	hsts := resp.Header.Get("Strict-Transport-Security")
	assert.Contains(t, hsts, "max-age=")
	assert.Contains(t, hsts, "includeSubDomains")
}
