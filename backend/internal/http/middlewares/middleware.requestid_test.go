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
	"templatev27/pkg/logger"
)

func TestRequestIDMiddleware_BridgesIDsToContext(t *testing.T) {
	r := fiber.New()
	r.Use(middlewares.RequestIDMiddleware())

	var seenRequestID, seenTraceID string
	r.Get("/probe", func(c fiber.Ctx) error {
		ctx := c.Context()
		seenRequestID = logger.GetRequestIDFromContext(ctx)
		seenTraceID = logger.GetTraceIDFromContext(ctx)
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/probe", http.NoBody)
	req.Header.Set("X-Request-ID", "abc-from-client")
	resp, err := r.Test(req)
	require.NoError(t, err)

	assert.Equal(t, "abc-from-client", seenRequestID,
		"request_id from client header must be visible to handlers via logger.GetRequestIDFromContext")
	assert.Empty(t, seenTraceID,
		"trace_id stays empty when the tracing middleware isn't mounted; populated end-to-end in production")
	assert.Equal(t, "abc-from-client", resp.Header.Get("X-Request-ID"),
		"request_id must be echoed in the response header")
}

func TestRequestIDMiddleware_GeneratesWhenAbsent(t *testing.T) {
	r := fiber.New()
	r.Use(middlewares.RequestIDMiddleware())

	var seen string
	r.Get("/probe", func(c fiber.Ctx) error {
		seen = logger.GetRequestIDFromContext(c.Context())
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/probe", http.NoBody)
	resp, err := r.Test(req)
	require.NoError(t, err)

	assert.NotEmpty(t, seen, "middleware must generate a UUID when no header is present")
	assert.Equal(t, seen, resp.Header.Get("X-Request-ID"))
}
