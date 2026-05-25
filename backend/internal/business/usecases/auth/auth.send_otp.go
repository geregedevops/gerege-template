// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth

import (
	"context"
	"fmt"
	"time"

	"templatev27/internal/apperror"
	"templatev27/internal/business/domain"
	"templatev27/internal/business/usecases/users"
	"templatev27/pkg/logger"
	"templatev27/pkg/observability"
)

// SendOTP нь GeregeCloud Verify (verify.gecloud.mn) API-руу OTP илгээх
// хүсэлт явуулж, буцсан request_id-г Redis-д хадгална (TTL = OTPTTL).
// VerifyOTP дараа нь тэр request_id-г хэрэглэгчийн оруулсан кодтой тулгаж
// шалгана. Дотооддоо код үүсгэх / SMTP-ээр илгээх / Redis-д код хадгалах
// ажлыг алсын үйлчилгээ хариуцдаг тул template нь зөвхөн request_id-ийн
// амьдрах хугацааг л хянана.
func (uc *usecase) SendOTP(ctx context.Context, req SendOTPRequest) (err error) {
	const (
		usecaseName = "auth"
		funcName    = "SendOTP"
		fileName    = "auth.send_otp.go"
	)
	// RLS: нэвтрэхээс өмнөх email хайлт тул DB рүү "service" үүргээр хандана.
	ctx = asService(ctx)
	startTime := time.Now()
	email := domain.NormalizeEmail(req.Email)

	logger.InfoWithContext(ctx, fmt.Sprintf("Upper %s", funcName), logger.Fields{
		"usecase": usecaseName,
		"method":  funcName,
		"file":    fileName,
		"request": logger.Fields{
			"email": email,
		},
	})

	defer func() {
		duration := time.Since(startTime)
		fields := logger.Fields{
			"usecase":  usecaseName,
			"method":   funcName,
			"file":     fileName,
			"duration": duration.Milliseconds(),
		}
		logger.InfoWithContext(ctx, fmt.Sprintf("Lower %s", funcName), fields)
	}()

	lookupResp, lookupErr := uc.users.GetByEmail(ctx, users.GetByEmailRequest{Email: email})
	if lookupErr != nil {
		err = apperror.NotFound("email not found")
		logger.ErrorWithContext(ctx, "Send OTP failed: user lookup error", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "get_user_by_email",
			"error":   lookupErr.Error(),
			"email":   email,
		})
		return err
	}
	user := lookupResp.User

	if user.Active {
		err = apperror.BadRequest("account already activated")
		logger.ErrorWithContext(ctx, "Send OTP failed: account already activated", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "check_active",
			"error":   err.Error(),
			"user_id": user.ID,
		})
		return err
	}

	if uc.verify == nil {
		err = apperror.InternalCause(fmt.Errorf("verify sender is not configured"))
		logger.ErrorWithContext(ctx, "Send OTP failed: verify sender missing", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "verify_sender_check",
			"error":   err.Error(),
			"email":   email,
		})
		return err
	}

	// Алсаас OTP илгээх — verify.gecloud.mn нь кодыг өөрөө үүсгэж, bcrypt-аар
	// hash-аар хадгалж, brute-force-ийг хязгаарладаг. Бид зөвхөн request_id-г л
	// хадгална. Илгээх алдаа гарвал Redis-д юу ч бичихгүй — хуучирсан
	// request_id-р VerifyOTP андуурахгүйн тулд.
	requestID, sendErr := uc.verify.Send(ctx, email)
	if sendErr != nil {
		observability.ObserveMailerOp("queue_full")
		err = apperror.InternalCause(fmt.Errorf("send otp: %w", sendErr))
		logger.ErrorWithContext(ctx, "Send OTP failed: verify api error", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "verify_send",
			"error":   sendErr.Error(),
			"email":   email,
		})
		return err
	}

	// request_id-г атомар SET ... EX <OTPTTL>-ээр хадгална. VerifyOTP нь
	// энэ түлхүүрийг GetDel-ээр уншиж устгаснаар нэг удаагийн semantic
	// баталгаажна (зэрэгцээ хоёр оролдлого ижил request_id-г replay
	// хийхээс сэргийлнэ).
	otpKey := UserOTPKey(email)
	if cacheErr := uc.redisCache.SetWithTTL(ctx, otpKey, requestID, uc.cfg.OTPTTL); cacheErr != nil {
		observability.ObserveCacheOp("redis", "set", "error")
		err = apperror.InternalCause(fmt.Errorf("persist verify request id: %w", cacheErr))
		logger.ErrorWithContext(ctx, "Send OTP failed: persist request_id error", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "redis_set_request_id",
			"error":   cacheErr.Error(),
			"email":   email,
		})
		return err
	}
	observability.ObserveCacheOp("redis", "set", "ok")
	observability.ObserveMailerOp("sent")

	_ = uc.redisCache.Del(ctx, OTPAttemptsKey(email))

	return nil
}
