// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"templatev27/internal/apperror"
	"templatev27/internal/business/domain"
	"templatev27/internal/business/usecases/users"
	"templatev27/pkg/logger"
	"templatev27/pkg/observability"
	"templatev27/pkg/verify"
)

// VerifyOTP нь Redis-д хадгалсан request_id-г олж, түүнийг хэрэглэгчийн
// оруулсан кодтой хамт verify.gecloud.mn-руу шалгуулна. Алсын сервис нь
// brute force, нэг удаагийн ашиглалт, hash-аар хадгалалтыг өөрөө хариуцдаг
// тул template нь зөвхөн request_id-ийн амьдрах хугацааг л хянана. Дотоодын
// email тус бүрийн оролдлогын тоолуурыг defense-in-depth байдлаар хадгалж
// үлдсэн — Config.OTPMaxAttempts босгоог хэтэрвэл зөв кодтой ч 403 буцна.
func (uc *usecase) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (err error) {
	const (
		usecaseName = "auth"
		funcName    = "VerifyOTP"
		fileName    = "auth.verify_otp.go"
	)
	// RLS: баталгаажаагүй хэрэглэгчийг идэвхжүүлдэг тул "service" үүрэг хэрэгтэй.
	ctx = asService(ctx)
	startTime := time.Now()
	email := domain.NormalizeEmail(req.Email)
	otpCode := req.OTPCode

	logger.InfoWithContext(ctx, fmt.Sprintf("Upper %s", funcName), logger.Fields{
		"usecase": usecaseName,
		"method":  funcName,
		"file":    fileName,
		"request": logger.Fields{
			"email":        email,
			"has_otp_code": otpCode != "",
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
		logger.ErrorWithContext(ctx, "Verify OTP failed: user lookup error", logger.Fields{
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
		logger.ErrorWithContext(ctx, "Verify OTP failed: account already activated", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "check_active",
			"error":   err.Error(),
			"user_id": user.ID,
		})
		return err
	}

	attemptsKey := OTPAttemptsKey(email)
	attempts, incrErr := uc.incrWithExpiry(ctx, attemptsKey, uc.cfg.OTPTTL, "otp_attempts")
	if incrErr != nil {
		logger.ErrorWithContext(ctx, "Verify OTP: failed to track attempts (non-fatal)", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "redis_incr_attempts",
			"error":   incrErr.Error(),
			"email":   email,
		})
	}
	if attempts > int64(uc.cfg.OTPMaxAttempts) {
		err = apperror.Forbidden("too many invalid otp attempts, please request a new code")
		logger.ErrorWithContext(ctx, "Verify OTP failed: lockout (max attempts exceeded)", logger.Fields{
			"usecase":  usecaseName,
			"method":   funcName,
			"file":     fileName,
			"step":     "check_lockout",
			"error":    err.Error(),
			"email":    email,
			"attempts": attempts,
		})
		return err
	}

	if uc.verify == nil {
		err = apperror.InternalCause(fmt.Errorf("verify sender is not configured"))
		logger.ErrorWithContext(ctx, "Verify OTP failed: verify sender missing", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "verify_sender_check",
			"error":   err.Error(),
			"email":   email,
		})
		return err
	}

	// request_id-г атомар GetDel-ээр уншиж устгана — нэг request_id-р replay
	// хийхийг хааж байна. Алсын /check нь өөрөө нэг удаагийн semantic-той
	// гэдгийг баталдаг ч defense-in-depth-ийн үүднээс энд бас барьж байна.
	otpKey := UserOTPKey(email)
	requestID, getErr := uc.redisCache.GetDel(ctx, otpKey)
	if getErr != nil || requestID == "" {
		observability.ObserveCacheOp("redis", "get", "miss")
		err = apperror.BadRequest("otp code expired or not found")
		fields := logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "redis_getdel_request_id",
			"error":   err.Error(),
			"email":   email,
		}
		if getErr != nil {
			fields["redis_error"] = getErr.Error()
		}
		logger.ErrorWithContext(ctx, "Verify OTP failed: request_id expired or not found", fields)
		return err
	}
	observability.ObserveCacheOp("redis", "get", "hit")

	if checkErr := uc.verify.Check(ctx, requestID, otpCode); checkErr != nil {
		// Код буруу үед request_id хэрэглэгдэх боломжтой хэвээр үлдэхийн тулд
		// Redis-д буцааж бичнэ (тэр TTL-тэйгээр) — өөрөөр бол нэг буруу
		// тоглуулалт хууль ёсны хэрэглэгчийг шинээр SendOTP хийхэд хүргэнэ.
		// Алсын сервис өөрөө attempts-ийг хязгаарладаг тул дахин ашиглах нь
		// аюулгүй.
		if !errors.Is(checkErr, verify.ErrNotApproved) {
			err = apperror.InternalCause(fmt.Errorf("verify otp: %w", checkErr))
		} else {
			err = apperror.BadRequest("invalid otp code")
		}
		_ = uc.redisCache.SetWithTTL(ctx, otpKey, requestID, uc.cfg.OTPTTL)
		logger.ErrorWithContext(ctx, "Verify OTP failed: verify api error", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "verify_check",
			"error":   checkErr.Error(),
			"email":   email,
		})
		return err
	}

	if activateErr := uc.users.Activate(ctx, users.ActivateRequest{UserID: user.ID}); activateErr != nil {
		err = activateErr
		logger.ErrorWithContext(ctx, "Verify OTP failed: activate error", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "users_activate",
			"error":   activateErr.Error(),
			"user_id": user.ID,
		})
		return err
	}

	_ = uc.redisCache.Del(ctx, attemptsKey)

	return nil
}
