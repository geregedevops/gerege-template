// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"templatev27/pkg/logger"
)

const RequestIDHeader = "X-Request-ID"

// RequestIDMiddleware нь ирж буй X-Request-ID-г хүлээж авна (эсвэл
// байхгүй бол UUID үүсгэдэг), хариунд буцаан тусгаж, Fiber Locals-д
// хадгалдаг (ингэснээр хариуны дугтуй үүнийг буцааж унших боломжтой), мөн
// хоёр корреляцийн ID-г хүсэлтийн context руу гүүрлэдэг тул
// logger.*WithContext нь тэдгээрийг log мөр бүрд гаргадаг:
//
//   - request_id: гадаад клиентэд харагдах ID. Үйлчилгээнүүдийн хооронд
//     ч клиентэд эхнээс эцэс хүртэл ижил хэвээр үлддэг.
//   - traceId: OTel-ийн үүсгэсэн W3C trace ID. tracing backend
//     (Jaeger / Tempo / г.м.) дахь span-уудтай log-уудыг холбоход
//     ашиглагддаг.
//
// Үүнийг tracing middleware-ийн ДАРАА суулга — ингэснээр бид trace ID-г
// гаргаж авахаар хүрэх үед OTel span context аль хэдийн тогтоогдсон байна.
func RequestIDMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := c.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Locals(RequestIDHeader, requestID)
		c.Set(RequestIDHeader, requestID)

		ctx := context.WithValue(c.Context(), logger.RequestIDKey, requestID)
		if span := trace.SpanFromContext(ctx); span.SpanContext().HasTraceID() {
			ctx = context.WithValue(ctx, logger.TraceIDKey, span.SpanContext().TraceID().String())
		}
		c.SetContext(ctx)

		return c.Next()
	}
}
