// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"
	"time"

	"templatev27/internal/apperror"
	"templatev27/internal/datasources/records"
	"templatev27/pkg/logger"

	"gorm.io/gorm"
)

func (r *postgreUserRepository) SoftDelete(ctx context.Context, id string) error {
	const (
		repositoryName = "users"
		funcName       = "SoftDelete"
		queryName      = "softDeleteUser"
		fileName       = "users.soft_delete.go"
	)
	// deleted_at + updated_at-г тодорхой тогтоо (GORM-ийн нүцгэн
	// Delete-ийн оронд) — ингэснээр анхны хоёр баганын бичилтийн зан
	// төлөв хадгалагдана. Амьд хэвээр буй мөр дээрх Where нь үйлдлийг
	// idempotent байлгана — аль хэдийн устгагдсан мөрийг gorm.DeletedAt
	// scope алгасч, RowsAffected == 0 гарна.
	now := time.Now().UTC()
	var rowsAffected int64
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		res := tx.Model(&records.Users{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"deleted_at": now,
				"updated_at": now,
			})
		rowsAffected = res.RowsAffected
		return res.Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to soft-delete user", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"query":      queryName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "users",
			"user_id":    id,
		})
		return err
	}
	if rowsAffected == 0 {
		return apperror.NotFound("user not found")
	}
	return nil
}
