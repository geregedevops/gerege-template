// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package users_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"templatev27/internal/apperror"
	"templatev27/internal/business/domain"
	usersuc "templatev27/internal/business/usecases/users"
	"templatev27/internal/constants"
	v1 "templatev27/internal/http/handlers/v1"
	usershandler "templatev27/internal/http/handlers/v1/users"
	"templatev27/internal/test/mocks"
	jwtpkg "templatev27/pkg/jwt"
)

func bodyString(t *testing.T, resp *http.Response) string {
	t.Helper()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	_ = resp.Body.Close()
	return buf.String()
}

func TestGetUserDataHandler(t *testing.T) {
	newApp := func() *fiber.App {
		return fiber.New(fiber.Config{
			ErrorHandler: func(c fiber.Ctx, err error) error {
				return v1.RespondWithError(c, err)
			},
		})
	}

	build := func(t *testing.T) (*mocks.UsersUsecase, *fiber.App) {
		uc := mocks.NewUsersUsecase(t)
		h := usershandler.NewHandler(uc)
		r := newApp()
		r.Get("/me", func(c fiber.Ctx) error {
			c.Locals(constants.CtxAuthenticatedUserKey, jwtpkg.JwtCustomClaim{
				UserID: "user-1", Email: "patrick@example.com",
			})
			return c.Next()
		}, h.GetUserData)
		return uc, r
	}

	t.Run("happy path returns user data", func(t *testing.T) {
		uc, r := build(t)
		uc.On("GetByEmail", mock.Anything, usersuc.GetByEmailRequest{Email: "patrick@example.com"}).
			Return(usersuc.GetByEmailResponse{User: domain.User{ID: "user-1", Email: "patrick@example.com", Username: "patrick"}}, nil).Once()

		req := httptest.NewRequest("GET", "/me", http.NoBody)
		resp, err := r.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, bodyString(t, resp), "patrick")
	})

	t.Run("usecase NotFound surfaces as 404", func(t *testing.T) {
		uc, r := build(t)
		uc.On("GetByEmail", mock.Anything, usersuc.GetByEmailRequest{Email: "patrick@example.com"}).
			Return(usersuc.GetByEmailResponse{}, apperror.NotFound("user not found")).Once()

		req := httptest.NewRequest("GET", "/me", http.NoBody)
		resp, err := r.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("missing claims returns 401", func(t *testing.T) {
		uc := mocks.NewUsersUsecase(t)
		h := usershandler.NewHandler(uc)
		r := newApp()
		r.Get("/me", h.GetUserData)

		req := httptest.NewRequest("GET", "/me", http.NoBody)
		resp, err := r.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
