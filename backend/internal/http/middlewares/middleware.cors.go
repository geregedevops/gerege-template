// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"templatev27/internal/config"
)

// CORSMiddleware нь тохируулсан зөвшөөрөгдсөн origin-уудын жагсаалтаас
// Fiber CORS handler-г бүтээдэг. Цорын ганц origin нь wildcard "*"
// байх үед credentials идэвхгүй болдог (спецификаци нь credentials +
// wildcard-г хориглодог); тодорхой allow-list-д credentials идэвхждэг.
func CORSMiddleware() fiber.Handler {
	origins := config.AppConfig.AllowedOriginsList()
	cfg := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Cache-Control", "X-Requested-With", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           int((12 * 60 * 60)), // 12 цаг, секундээр
	}
	if !(len(origins) == 1 && origins[0] == "*") {
		cfg.AllowCredentials = true
	}
	return cors.New(cfg)
}
