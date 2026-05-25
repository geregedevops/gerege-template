// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"

	"templatev27/internal/business/domain"
	"templatev27/internal/datasources/records"
	repointerface "templatev27/internal/datasources/repositories/interface"
	"templatev27/pkg/logger"

	"gorm.io/gorm"
)

// hardLimit нь List хуудасны хэмжээг хязгаарладаг тул буруу ажиллаж
// буй дуудагч бүх хүснэгтийг нэг хүсэлтэд татаж чадахгүй. Энэ
// хязгаарыг энд давтах нь (handler-ийн хийдэг ямар ч хязгаарлалтаас
// гадна) гүний хамгаалалт юм.
const hardLimit = 200

func (r *postgreUserRepository) List(ctx context.Context, filter repointerface.UserListFilter, offset, limit int) ([]domain.User, error) {
	const (
		repositoryName = "users"
		funcName       = "List"
		queryName      = "selectUsersList"
		fileName       = "users.list.go"
	)
	if limit <= 0 || limit > hardLimit {
		limit = hardLimit
	}
	if offset < 0 {
		offset = 0
	}

	// Query-г GORM-ийн гинжлэгдэх нөхцлүүдээр бүтээ — утга бүр parameter
	// болж холбогддог, хэзээ ч SQL мөр рүү залгагддаггүй. withRLS нь
	// query-г транзакцид боож, context дахь identity-г RLS GUC болгоно.
	var rows []records.Users
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		q := tx.Model(&records.Users{})
		if filter.IncludeDeleted {
			// deleted_at IS NOT NULL мөрүүдийг оруулахын тулд soft-delete
			// scope-г алгасна.
			q = q.Unscoped()
		}
		// IncludeDeleted нь false үед gorm.DeletedAt нь deleted_at IS NULL
		// предикатыг автоматаар нэмдэг.
		if filter.RoleID != 0 {
			q = q.Where("role_id = ?", filter.RoleID)
		}
		if filter.ActiveOnly {
			q = q.Where("active = ?", true)
		}
		return q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list users", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"query":      queryName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "users",
			"limit":      limit,
			"offset":     offset,
		})
		return nil, err
	}
	out := make([]domain.User, 0, len(rows))
	for i := range rows {
		out = append(out, rows[i].ToV1Domain())
	}
	return out, nil
}
