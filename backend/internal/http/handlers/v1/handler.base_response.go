// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package v1

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"templatev27/internal/apperror"
	"templatev27/internal/constants"
	"templatev27/pkg/logger"
	"templatev27/pkg/validators"
)

// RequestIDLocalsKey нь request-id middleware корреляцийн ID-г хадгалдаг
// Locals түлхүүр юм. Import cycle-аас зайлсхийхийн тулд (middlewares-ээс
// import хийхийн оронд) энд давхардуулсан: middlewares нь энэ package-г
// import хийдэг.
const RequestIDLocalsKey = "X-Request-ID"

type BaseResponse struct {
	Status    bool        `json:"status"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// requestID нь request-id middleware-ийн бөглөсөн X-Request-ID-г
// Fiber-ийн хүсэлтийн Locals-аас уншина.
func requestID(c fiber.Ctx) string {
	if v, ok := c.Locals(RequestIDLocalsKey).(string); ok {
		return v
	}
	return ""
}

// NewSuccessResponse нь амжилтын дугтуй бичиж, JSON encoder-аас гарсан
// алдааг (хэрэв байгаа бол) буцаадаг тул handler-ууд үүнийг `return`
// хийж чадна.
func NewSuccessResponse(c fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(BaseResponse{
		Status:    true,
		Message:   message,
		Data:      data,
		RequestID: requestID(c),
	})
}

// NewErrorResponse нь алдааны дугтуй бичнэ.
func NewErrorResponse(c fiber.Ctx, statusCode int, err string) error {
	return c.Status(statusCode).JSON(BaseResponse{
		Status:    false,
		Message:   err,
		RequestID: requestID(c),
	})
}

// NewAbortResponse нь нэгдсэн BaseResponse дугтуйтай 401-г үзүүлнэ.
// Fiber-т middleware нь Next дуудахын оронд хариуг буцааж богино
// холбодог тул энэ нь зүгээр л encoder-ийн алдааг дуудагчид түгээхээр
// буцаана.
func NewAbortResponse(c fiber.Ctx, message string) error {
	return c.Status(http.StatusUnauthorized).JSON(BaseResponse{
		Status:    false,
		Message:   message,
		RequestID: requestID(c),
	})
}

// mapDomainErrorToHTTP нь домэйн алдааг HTTP статус код руу хувиргана.
func mapDomainErrorToHTTP(err error) int {
	var domErr *apperror.DomainError
	if errors.As(err, &domErr) {
		switch domErr.Type {
		case apperror.ErrTypeNotFound:
			return http.StatusNotFound
		case apperror.ErrTypeUnauthorized:
			return http.StatusUnauthorized
		case apperror.ErrTypeForbidden:
			return http.StatusForbidden
		case apperror.ErrTypeConflict:
			return http.StatusConflict
		case apperror.ErrTypeBadRequest:
			return http.StatusBadRequest
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// RespondWithError нь цэвэрлэсэн алдааны хариу гаргаж, дотоод (5xx)
// бүх алдааны үндсэн шалтгааныг бүртгэдэг. "Дэлгэрэнгүй мэдээллийг
// бүртгэж, хэрэглэгчид цэвэр мессеж харуулах" дүрмийг төвлөрүүлдэг тул
// аль ч handler ороомог болсон сангийн алдааг body руу санамсаргүй
// түлхэхгүй.
func RespondWithError(c fiber.Ctx, err error) error {
	// Баталгаажуулалтын алдаанууд өөрсдийн бүтэцлэгдсэн хариуг авдаг —
	// клиентүүд зөвхөн нэгтгэсэн товчоог биш, *аль талбар* буруу байгааг
	// мэдэх хэрэгтэй. `data` дотор талбар бүрийн дэлгэрэнгүйтэйгээр 422
	// болгон үзүүл.
	var ve *validators.ValidationErrors
	if errors.As(err, &ve) {
		return c.Status(http.StatusUnprocessableEntity).JSON(BaseResponse{
			Status:    false,
			Message:   "validation failed",
			Data:      map[string]any{"errors": ve.Errors},
			RequestID: requestID(c),
		})
	}

	status := mapDomainErrorToHTTP(err)
	message := err.Error()

	if status >= http.StatusInternalServerError {
		// Жинхэнэ шалтгааныг (эсвэл шалтгаан хавсаргаагүй бол алдаа
		// өөрийг нь) бүртгэж, ерөнхий мессеж буцаа. Үүнгүйгээр клиентүүд
		// "hash password: bcrypt: …" гэх мэт зүйлсийг хардаг.
		fields := logger.Fields{
			constants.LoggerCategory: constants.LoggerCategoryHTTP,
			"path":                   c.Path(),
		}
		var domErr *apperror.DomainError
		if errors.As(err, &domErr) && domErr.Cause != nil {
			fields["cause"] = domErr.Cause.Error()
		} else {
			fields["cause"] = err.Error()
		}
		if rid := requestID(c); rid != "" {
			fields["request_id"] = rid
		}
		logger.Error("internal error while handling request", fields)
		message = "internal server error"
	}

	return NewErrorResponse(c, status, message)
}
