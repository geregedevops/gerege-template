// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewHealthHandler(db *gorm.DB, redisClient *redis.Client) HealthHandler {
	return HealthHandler{db: db, redisClient: redisClient}
}

func (h HealthHandler) Health(c fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "service is healthy",
	})
}

func (h HealthHandler) Ready(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	healthy := true

	// өгөгдлийн санг шалга — GORM handle-аас авсан үндсэн *sql.DB-ээр
	// дамжуулан ping хий.
	if sqlDB, err := h.db.DB(); err != nil {
		checks["database"] = "unreachable: " + err.Error()
		healthy = false
	} else if err := sqlDB.PingContext(ctx); err != nil {
		checks["database"] = "unreachable: " + err.Error()
		healthy = false
	} else {
		checks["database"] = "ok"
	}

	// redis-г шалга
	if h.redisClient != nil {
		if err := h.redisClient.Ping(ctx).Err(); err != nil {
			checks["redis"] = "unreachable: " + err.Error()
			healthy = false
		} else {
			checks["redis"] = "ok"
		}
	}

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	return c.Status(status).JSON(fiber.Map{
		"status": healthy,
		"checks": checks,
	})
}
