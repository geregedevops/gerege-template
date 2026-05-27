// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"templatev27/internal/business/ports"
	"templatev27/internal/business/usecases/auth"
	"templatev27/internal/constants"
	"templatev27/internal/datasources/rls"
	V1Handler "templatev27/internal/http/handlers/v1"
	"templatev27/pkg/jwt"
	"templatev27/pkg/logger"
)

type AuthMiddleware struct {
	jwtService jwt.JWTService
	redisCache ports.Cache
	isAdmin    bool
}

// NewAuthMiddleware нь Bearer токеныг баталгаажуулж, нууц үг солих
// (rotation) хязгаарыг хүндэтгэж, задлан шинжилсэн claim-уудыг хүсэлтийн
// Locals-д хадгалдаг Fiber handler буцаана. Хариуг буцааж 401-ээр богино
// холбодог (гинжийг таслах Fiber-ийн арга барил).
func NewAuthMiddleware(jwtService jwt.JWTService, redisCache ports.Cache, isAdmin bool) fiber.Handler {
	return (&AuthMiddleware{
		jwtService: jwtService,
		redisCache: redisCache,
		isAdmin:    isAdmin,
	}).Handle
}

func (m *AuthMiddleware) Handle(c fiber.Ctx) error {
	const (
		middlewareName = "AuthMiddleware"
		fileName       = "middleware.auth.go"
	)
	logCtx := c.Context()
	path := c.Path()

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		logger.WarnWithContext(logCtx, "Auth: missing Authorization header", logger.Fields{
			"middleware": middlewareName,
			"file":       fileName,
			"step":       "read_header",
			"path":       path,
		})
		return V1Handler.NewAbortResponse(c, "missing authorization header")
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		logger.WarnWithContext(logCtx, "Auth: invalid Authorization header format", logger.Fields{
			"middleware": middlewareName,
			"file":       fileName,
			"step":       "parse_header",
			"path":       path,
		})
		return V1Handler.NewAbortResponse(c, "invalid header format")
	}

	if headerParts[0] != "Bearer" {
		logger.WarnWithContext(logCtx, "Auth: non-Bearer scheme", logger.Fields{
			"middleware": middlewareName,
			"file":       fileName,
			"step":       "scheme_check",
			"path":       path,
			"scheme":     headerParts[0],
		})
		return V1Handler.NewAbortResponse(c, "token must content bearer")
	}

	user, err := m.jwtService.ParseToken(headerParts[1])
	if err != nil {
		logger.WarnWithContext(logCtx, "Auth: token parse failed", logger.Fields{
			"middleware": middlewareName,
			"file":       fileName,
			"step":       "parse_token",
			"path":       path,
			"error":      err.Error(),
		})
		return V1Handler.NewAbortResponse(c, "invalid token")
	}

	// Хэрэглэгчийн хамгийн сүүлийн нууц үг солихоос (rotation) өмнө
	// олгогдсон access токенуудыг татгалз. Хязгаарыг ChangePassword
	// Redis руу нийтэлдэг; байхгүй (Redis miss) нь сүүлийн үед солилт
	// хийгдээгүй гэсэн үг тул токен нэвтэрнэ. Redis алдаа нь нээлттэй
	// бүтэлгүйтдэг — токен аль хэдийн гарын үсэг + хугацааг давсан бөгөөд
	// бид Redis-ийн түр зуурын саатлаас болж бүх хүнийг түгжихийг хүсэхгүй.
	if m.redisCache != nil && user.IssuedAt != nil {
		if cutoffStr, getErr := m.redisCache.Get(logCtx, auth.TokenCutoffKey(user.UserID)); getErr == nil && cutoffStr != "" {
			// JWT IssuedAt нь секунд хүртэл бутархайгүй болгогддог тул нууц
			// үг солихтой яг нэг секундэд олгогдсон токеныг бас татгалзахын
			// тулд <= ашиглана (хил дээрх секундын цоорхойг хаана).
			if cutoff, parseErr := strconv.ParseInt(cutoffStr, 10, 64); parseErr == nil && user.IssuedAt.Unix() <= cutoff {
				logger.WarnWithContext(logCtx, "Auth: token revoked by password rotation", logger.Fields{
					"middleware": middlewareName,
					"file":       fileName,
					"step":       "check_pwd_cutoff",
					"path":       path,
					"user_id":    user.UserID,
					"issued_at":  user.IssuedAt.Unix(),
					"cutoff":     cutoff,
				})
				return V1Handler.NewAbortResponse(c, "token has been revoked")
			}
		}
	}

	// Admin шаардлагатай route-д зөвхөн admin зөвшөөрнө; admin биш route-уудад
	// (m.isAdmin == false) хэн ч (хэрэглэгч ч, admin ч) дамжина. Дэлгэрэнгүй: admin
	// нь нийтийн / хэрэглэгчийн endpoint-уудыг дандаа дуудаж чаддаг тул хязгаарлах
	// нэг л чиглэлийг (m.isAdmin && !user.IsAdmin) шалгана.
	if m.isAdmin && !user.IsAdmin {
		logger.WarnWithContext(logCtx, "Auth: insufficient privilege", logger.Fields{
			"middleware":     middlewareName,
			"file":           fileName,
			"step":           "privilege_check",
			"path":           path,
			"user_id":        user.UserID,
			"required_admin": m.isAdmin,
			"user_is_admin":  user.IsAdmin,
		})
		return V1Handler.NewAbortResponse(c, "you don't have access for this action")
	}

	// RLS: доош урсгах repository query-үүд хэрэглэгчийн өөрийнх нь мөрөөр
	// (admin бол бүх мөрөөр) хязгаарлагдахаар request context-д identity
	// суулгана. SetContext-ийн дараа handler-ийн c.Context() энэ баяжуулсан
	// context-г буцаадаг тул identity нь usecase → repository хүртэл дамжина.
	role := rls.RoleUser
	if user.IsAdmin {
		role = rls.RoleAdmin
	}
	c.SetContext(rls.With(c.Context(), rls.Identity{UserID: user.UserID, Role: role}))

	c.Locals(constants.CtxAuthenticatedUserKey, user)
	return c.Next()
}
