// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"templatev27/internal/apperror"
	authuc "templatev27/internal/business/usecases/auth"
	"templatev27/internal/constants"
	v1 "templatev27/internal/http/handlers/v1"
	authhandler "templatev27/internal/http/handlers/v1/auth"
	"templatev27/internal/test/mocks"
	jwtpkg "templatev27/pkg/jwt"
	"templatev27/pkg/validators"
)

func init() {
	_ = validators.ValidatePayloads(struct{}{})
}

type authHarness struct {
	uc     *mocks.AuthUsecase
	router *fiber.App
}

func newAuthHarness(t *testing.T) authHarness {
	t.Helper()
	uc := mocks.NewAuthUsecase(t)
	h := authhandler.NewHandler(uc)
	// Төвлөрсөн алдааны handler нь production сервертэй ижил байдаг тул
	// handler-уудын буцаасан алдаанууд ижил статус буулгалттайгаар
	// үзүүлэгдэнэ.
	r := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return v1.RespondWithError(c, err)
		},
	})
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	r.Post("/password/forgot", h.ForgotPassword)
	r.Post("/password/reset", h.ResetPassword)
	r.Put("/password/change", injectClaims("user-1", "patrick@example.com"), h.ChangePassword)
	return authHarness{uc: uc, router: r}
}

func injectClaims(userID, email string) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals(constants.CtxAuthenticatedUserKey, jwtpkg.JwtCustomClaim{
			UserID: userID,
			Email:  email,
		})
		return c.Next()
	}
}

func doJSON(t *testing.T, h authHarness, method, path string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.router.Test(req)
	require.NoError(t, err)
	return resp
}

func bodyString(t *testing.T, resp *http.Response) string {
	t.Helper()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	_ = resp.Body.Close()
	return buf.String()
}

func TestLoginHandler(t *testing.T) {
	t.Run("happy path returns 200 and tokens", func(t *testing.T) {
		h := newAuthHarness(t)
		h.uc.On("Login", mock.Anything, authuc.LoginRequest{Email: "patrick@example.com", Password: "Pwd_123!"}).
			Return(authuc.LoginResponse{
				AccessToken: "access-tok", RefreshToken: "refresh-tok",
			}, nil).Once()

		resp := doJSON(t, h, "POST", "/login", map[string]string{
			"email": "patrick@example.com", "password": "Pwd_123!",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, bodyString(t, resp), "access-tok")
	})

	t.Run("malformed JSON returns 400", func(t *testing.T) {
		h := newAuthHarness(t)
		req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("{bad json"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := h.router.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("validation failure returns 422", func(t *testing.T) {
		h := newAuthHarness(t)
		resp := doJSON(t, h, "POST", "/login", map[string]string{
			"email": "not-an-email", "password": "p",
		})
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("usecase Unauthorized returns 401", func(t *testing.T) {
		h := newAuthHarness(t)
		h.uc.On("Login", mock.Anything, authuc.LoginRequest{Email: "x@y.com", Password: "wrong"}).
			Return(authuc.LoginResponse{}, apperror.Unauthorized("invalid email or password")).Once()
		resp := doJSON(t, h, "POST", "/login", map[string]string{
			"email": "x@y.com", "password": "wrong",
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestForgotPasswordHandler(t *testing.T) {
	t.Run("happy path returns 200", func(t *testing.T) {
		h := newAuthHarness(t)
		h.uc.On("ForgotPassword", mock.Anything, authuc.ForgotPasswordRequest{Email: "patrick@example.com"}).Return(nil).Once()
		resp := doJSON(t, h, "POST", "/password/forgot", map[string]string{"email": "patrick@example.com"})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("infra error from usecase still returns non-2xx", func(t *testing.T) {
		h := newAuthHarness(t)
		h.uc.On("ForgotPassword", mock.Anything, authuc.ForgotPasswordRequest{Email: "patrick@example.com"}).
			Return(apperror.InternalCause(assertErr("redis down"))).Once()
		resp := doJSON(t, h, "POST", "/password/forgot", map[string]string{"email": "patrick@example.com"})
		assert.GreaterOrEqual(t, resp.StatusCode, 500)
	})
}

func TestResetPasswordHandler(t *testing.T) {
	t.Run("happy path returns 200", func(t *testing.T) {
		h := newAuthHarness(t)
		h.uc.On("ResetPassword", mock.Anything, authuc.ResetPasswordRequest{Token: "tok-1", NewPassword: "Newpwd_9999!"}).Return(nil).Once()
		resp := doJSON(t, h, "POST", "/password/reset", map[string]string{
			"token": "tok-1", "new_password": "Newpwd_9999!",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		h := newAuthHarness(t)
		h.uc.On("ResetPassword", mock.Anything, authuc.ResetPasswordRequest{Token: "stale", NewPassword: "Newpwd_9999!"}).
			Return(apperror.Unauthorized("reset token is invalid or expired")).Once()
		resp := doJSON(t, h, "POST", "/password/reset", map[string]string{
			"token": "stale", "new_password": "Newpwd_9999!",
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestChangePasswordHandler(t *testing.T) {
	t.Run("happy path returns 200 with claims injected", func(t *testing.T) {
		h := newAuthHarness(t)
		h.uc.On("ChangePassword", mock.Anything, authuc.ChangePasswordRequest{
			UserID:          "user-1",
			CurrentPassword: "Pwd_123!",
			NewPassword:     "Newpwd_9999!",
		}).Return(nil).Once()
		resp := doJSON(t, h, "PUT", "/password/change", map[string]string{
			"current_password": "Pwd_123!", "new_password": "Newpwd_9999!",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("usecase Unauthorized when current password wrong", func(t *testing.T) {
		h := newAuthHarness(t)
		h.uc.On("ChangePassword", mock.Anything, authuc.ChangePasswordRequest{
			UserID:          "user-1",
			CurrentPassword: "wrong",
			NewPassword:     "Newpwd_9999!",
		}).Return(apperror.Unauthorized("current password is incorrect")).Once()
		resp := doJSON(t, h, "PUT", "/password/change", map[string]string{
			"current_password": "wrong", "new_password": "Newpwd_9999!",
		})
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func assertErr(s string) error { return &simpleErr{msg: s} }

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }
