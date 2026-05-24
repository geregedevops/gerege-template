// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	repointerface "templatev27/internal/datasources/repositories/interface"

	"gorm.io/gorm"
)

// postgreUserRepository нь GORM handle-г агуулна. Interface-ийн method
// бүр өөрийн файлд (users.store.go, users.get_by_email.go, ...)
// байрладаг тул нэг query-д хүрэх PR diff-үүд нарийн тодорхой хэвээр
// үлддэг.
type postgreUserRepository struct {
	conn *gorm.DB
}

func NewUserRepository(conn *gorm.DB) repointerface.UserRepository {
	return &postgreUserRepository{conn: conn}
}
