// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"github.com/gofiber/fiber/v3"
	"templatev27/internal/business/usecases/users"
	usershandler "templatev27/internal/http/handlers/v1/users"
)

// usersRoute нь /users/* бүлгийг холбоно — хэрэглэгчийн өөрийнх нь
// профайл / өгөгдөлд хамаарах endpoint-ууд. Auth урсгалууд нь
// route.auth.go-д байрладаг.
type usersRoute struct {
	handler        usershandler.Handler
	router         fiber.Router
	authMiddleware fiber.Handler
}

// NewUsersRoute нь route модулийг бүтээдэг. Auth middleware-г (энд
// бүтээхийн оронд) дамжуулдаг тул ижил JWT баталгаажуулдаг middleware нь
// хамгаалагдсан route бүлэг бүрд хуваалцагддаг.
func NewUsersRoute(router fiber.Router, usersUC users.Usecase, authMiddleware fiber.Handler) *usersRoute {
	return &usersRoute{
		handler:        usershandler.NewHandler(usersUC),
		router:         router,
		authMiddleware: authMiddleware,
	}
}

// Routes нь /users бүлэг болон түүний endpoint-уудыг суулгана.
func (r *usersRoute) Routes() {
	v1 := r.router.Group("/v1")
	usersGrp := v1.Group("/users")
	usersGrp.Use(r.authMiddleware)
	usersGrp.Get("/me", r.handler.GetUserData)
}
