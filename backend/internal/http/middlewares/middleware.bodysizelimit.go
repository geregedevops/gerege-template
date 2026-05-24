// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// Нийтлэг body-хэмжээний дээд хязгаарууд. Route-ууд хүлээж авдаг
// payload-доо тохирох хамгийн чанга хязгаарыг хэрэглэдэг. Глобал
// өгөгдмөл (DefaultBodyMaxBytes) нь өөрийн хязгаар тогтоогоогүй аль ч
// route-ийн сүүлчийн хамгаалалтын шугам юм.
const (
	// DefaultBodyMaxBytes нь бүх зүйлийг барих глобал дээд хязгаар — 1 MiB.
	DefaultBodyMaxBytes int64 = 1 << 20

	// AuthBodyMaxBytes нь register / login / refresh / logout payload-уудыг
	// хамардаг. Эдгээрийн аль нь ч хэдэн зуун байтаас илүү JSON авч
	// явдаггүй; 4 KiB-д хязгаарлах нь нэрээ нууцалсан урсгал хүлээн авдаг
	// цорын ганц route-уудын эсрэг хэт том payload-ийн дайралтыг хууль
	// ёсны ямар ч хүсэлтэд нөлөөлөхгүйгээр хаадаг.
	AuthBodyMaxBytes int64 = 4 << 10
)

// BodySizeLimitMiddleware нь body нь maxBytes-ээс хэтэрсэн аль ч
// хүсэлтийг 413 Payload Too Large-ээр татгалздаг. Fiber (fasthttp) нь
// middleware ажиллах үед body-г аль хэдийн санах ойд уншсан байдаг тул
// энэ нь net/http шиг reader-г ороохын оронд Content-Length / бодит
// уртыг шалгадаг. fiber.Config-д тогтоосон framework-түвшний BodyLimit
// нь үнэхээр асар том upload-ийн эсрэг жинхэнэ эхний хамгаалалтын шугам;
// энэ нь түүн дээр route бүрийн чангалалт өгдөг.
func BodySizeLimitMiddleware(maxBytes int64) fiber.Handler {
	return func(c fiber.Ctx) error {
		if int64(len(c.Body())) > maxBytes {
			return c.Status(http.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"status":  false,
				"message": "request entity too large",
			})
		}
		return c.Next()
	}
}
