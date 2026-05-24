// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package records

import (
	"time"

	"gorm.io/gorm"
)

// Users нь users хүснэгтийн GORM model юм. Struct tag-ууд нь одоо
// байгаа snake_case schema руу буудаг; `gorm.DeletedAt` нь GORM-ийн
// soft-delete механизмыг deleted_at багана дээр холбодог тул өгөгдмөл
// query-үүд soft-delete хийгдсэн мөрүүдийг ил тод хасдаг
// (WHERE deleted_at IS NULL).
type Users struct {
	Id                string         `gorm:"column:id;primaryKey"`
	Username          string         `gorm:"column:username"`
	Email             string         `gorm:"column:email"`
	Password          string         `gorm:"column:password"`
	Active            bool           `gorm:"column:active"`
	RoleId            int            `gorm:"column:role_id"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
	UpdatedAt         *time.Time     `gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;index"`
	PasswordChangedAt *time.Time     `gorm:"column:password_changed_at"`
}

// TableName нь хүснэгтийн нэрийг тогтоодог тул GORM-ийн олон тоо болгогч
// биднийг гайхшруулж чадахгүй.
func (Users) TableName() string { return "users" }
