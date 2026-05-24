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

func (r *postgreUserRepository) GetByEmail(ctx context.Context, inDom *domain.User) (outDomain domain.User, err error) {
	const (
		repositoryName = "users"
		funcName       = "GetByEmail"
		queryName      = "selectUserByEmail"
		fileName       = "users.get_by_email.go"
	)
	userRecord := records.FromUsersV1Domain(inDom)

	// Soft-delete хийгдсэн мөрүүдийг хас — schema нь deleted_at баганыг
	// хадгалдаг тул "устгагдсан" хэрэглэгчид audit/сэргээх зорилгоор
	// query хийгдэх боломжтой хэвээр үлддэг боловч тэд нэвтрэх эсвэл OTP
	// урсгалуудыг хангах ёсгүй. GORM-ийн gorm.DeletedAt scope нь
	// deleted_at IS NULL предикатыг автоматаар нэмдэг.
	var stored records.Users
	err = r.conn.WithContext(ctx).Where(`"email" = ?`, userRecord.Email).First(&stored).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, apperror.NotFound("user not found")
		}
		logger.ErrorWithContext(ctx, "Failed to query user by email", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"query":      queryName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "users",
			"email":      userRecord.Email,
		})
		return domain.User{}, err
	}

	return stored.ToV1Domain(), nil
}
