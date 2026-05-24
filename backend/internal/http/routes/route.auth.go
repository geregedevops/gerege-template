// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"github.com/gofiber/fiber/v3"
	"templatev27/internal/business/usecases/auth"
	authhandler "templatev27/internal/http/handlers/v1/auth"
	"templatev27/internal/http/middlewares"
)

// authRoute нь /auth/* бүлгийг холбоно — register / login / OTP /
// refresh / logout / forgot / reset. Нэрээ нууцалсан endpoint-ууд нь
// rate limiter болон чанга body хязгаар авдаг; /password/change нь
// нэмэлтээр JWT шаарддаг.
type authRoute struct {
	handler        authhandler.Handler
	router         fiber.Router
	rateLimiter    *middlewares.RateLimiter
	authMiddleware fiber.Handler
}

// NewAuthRoute нь route модулийг бүтээдэг. Rate limiter-г дуудагч
// эзэмшдэг тул түүний cleanup goroutine-г graceful shutdown (эвсэг
// унтраалт) үед Stop() хийж болно; auth middleware нь users route-той
// хуваалцагддаг тул ChangePassword нь ижил JWT баталгаажуулалтыг дахин
// ашиглаж чадна.
func NewAuthRoute(router fiber.Router, authUC auth.Usecase, authMiddleware fiber.Handler, rateLimiter *middlewares.RateLimiter) *authRoute {
	return &authRoute{
		handler:        authhandler.NewHandler(authUC),
		router:         router,
		rateLimiter:    rateLimiter,
		authMiddleware: authMiddleware,
	}
}

// Routes нь /auth бүлэг болон түүний endpoint-уудыг суулгана.
func (r *authRoute) Routes() {
	v1 := r.router.Group("/v1")
	authGrp := v1.Group("/auth")
	// Auth payload-ууд жижиг JSON хэсгүүд — 4 KiB-д хязгаарлах нь нэрээ
	// нууцалсан урсгал хүлээн авдаг цорын ганц route-уудын эсрэг хэт том
	// payload-ийн дайралтыг хууль ёсны ямар ч хүсэлтэд нөлөөлөхгүйгээр
	// хаадаг.
	authGrp.Use(r.rateLimiter.Middleware())
	authGrp.Use(middlewares.BodySizeLimitMiddleware(middlewares.AuthBodyMaxBytes))
	authGrp.Post("/register", r.handler.Register)
	authGrp.Post("/login", r.handler.Login)
	authGrp.Post("/send-otp", r.handler.SendOTP)
	authGrp.Post("/verify-otp", r.handler.VerifyOTP)
	authGrp.Post("/refresh", r.handler.Refresh)
	authGrp.Post("/logout", r.handler.Logout)
	authGrp.Post("/password/forgot", r.handler.ForgotPassword)
	authGrp.Post("/password/reset", r.handler.ResetPassword)
	// /password/change нь JWT шаарддаг — бусадтай ижил rate limiter /
	// body хязгаарын дээр auth middleware авдаг.
	authGrp.Put("/password/change", r.authMiddleware, r.handler.ChangePassword)
}
