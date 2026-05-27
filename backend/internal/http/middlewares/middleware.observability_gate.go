// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"crypto/subtle"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// ObservabilityGate нь /metrics ба /swagger/doc.json гэх мэт операторын
// endpoint-уудыг хамгаалах нимгэн middleware юм.
//
// Стратеги:
//   - production биш үед: үргэлж зөвшөөрнө (dev UX-ийг хадгална).
//   - production-д token хоосон үед: 404 буцаана. endpoint бүхэлдээ
//     байхгүй мэт харагдах нь reconnaissance-ыг хүндрүүлнэ.
//   - production-д token тохируулсан үед: "Authorization: Bearer <token>"
//     яг тааравал зөвшөөрнө; өөр бол 404 (401 биш — token шаардлагатай
//     гэдгийг ил гаргахгүй).
//
// Token харьцуулалт нь crypto/subtle.ConstantTimeCompare ашиглан timing
// oracle-ыг хаана. Bearer prefix-ийг case-insensitive нөхдөг.
func ObservabilityGate(isProduction bool, token string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !isProduction {
			return c.Next()
		}
		if token == "" {
			return fiber.ErrNotFound
		}
		header := c.Get(fiber.HeaderAuthorization)
		const prefix = "bearer "
		if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
			return fiber.ErrNotFound
		}
		provided := strings.TrimSpace(header[len(prefix):])
		if subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			return fiber.ErrNotFound
		}
		return c.Next()
	}
}
