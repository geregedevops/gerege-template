// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "templatev27/internal/http/handlers/v1"
	"templatev27/internal/http/middlewares"
	"templatev27/pkg/jwt"
)

var (
	jwtService          jwt.JWTService
	s                   *fiber.App
	authBasicMiddleware fiber.Handler
	authAdminMiddleware fiber.Handler
)

const (
	adminEndpoint = "/admin"
	forEveryone   = "/everyone"
)

func authenticatedHandler(c fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"status":  true,
		"message": "nice to meet you again sir...",
	})
}

func setup(t *testing.T) {
	jwtService = jwt.NewJWTService("test-secret-key", "test-issuer", 5)
	authBasicMiddleware = middlewares.NewAuthMiddleware(jwtService, nil, false)
	authAdminMiddleware = middlewares.NewAuthMiddleware(jwtService, nil, true)

	s = fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return v1.RespondWithError(c, err)
		},
	})
	s.Get(forEveryone, authBasicMiddleware, authenticatedHandler)
	s.Get(adminEndpoint, authAdminMiddleware, authenticatedHandler)
}

func generateToken(isAdmin bool) (token string, err error) {
	token, err = jwtService.GenerateToken("ddfcea5c-d919-4a8f-a631-4ace39337s3a", isAdmin, "najibfikri13@gmail.com")
	return
}

func getAdminToken() (string, error) {
	return generateToken(true)
}

func getBasicToken() (string, error) {
	return generateToken(false)
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	_ = resp.Body.Close()
	return buf.String()
}

func TestAuthMiddleware(t *testing.T) {
	setup(t)

	t.Run("Test 1 | Success Get Admin Handler", func(t *testing.T) {
		token, err := getAdminToken()
		if err != nil {
			t.Error(err)
		}

		r := httptest.NewRequest(http.MethodGet, adminEndpoint, http.NoBody)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := s.Test(r)
		require.NoError(t, err)
		body := readBody(t, resp)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		assert.Contains(t, body, "nice to meet you again sir")
	})
	t.Run("Test 2 | Invalid Token", func(t *testing.T) {
		token := "mwehehe"

		r := httptest.NewRequest(http.MethodGet, forEveryone, http.NoBody)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := s.Test(r)
		require.NoError(t, err)
		body := readBody(t, resp)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		assert.Contains(t, body, "invalid token")
	})
	t.Run("Test 3 | Must Content Bearer", func(t *testing.T) {
		token, err := getBasicToken()
		if err != nil {
			t.Error(err)
		}

		r := httptest.NewRequest(http.MethodGet, forEveryone, http.NoBody)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", fmt.Sprintf("Token %s", token))

		resp, err := s.Test(r)
		require.NoError(t, err)
		body := readBody(t, resp)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		assert.Contains(t, body, "token must content bearer")
	})
	t.Run("Test 4 | Invalid Format", func(t *testing.T) {
		token, err := getBasicToken()
		if err != nil {
			t.Error(err)
		}

		r := httptest.NewRequest(http.MethodGet, forEveryone, http.NoBody)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", fmt.Sprintf("Bearer token: %s", token))

		resp, err := s.Test(r)
		require.NoError(t, err)
		body := readBody(t, resp)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		assert.Contains(t, body, "invalid header format")
	})
	t.Run("Test 5 | Not Authorize", func(t *testing.T) {
		token, err := getBasicToken()
		if err != nil {
			t.Error(err)
		}

		r := httptest.NewRequest(http.MethodGet, adminEndpoint, http.NoBody)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := s.Test(r)
		require.NoError(t, err)
		body := readBody(t, resp)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		assert.Contains(t, body, "you don't have access for this action")
	})
}
