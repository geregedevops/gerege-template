// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

// MetricsMiddleware нь route тус бүрийн хүсэлтийн тоо + үргэлжлэх
// хугацааг бүртгэдэг. Өндөр-кардиналтай path параметрүүд метрик цувааг
// тэсрүүлэхгүйн тулд таарсан route загвар (c.Route().Path)-г path
// шошго болгон ашигладаг.
func MetricsMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		path := c.Route().Path
		if path == "" {
			path = "unknown"
		}

		httpRequestsTotal.WithLabelValues(c.Method(), path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Method(), path).Observe(duration)

		return err
	}
}
