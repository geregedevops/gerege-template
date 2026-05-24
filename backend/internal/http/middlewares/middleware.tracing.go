// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"github.com/gofiber/fiber/v3"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"templatev27/pkg/observability"
)

// TracingMiddleware нь Fiber v3-д зориулсан гар хийцийн OpenTelemetry
// middleware юм.
//
// Үндэслэл: otelgin (анхны boilerplate-д ашигласан Gin хэмжилт) нь
// арчилгаатай Fiber v3 хувилбаргүй бөгөөд community otelfiber package
// нь v3 fiber.Ctx interface биш, Fiber v2-ийн *fiber.Ctx-г зорьдог.
// Contrib хамаарлыг тогтоохын оронд бид хүсэлт тус бүрд глобал
// tracer-ээс (pkg/observability.SetupTracing-ээр тохируулагдсан) span
// эхлүүлдэг. Tracing идэвхгүй үед глобал provider нь OTel-ийн no-op
// байх тул энэ нь бараг ямар ч зардалгүй.
//
// Үүнийг ЭХЭНД суулга — ингэснээр RequestIDMiddleware нь span context
// (trace_id)-г logger context руу гүүрлэхээр хүрэхээс өмнө тогтоогдох
// бөгөөд stack-ийн доор гарсан span-ууд (DB, Redis) энэ серверийн
// span-ийн child болж автоматаар үүснэ.
func TracingMiddleware(serviceName string) fiber.Handler {
	tracer := observability.Tracer()
	return func(c fiber.Ctx) error {
		ctx, span := tracer.Start(
			c.Context(),
			c.Method()+" "+c.Path(),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Method()),
				semconv.URLPath(c.Path()),
				semconv.ServerAddress(c.Hostname()),
			),
		)
		defer span.End()

		// span-тай context-г доош түгээ — ингэснээр DB / Redis span-ууд
		// түүн доор үүрлэх бөгөөд RequestIDMiddleware нь trace_id-г
		// гаргаж авч чадна.
		c.SetContext(ctx)

		err := c.Next()

		span.SetAttributes(semconv.HTTPResponseStatusCode(c.Response().StatusCode()))
		return err
	}
}
