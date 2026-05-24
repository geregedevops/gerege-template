// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	authuc "templatev27/internal/business/usecases/auth"
	"templatev27/internal/http/datatransfers/requests"
	v1 "templatev27/internal/http/handlers/v1"
	"templatev27/pkg/audit"
	"templatev27/pkg/logger"
	"templatev27/pkg/validators"
)

// Logout godoc
// @Summary      Refresh токеныг хүчингүй болгох
// @Description  /auth/refresh татгалзахын тулд refresh-токены jti-г Redis-ээс устгана. Access токенууд байгалийн хугацаа дуустлаа хүчинтэй хэвээр үлдэнэ — клиентүүд тэдгээрийг logout хийх үед устгах ёстой.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      requests.RefreshRequest  true  "Refresh token to revoke"
// @Success      200      {object}  v1.BaseResponse  "Logged out"
// @Failure      401      {object}  v1.BaseResponse  "Refresh token invalid"
// @Failure      422      {object}  v1.BaseResponse  "Validation error"
// @Router       /auth/logout [post]
func (h Handler) Logout(c fiber.Ctx) error {
	const (
		controllerName = "auth"
		funcName       = "Logout"
		fileName       = "auth.logout.go"
	)
	ctx := c.Context()
	var req requests.RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.WarnWithContext(ctx, "Logout: invalid request body", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
		})
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		logger.WarnWithContext(ctx, "Logout: validation error", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"request": logger.Fields{
				"has_refresh_token": req.RefreshToken != "",
			},
		})
		return v1.RespondWithError(c, err)
	}

	if err := h.usecase.Logout(ctx, authuc.LogoutRequest{RefreshToken: req.RefreshToken}); err != nil {
		logger.ErrorWithContext(ctx, "Logout failed in controller", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
		})
		return v1.RespondWithError(c, err)
	}

	ev := auditFromFiber(c)
	ev.Type = audit.EventLogout
	ev.Success = true
	audit.Record(ev)

	return v1.NewSuccessResponse(c, http.StatusOK, "logout success", nil)
}
