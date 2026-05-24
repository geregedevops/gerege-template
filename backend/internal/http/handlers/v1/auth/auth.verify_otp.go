// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"templatev27/internal/apperror"
	authuc "templatev27/internal/business/usecases/auth"
	"templatev27/internal/http/datatransfers/requests"
	v1 "templatev27/internal/http/handlers/v1"
	"templatev27/pkg/audit"
	"templatev27/pkg/logger"
	"templatev27/pkg/validators"
)

// VerifyOTP godoc
// @Summary      OTP кодыг шалгаж бүртгэлийг идэвхжүүлэх
// @Description  Өгөгдсөн кодыг Redis-тэй тулгаж шалгаж, амжилттай үед хэрэглэгчийн active flag-г true болгоно. Brute-force-оос хамгаалагдсан — OTP_MAX_ATTEMPTS алдаа (өгөгдмөл 5) гарсны дараа зөв кодтой ч email нь OTP TTL цонхны турш түгжигдэнэ.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      requests.VerifyOTPRequest  true  "Email + OTP code"
// @Success      200      {object}  v1.BaseResponse  "Account activated"
// @Failure      400      {object}  v1.BaseResponse  "Invalid OTP code"
// @Failure      403      {object}  v1.BaseResponse  "Locked out — too many invalid attempts"
// @Failure      404      {object}  v1.BaseResponse  "Email not registered"
// @Failure      422      {object}  v1.BaseResponse  "Validation error"
// @Router       /auth/verify-otp [post]
func (h Handler) VerifyOTP(c fiber.Ctx) error {
	const (
		controllerName = "auth"
		funcName       = "VerifyOTP"
		fileName       = "auth.verify_otp.go"
	)
	ctx := c.Context()
	var req requests.VerifyOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.WarnWithContext(ctx, "VerifyOTP: invalid request body", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
		})
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		logger.WarnWithContext(ctx, "VerifyOTP: validation error", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"request": logger.Fields{
				"email":    req.Email,
				"has_code": req.Code != "",
			},
		})
		return v1.RespondWithError(c, err)
	}

	if err := h.usecase.VerifyOTP(ctx, authuc.VerifyOTPRequest{Email: req.Email, OTPCode: req.Code}); err != nil {
		ev := auditFromFiber(c)
		ev.Email = req.Email
		ev.Reason = err.Error()
		// "Та буруу оруулсан"-г "бид бүртгэлийг түгжсэн"-ээс ялгаж байна
		// — rate-limit + сэрэмжлүүлэг өөр өөр дохио шаарддаг.
		var domErr *apperror.DomainError
		if errors.As(err, &domErr) && domErr.Type == apperror.ErrTypeForbidden {
			ev.Type = audit.EventOTPLockout
		} else {
			ev.Type = audit.EventOTPVerifyFail
		}
		audit.Record(ev)
		logger.ErrorWithContext(ctx, "VerifyOTP failed in controller", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"email":      req.Email,
		})
		return v1.RespondWithError(c, err)
	}

	ev := auditFromFiber(c)
	ev.Type = audit.EventOTPVerifyOK
	ev.Success = true
	ev.Email = req.Email
	audit.Record(ev)

	return v1.NewSuccessResponse(c, http.StatusOK, "otp verification success", nil)
}
