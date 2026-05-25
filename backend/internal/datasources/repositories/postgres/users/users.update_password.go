// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"

	"templatev27/internal/apperror"
	"templatev27/internal/business/domain"
	"templatev27/internal/datasources/records"
	"templatev27/pkg/logger"

	"gorm.io/gorm"
)

func (r *postgreUserRepository) UpdatePassword(ctx context.Context, inDom *domain.User) error {
	const (
		repositoryName = "users"
		funcName       = "UpdatePassword"
		queryName      = "updateUserPassword"
		fileName       = "users.update_password.go"
	)
	userRecord := records.FromUsersV1Domain(inDom)
	// gorm.DeletedAt нь UPDATE-г deleted_at IS NULL мөрүүдээр хязгаарлана.
	// Тодорхой баганын map нь бичилтийг анхны UPDATE-ийн хүрсэн яг гурван
	// баганаар хязгаарладаг.
	var rowsAffected int64
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		res := tx.Model(&records.Users{}).
			Where("id = ?", userRecord.Id).
			Updates(map[string]interface{}{
				"password":            userRecord.Password,
				"password_changed_at": userRecord.PasswordChangedAt,
				"updated_at":          userRecord.UpdatedAt,
			})
		rowsAffected = res.RowsAffected
		return res.Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to update user password", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"query":      queryName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "users",
			"user_id":    userRecord.Id,
		})
		return err
	}
	if rowsAffected == 0 {
		return apperror.NotFound("user not found")
	}
	return nil
}
