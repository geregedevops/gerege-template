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

// ForgotPassword godoc
// @Summary      Нууц үг сэргээх токен олгох
// @Description  Хаяг руу битүүмжилсэн сэргээх токеныг и-мэйлээр илгээнэ. Хэрэглэгчийг тоолохоос сэргийлэхийн тулд email бүртгэлгүй байсан ч үргэлж 200 буцаана.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      requests.ForgotPasswordRequest  true  "Email address"
// @Success      200  {object}  v1.BaseResponse  "Reset email queued (or email not registered — same response either way)"
// @Failure      400  {object}  v1.BaseResponse  "Malformed JSON body"
// @Failure      422  {object}  v1.BaseResponse  "Validation error"
// @Router       /auth/password/forgot [post]
func (h Handler) ForgotPassword(c fiber.Ctx) error {
	const (
		controllerName = "auth"
		funcName       = "ForgotPassword"
		fileName       = "auth.forgot_password.go"
	)
	ctx := c.Context()
	var req requests.ForgotPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.WarnWithContext(ctx, "ForgotPassword: invalid request body", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
		})
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		logger.WarnWithContext(ctx, "ForgotPassword: validation error", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"request": logger.Fields{
				"email": req.Email,
			},
		})
		return v1.RespondWithError(c, err)
	}

	if err := h.usecase.ForgotPassword(ctx, authuc.ForgotPasswordRequest{Email: req.Email}); err != nil {
		ev := auditFromFiber(c)
		ev.Type = audit.EventPasswordForgotFail
		ev.Email = req.Email
		ev.Reason = err.Error()
		audit.Record(ev)
		logger.ErrorWithContext(ctx, "ForgotPassword failed in controller", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"email":      req.Email,
		})
		return v1.RespondWithError(c, err)
	}

	ev := auditFromFiber(c)
	ev.Type = audit.EventPasswordForgotOK
	ev.Success = true
	ev.Email = req.Email
	audit.Record(ev)

	return v1.NewSuccessResponse(c, http.StatusOK, "if the email is registered, a reset link has been sent", nil)
}
