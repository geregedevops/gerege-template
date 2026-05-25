// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"
	"fmt"

	repointerface "templatev27/internal/datasources/repositories/interface"
	"templatev27/internal/datasources/rls"

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

// withRLS нь нэг query-г транзакцид боож, тухайн транзакцид зориулж
// Postgres-ийн Row-Level Security session хувьсагчдыг (app.user_id,
// app.user_role) тогтооно. Утгуудыг context-оос (rls.FromContext)
// уншиж авдаг.
//
// Яагаад транзакц шаардлагатай вэ: set_config-ийн гурав дахь аргумент
// (is_local) нь `true` — энэ нь `SET LOCAL`-той дүйцэх бөгөөд утгыг
// зөвхөн ИДЭВХТЭЙ транзакцийн туршид хадгална. GORM нь холболтын pool
// ашигладаг тул жирийн `SET` нь нэг хүсэлтийн identity-г pool дахь
// холболтод үлдээж, дараагийн хамааралгүй хүсэлт рүү "алдагдуулах"
// эрсдэлтэй; SET LOCAL транзакц commit/rollback хийгдмэгц автоматаар
// арилдаг тул энэ алдагдлаас сэргийлнэ.
//
// context-д Identity байхгүй бол UserID/Role нь хоосон болж, RLS
// бодлогууд бүх мөрийг хаана — аюулгүй өгөгдмөл (fail-closed).
func (r *postgreUserRepository) withRLS(ctx context.Context, fn func(tx *gorm.DB) error) error {
	id, _ := rls.FromContext(ctx)
	return r.conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Хоёр GUC-г нэг round-trip-д тогтооно. `?` placeholder-ууд нь
		// утгуудыг параметр болгон холбодог тул SQL injection боломжгүй.
		if err := tx.Exec(
			`SELECT set_config('app.user_id', ?, true), set_config('app.user_role', ?, true)`,
			id.UserID, string(id.Role),
		).Error; err != nil {
			return fmt.Errorf("set rls session context: %w", err)
		}
		return fn(tx)
	})
}
