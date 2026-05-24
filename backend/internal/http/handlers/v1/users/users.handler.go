// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package users нь /users/* HTTP endpoint-уудыг үйлчилдэг —
// баталгаажуулагдсан хэрэглэгчийн өөрийнх нь профайл / өгөгдөлд
// хамаарах бүх зүйл. Auth урсгалууд нь ах дүү package болох
// internal/http/handlers/v1/auth-д байрладаг.
package users

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"templatev27/internal/business/usecases/users"
	httpauth "templatev27/internal/http/auth"
	"templatev27/internal/http/datatransfers/responses"
	v1 "templatev27/internal/http/handlers/v1"
	"templatev27/pkg/logger"
)

// Handler нь user-домэйн endpoint-уудыг үйлчилдэг. Энэ нь зөвхөн
// users.Usecase руу дууддаг — хэзээ ч repository эсвэл auth context
// руу шууд дууддаггүй.
type Handler struct {
	usecase users.Usecase
}

func NewHandler(usecase users.Usecase) Handler {
	return Handler{usecase: usecase}
}

// GetUserData godoc
// @Summary      Одоогийн хэрэглэгчийн профайлыг буцаах
// @Description  Authorization header дахь JWT-ээс баталгаажуулагдсан хэрэглэгчийг уншиж, тохирох бичлэгийг буцаана (эхлээд in-memory кэш, олдоогүй үед Postgres).
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=responses.UserResponse}  "User profile"
// @Failure      401  {object}  v1.BaseResponse  "Missing or invalid token"
// @Failure      404  {object}  v1.BaseResponse  "User no longer exists"
// @Router       /users/me [get]
func (h Handler) GetUserData(c fiber.Ctx) error {
	const (
		controllerName = "users"
		funcName       = "GetUserData"
		fileName       = "users.handler.go"
	)
	ctx := c.Context()
	user, err := httpauth.CurrentUserFromContext(c)
	if err != nil {
		logger.WarnWithContext(ctx, "GetUserData: not authenticated", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
		})
		return v1.NewErrorResponse(c, http.StatusUnauthorized, err.Error())
	}

	resp, err := h.usecase.GetByEmail(ctx, users.GetByEmailRequest{Email: user.Email})
	if err != nil {
		logger.ErrorWithContext(ctx, "GetUserData failed in controller", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"email":      user.Email,
		})
		return v1.RespondWithError(c, err)
	}

	return v1.NewSuccessResponse(c, http.StatusOK, "user data fetched successfully", map[string]interface{}{
		"user": responses.FromV1Domain(resp.User),
	})
}
