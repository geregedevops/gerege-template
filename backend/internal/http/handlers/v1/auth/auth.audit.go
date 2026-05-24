// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth

import (
	"github.com/gofiber/fiber/v3"
	"templatev27/pkg/audit"
	"templatev27/pkg/logger"
)

// auditFromFiber нь audit Event-ийн HTTP-context хэсгийг (IP,
// user-agent, request_id, trace_id) бүтээдэг тул дуудах газрууд зөвхөн
// event-д хамаарах талбаруудыг бөглөхөд хангалттай. Корреляцийн ID-ууд
// нь audit бичлэгүүдийг бүтэцлэгдсэн аппликейшний log-ууд руу болон
// (trace_id-ээр дамжуулан) tracing backend дахь span-ууд руу буцаан
// холбох боломжийг олгоно.
//
// Fiber port-ийн тэмдэглэл: request-id middleware нь корреляцийн ID-г
// Fiber-ийн Locals дотор "X-Request-ID" дор хадгалдаг; trace ID нь
// tracing middleware c.SetContext-ээр тогтоосон span-тай хүсэлтийн
// context-ээс татагддаг.
func auditFromFiber(c fiber.Ctx) audit.Event {
	requestID := ""
	if v, ok := c.Locals("X-Request-ID").(string); ok {
		requestID = v
	}
	return audit.Event{
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent"),
		RequestID: requestID,
		TraceID:   logger.GetTraceIDFromContext(c.Context()),
	}
}
