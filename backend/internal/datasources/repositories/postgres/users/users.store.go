// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"
	"errors"
	"fmt"

	"templatev27/internal/apperror"
	"templatev27/internal/business/domain"
	"templatev27/internal/datasources/records"
	"templatev27/pkg/logger"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// pgUniqueViolation нь Postgres-ийн unique_violation-ийн SQLSTATE код юм.
const pgUniqueViolation = "23505"

func (r *postgreUserRepository) Store(ctx context.Context, inDom *domain.User) (domain.User, error) {
	const (
		repositoryName = "users"
		funcName       = "Store"
		queryName      = "insertUser"
		fileName       = "users.store.go"
	)
	userRecord := records.FromUsersV1Domain(inDom)

	// INSERT ... RETURNING * — ингэснээр дуудагч хадгалагдсан мөрийг нэг
	// round-trip-д авна. id нь uuid_generate_v4() баганын өгөгдмөл утгаар
	// (SQL migration-уудаар бэлтгэгдсэн) сервер талд үүсгэгддэг тул бид
	// төрөлжсөн Create-ийн оронд GORM-ээр түүхий SQL гаргадаг — GORM-ийн
	// Create нь хоосон Id мөрийг бичихийг оролдох болно. Өмнө нь бид
	// INSERT хийгээд дараа нь GetByEmail хийдэг байсан; хэрэв GetByEmail
	// амжилтгүй болбол (сүлжээний саатал, replica lag) INSERT аль хэдийн
	// commit хийгдсэн байсан бөгөөд хэрэглэгч хариунд өнчирдөг байсан.
	var stored records.Users
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			INSERT INTO users(id, username, email, password, active, role_id, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, false, ?, ?)
			RETURNING id, username, email, password, active, role_id, created_at, updated_at, deleted_at, password_changed_at
		`, userRecord.Username, userRecord.Email, userRecord.Password, userRecord.RoleId, userRecord.CreatedAt).Scan(&stored).Error
	})
	if err != nil {
		// GORM-ийн Raw().Scan() зам нь TranslateError-г ажиллуулдаггүй тул
		// gorm.ErrDuplicatedKey хэзээ ч үүсэхгүй. Иймд pgx драйверын буцаасан
		// *pgconn.PgError-г шууд шалгаж, 23505 unique_violation-г 409 Conflict
		// болгон буулгана.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			logger.ErrorWithContext(ctx, "Failed to insert user: unique violation", logger.Fields{
				"repository": repositoryName,
				"method":     funcName,
				"query":      queryName,
				"file":       fileName,
				"error":      err.Error(),
				"table":      "users",
				"email":      userRecord.Email,
			})
			return domain.User{}, apperror.Conflict("username or email already exists")
		}
		logger.ErrorWithContext(ctx, "Failed to insert user into database", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"query":      queryName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "users",
		})
		return domain.User{}, err
	}

	if stored.Id == "" {
		// RETURNING нь амжилттай INSERT дээр хэзээ ч тэг мөр гаргадаггүй
		// боловч ирээдүйн schema өөрчлөлт хоосон struct-г чимээгүй буцааж
		// чадахааргүй байхын тулд ямар ч байсан шалга.
		err := fmt.Errorf("insert succeeded but RETURNING produced no row")
		logger.ErrorWithContext(ctx, "Insert returned no row", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"query":      queryName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "users",
		})
		return domain.User{}, err
	}
	return stored.ToV1Domain(), nil
}
