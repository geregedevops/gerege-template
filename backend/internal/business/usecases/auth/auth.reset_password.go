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
)

// ResetPassword нь нэг удаагийн reset токеныг (ForgotPassword-оос) хэрэглэж,
// хэрэглэгчийн нууц үгийг солино. Дахин тоглуулах (replay) боломжгүй болгохын
// тулд амжилттай болоход токеныг устгадаг.
func (uc *usecase) ResetPassword(ctx context.Context, req ResetPasswordRequest) (err error) {
	const (
		usecaseName = "auth"
		funcName    = "ResetPassword"
		fileName    = "auth.reset_password.go"
	)
	// RLS: хэрэглэгч нэвтрээгүй (зөвхөн reset токентой) тул нууц үг шинэчлэхэд
	// "service" үүрэг шаардлагатай.
	ctx = asService(ctx)
	startTime := time.Now()
	token := req.Token
	newPassword := req.NewPassword

	logger.InfoWithContext(ctx, fmt.Sprintf("Upper %s", funcName), logger.Fields{
		"usecase": usecaseName,
		"method":  funcName,
		"file":    fileName,
		"request": logger.Fields{
			"has_token":        token != "",
			"has_new_password": newPassword != "",
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

	if newPassword == "" {
		err = apperror.BadRequest("new password is required")
		logger.ErrorWithContext(ctx, "Reset password failed: empty new password", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "validate_new_password",
			"error":   err.Error(),
		})
		return err
	}
	if token == "" {
		err = apperror.BadRequest("reset token is required")
		logger.ErrorWithContext(ctx, "Reset password failed: empty token", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "validate_token",
			"error":   err.Error(),
		})
		return err
	}

	// GetDel-ээр атомаар уншиж-устгана — Get + дараа нь Del нь TOCTOU цоорхой
	// үлдээдэг тул зэрэгцээ хоёр хүсэлт ижил token-оор амжилттай дуусдаг байв.
	// Эндээс эхлээд token нь хэрэгтэй болсон ч "хэрэглэгдсэн" төлөвт оров —
	// доорх алхмын аль нэг нь алдахад токенаа дахин ашиглах боломжгүй болсон
	// гэдгийг хэрэглэгчид нь /password/forgot-оор шинээр авах ёстой.
	userID, getErr := uc.redisCache.GetDel(ctx, PasswordResetKey(token))
	if getErr != nil || userID == "" {
		err = apperror.Unauthorized("reset token is invalid or expired")
		fields := logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "redis_getdel_reset_token",
			"error":   err.Error(),
		}
		if getErr != nil {
			fields["redis_error"] = getErr.Error()
		}
		logger.ErrorWithContext(ctx, "Reset password failed: invalid or expired token", fields)
		return err
	}

	lookupResp, lookupErr := uc.users.GetByID(ctx, users.GetByIDRequest{ID: userID})
	if lookupErr != nil {
		err = lookupErr
		logger.ErrorWithContext(ctx, "Reset password failed: user lookup error", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "get_user_by_id",
			"error":   lookupErr.Error(),
			"user_id": userID,
		})
		return err
	}
	user := lookupResp.User
	if hashErr := user.ChangePassword(newPassword, uc.cfg.BcryptCost); hashErr != nil {
		if errors.Is(hashErr, domain.ErrEmptyPassword) {
			err = apperror.BadRequest(hashErr.Error())
		} else {
			err = apperror.InternalCause(fmt.Errorf("hash reset password: %w", hashErr))
		}
		logger.ErrorWithContext(ctx, "Reset password failed: hash error", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "domain_change_password",
			"error":   hashErr.Error(),
			"user_id": userID,
		})
		return err
	}
	if updateErr := uc.users.UpdatePassword(ctx, users.UpdatePasswordRequest{User: &user}); updateErr != nil {
		err = updateErr
		logger.ErrorWithContext(ctx, "Reset password failed: update error", logger.Fields{
			"usecase": usecaseName,
			"method":  funcName,
			"file":    fileName,
			"step":    "users_update_password",
			"error":   updateErr.Error(),
			"user_id": userID,
		})
		return err
	}
	// PasswordResetKey-г GetDel дээр аль хэдийн устгасан; зөвхөн user→token
	// индексийг л цэвэрлэнэ.
	_ = uc.redisCache.Del(ctx, UserResetIndexKey(userID))
	if user.PasswordChangedAt != nil {
		uc.recordTokenCutoff(ctx, userID, *user.PasswordChangedAt)
	}
	return nil
}
