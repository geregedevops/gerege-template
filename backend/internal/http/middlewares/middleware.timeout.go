// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

// DefaultRequestTimeout нь нэг хүсэлтийн боловсруулалтын дээд хугацаа.
// Удаан гацсан handler / query нь холболтыг хэт удаан эзлэхээс сэргийлэх
// хамгаалалт (secure_system_guide §5.3, OWASP API4 Unrestricted Resource
// Consumption). Mailer зэрэг урт ажил нь хүсэлтийн замаас гадуур async
// ажилладаг тул энэ хязгаар тэдгээрт нөлөөлөхгүй.
const DefaultRequestTimeout = 30 * time.Second

// TimeoutMiddleware нь хүсэлтийн context дээр deadline тогтооно. Уг
// deadline нь handler-аас usecase → repository руу дамжиж, эцэст нь
// GORM-ийн WithContext(ctx) query-д хүрдэг тул хугацаа хэтэрсэн query
// автоматаар цуцлагдана. Энэ нь tracing / request-id middleware-ийн
// дараа байрлах ёстой — ингэснээр deadline-тай context нь тэдгээрийн
// тавьсан утгуудыг (trace_id, request_id) хадгална.
func TimeoutMiddleware(d time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), d)
		defer cancel()
		c.SetContext(ctx)
		return c.Next()
	}
}
