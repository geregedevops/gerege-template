// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
	authuc "templatev27/internal/business/usecases/auth"
	"templatev27/internal/http/datatransfers/requests"
	v1 "templatev27/internal/http/handlers/v1"
	"templatev27/pkg/audit"
	"templatev27/pkg/logger"
	"templatev27/pkg/validators"
)

// SendOTP godoc
// @Summary      Хэрэглэгчийн email рүү OTP код илгээх
// @Description  6 оронтой OTP үүсгэж, TTL-тэйгээр Redis-д хадгалж, async mailer-ээр и-мэйлийг дараалалд оруулна. HTTP хариу нь бодит SMTP хүргэлт дээр биш, дараалалд оруулах үед буцдаг.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      requests.SendOTPRequest  true  "Email to send OTP to"
// @Success      200      {object}  v1.BaseResponse  "OTP enqueued"
// @Failure      404      {object}  v1.BaseResponse  "Email not registered"
// @Failure      400      {object}  v1.BaseResponse  "Account already activated"
// @Failure      422      {object}  v1.BaseResponse  "Validation error"
// @Failure      500      {object}  v1.BaseResponse  "Failed to enqueue mail"
// @Router       /auth/send-otp [post]
func (h Handler) SendOTP(c fiber.Ctx) error {
	const (
		controllerName = "auth"
		funcName       = "SendOTP"
		fileName       = "auth.send_otp.go"
	)
	ctx := c.Context()
	var req requests.SendOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.WarnWithContext(ctx, "SendOTP: invalid request body", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
		})
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		logger.WarnWithContext(ctx, "SendOTP: validation error", logger.Fields{
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

	if err := h.usecase.SendOTP(ctx, authuc.SendOTPRequest{Email: req.Email}); err != nil {
		logger.ErrorWithContext(ctx, "SendOTP failed in controller", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"email":      req.Email,
		})
		return v1.RespondWithError(c, err)
	}

	ev := auditFromFiber(c)
	ev.Type = audit.EventOTPSent
	ev.Success = true
	ev.Email = req.Email
	audit.Record(ev)

	return v1.NewSuccessResponse(c, http.StatusOK, fmt.Sprintf("otp code has been send to %s", req.Email), nil)
}
