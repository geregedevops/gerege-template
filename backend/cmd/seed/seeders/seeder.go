// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package seeders

import (
	"errors"
	"time"

	"templatev27/internal/constants"
	"templatev27/internal/datasources/records"
	"templatev27/internal/datasources/rls"
	"templatev27/pkg/logger"

	"gorm.io/gorm"
)

type Seeder interface {
	UserSeeder(userData []records.Users) (err error)
}

type seeder struct {
	db *gorm.DB
}

func NewSeeder(db *gorm.DB) Seeder {
	return &seeder{db: db}
}

func (s *seeder) UserSeeder(userData []records.Users) (err error) {
	if len(userData) == 0 {
		return errors.New("users data is empty")
	}

	logger.Info("inserting users data...", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})

	// Бүхэл багцад нэг транзакц — хагас дутуу seed нь огт байхгүйгээс дор.
	// Id-г INSERT-ээс хассан тул Postgres түүнийг uuid_generate_v4() баганы
	// анхдагч утгаар (migration-ууд бэлдсэн) дүүргэнэ.
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// users хүснэгт дээр RLS асаалттай бөгөөд FORCE хийгдсэн тул seeder
		// нь хүснэгтийн эзэн (owner) байсан ч мөр оруулахын тулд "service"
		// үүргийг авах ёстой. set_config(..., true) нь SET LOCAL — зөвхөн
		// энэ транзакцид хүчинтэй.
		if cfgErr := tx.Exec(
			`SELECT set_config('app.user_role', ?, true)`, string(rls.RoleService),
		).Error; cfgErr != nil {
			return cfgErr
		}
		for i := range userData {
			userData[i].CreatedAt = time.Now().UTC()
			if createErr := tx.Omit("Id").Create(&userData[i]).Error; createErr != nil {
				return createErr
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	logger.Info("users data inserted successfully", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	return
}
