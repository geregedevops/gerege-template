// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"
	"errors"

	"templatev27/internal/apperror"
	"templatev27/internal/business/domain"
	"templatev27/internal/datasources/records"
	"templatev27/pkg/logger"

	"gorm.io/gorm"
)

func (r *postgreUserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	const (
		repositoryName = "users"
		funcName       = "GetByID"
		queryName      = "selectUserByID"
		fileName       = "users.get_by_id.go"
	)
	var stored records.Users
	// GORM-ийн soft-delete (gorm.DeletedAt) нь үүнийг автоматаар
	// deleted_at IS NULL байх мөрүүдээр хязгаарлана.
	err := r.conn.WithContext(ctx).Where("id = ?", id).First(&stored).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, apperror.NotFound("user not found")
		}
		logger.ErrorWithContext(ctx, "Failed to query user by id", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"query":      queryName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "users",
			"user_id":    id,
		})
		return domain.User{}, err
	}
	return stored.ToV1Domain(), nil
}
