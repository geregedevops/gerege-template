// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

// access-log өнгөний кодууд (xterm SGR background).
const (
	accessLogRed    = "41"
	accessLogYellow = "43"
	accessLogGreen  = "42"
)

// requestIDContextKey нь RequestIDHeader-г тусгана. Тодорхой байх үүднээс
// дотооддоо хадгалсан.
const requestIDContextKey = "X-Request-ID"

// AccessLogMiddleware нь хүсэлт тус бүрд нэг мөр access log үзүүлнэ.
// Статус код өнгөтэй болгогдсон тул энгийн `tail -f` session-д 5xx / 4xx
// тодорч харагдана. Энэ нь Gin-ийн LoggerWithFormatter(AccessLogFormatter)-ийн
// Fiber-д төрөлх орлуулагч юм.
func AccessLogMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		var color string
		switch {
		case status >= 500:
			color = accessLogRed
		case status >= 400:
			color = accessLogYellow
		default:
			color = accessLogGreen
		}

		requestID := "-"
		if v, ok := c.Locals(requestIDContextKey).(string); ok && v != "" {
			requestID = v
		}

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}

		fmt.Printf("[LOGGING HTTP] [%s] req=%s \033[%sm %d \033[0m %s %s %s %s %s %s\n",
			start.Format("2006-01-02 15:04:05"),
			requestID,
			color,
			status,
			c.Method(),
			c.Path(),
			latency,
			c.IP(),
			errMsg,
			c.Get("User-Agent"),
		)

		return err
	}
}
