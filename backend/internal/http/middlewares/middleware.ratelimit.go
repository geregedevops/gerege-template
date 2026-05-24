// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*ipLimiter
	rate     rate.Limit
	burst    int
	stop     chan struct{}
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*ipLimiter),
		rate:     r,
		burst:    burst,
		stop:     make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &ipLimiter{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanup нь Stop дуудагдах хүртэл 3 минут тутамд хуучирсан бичлэгүүдийг
// устгана.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 5*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stop:
			return
		}
	}
}

// Stop нь cleanup goroutine-г зогсооно.
func (rl *RateLimiter) Stop() {
	close(rl.stop)
}

// rateLimitedResponse нь v1.BaseResponse-ийн JSON хэлбэрийг тусгана.
// middleware нь handlers package-г import cycle-гүйгээр import хийж
// чадахгүй тул дугтуйг энд давхардуулсан — талбарын tag-уудыг ижил
// байлга.
type rateLimitedResponse struct {
	Status    bool   `json:"status"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// Middleware нь IP бүрийн token-bucket Fiber handler буцаана. Bucket
// хоосон үед 429-ээр (хариуг буцааж) богино холбоно.
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		ip := c.IP()
		limiter := rl.getLimiter(ip)

		writeRateLimitHeaders(c, rl.burst, limiter)

		if !limiter.Allow() {
			retryAfter := retryAfterSeconds(limiter)
			c.Set("Retry-After", strconv.Itoa(retryAfter))
			rid, _ := c.Locals(RequestIDHeader).(string)
			return c.Status(http.StatusTooManyRequests).JSON(rateLimitedResponse{
				Status:    false,
				Message:   "too many requests, please try again later",
				RequestID: rid,
			})
		}

		return c.Next()
	}
}

// writeRateLimitHeaders нь хязгаар, одоогийн үлдсэн токенууд болон
// bucket дахин дүүрэх unix timestamp-г зарладаг. Клиентүүд эдгээрийг
// эхлээд 429 шатаалгүйгээр буцахад ашигладаг.
func writeRateLimitHeaders(c fiber.Ctx, burst int, limiter *rate.Limiter) {
	tokens := limiter.Tokens()
	remaining := max(int(math.Floor(tokens)), 0)
	resetSeconds := 0
	if r := float64(limiter.Limit()); r > 0 {
		// Одоогийн түвшнээс bucket дахин дүүртэл хэдэн секунд үлдсэн.
		missing := float64(burst) - tokens
		if missing > 0 {
			resetSeconds = int(math.Ceil(missing / r))
		}
	}
	c.Set("X-RateLimit-Limit", strconv.Itoa(burst))
	c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(resetSeconds)*time.Second).Unix(), 10))
}

// retryAfterSeconds нь нэг шинэ токен бэлэн болохоос өмнөх хүлээх
// хугацааг тооцоолно. RFC 7231 Retry-After нь зөвхөн бүхэл секунд
// хүлээн авдаг тул дараагийн бүхэл секунд хүртэл дээш дугуйрсан.
func retryAfterSeconds(limiter *rate.Limiter) int {
	r := float64(limiter.Limit())
	if r <= 0 {
		return 1
	}
	deficit := 1.0 - limiter.Tokens()
	if deficit <= 0 {
		return 0
	}
	return int(math.Ceil(deficit / r))
}
