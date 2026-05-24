// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"
	"time"

	"templatev27/internal/business/domain"
	"templatev27/internal/datasources/records"
	"templatev27/pkg/logger"
)

func (r *postgreUserRepository) ChangeActiveUser(ctx context.Context, inDom *domain.User) (err error) {
	const (
		repositoryName = "users"
		funcName       = "ChangeActiveUser"
		queryName      = "updateUserActive"
		fileName       = "users.change_active_user.go"
	)
	userRecord := records.FromUsersV1Domain(inDom)

	// gorm.DeletedAt нь UPDATE-г deleted_at IS NULL мөрүүдээр хязгаарлана.
	err = r.conn.WithContext(ctx).
		Model(&records.Users{}).
		Where("id = ?", userRecord.Id).
		Updates(map[string]interface{}{
			"active":     userRecord.Active,
			"updated_at": time.Now().UTC(),
		}).Error
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to update user active flag", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"query":      queryName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "users",
			"user_id":    userRecord.Id,
		})
	}
	return
}
