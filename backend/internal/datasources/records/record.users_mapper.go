// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package records

import (
	"time"

	"gorm.io/gorm"
	"templatev27/internal/business/domain"
)

func (u *Users) ToV1Domain() domain.User {
	// gorm.DeletedAt -> *time.Time: мөр үнэхээр soft-delete хийгдсэн үед
	// (Valid == true) л nil биш pointer-г гаргана.
	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		t := u.DeletedAt.Time
		deletedAt = &t
	}
	return domain.User{
		ID:                u.Id,
		Username:          u.Username,
		Email:             u.Email,
		Password:          u.Password,
		Active:            u.Active,
		RoleID:            u.RoleId,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
		DeletedAt:         deletedAt,
		PasswordChangedAt: u.PasswordChangedAt,
	}
}

func FromUsersV1Domain(u *domain.User) Users {
	// *time.Time -> gorm.DeletedAt: nil pointer нь тэг (хүчингүй)
	// gorm.DeletedAt руу буудаг бөгөөд GORM үүнийг "устгаагүй" гэж үздэг.
	var deletedAt gorm.DeletedAt
	if u.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *u.DeletedAt, Valid: true}
	}
	return Users{
		Id:                u.ID,
		Username:          u.Username,
		Email:             u.Email,
		Password:          u.Password,
		Active:            u.Active,
		RoleId:            u.RoleID,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
		DeletedAt:         deletedAt,
		PasswordChangedAt: u.PasswordChangedAt,
	}
}

func ToArrayOfUsersV1Domain(u *[]Users) []domain.User {
	var result []domain.User

	for _, val := range *u {
		result = append(result, val.ToV1Domain())
	}

	return result
}
