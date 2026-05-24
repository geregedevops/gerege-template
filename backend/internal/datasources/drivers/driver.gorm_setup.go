// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package drivers

import (
	"time"

	"templatev27/internal/config"
	"templatev27/internal/constants"

	"gorm.io/gorm"
)

// SetupGORMPostgres нь config.AppConfig-оос DB_POSTGRE_* түлхүүрүүдийг
// уншиж, Postgres руу чиглэсэн *gorm.DB-г бүтээж ping хийдэг. Хоёр
// хэсэгтэй нэр нь хоёр давхаргыг хоёуланг нь илэрхийлнэ: ORM (GORM)
// болон engine (Postgres). Ирээдүйд GORM-аар дамжих MySQL холболт нь
// үүний хажууд SetupGORMMySQL нэрээр нэмэгдэнэ. Энэ нь config утгуудыг
// холбогдсон драйвер болгон хувиргадаг композици тул drivers package-д
// байрладаг — драйверийн хэрэгжилттэй ижил давхарга.
func SetupGORMPostgres() (*gorm.DB, error) {
	var dsn string
	switch config.AppConfig.Environment {
	case constants.EnvironmentDevelopment:
		dsn = config.AppConfig.DBPostgreDsn
	case constants.EnvironmentProduction:
		dsn = config.AppConfig.DBPostgreURL
	}

	cfg := GORMConfig{
		DataSourceName: dsn,
		MaxOpenConns:   config.AppConfig.DBMaxOpenConns,
		MaxIdleConns:   config.AppConfig.DBMaxIdleConns,
		MaxLifetime:    time.Duration(config.AppConfig.DBConnMaxLifeMins) * time.Minute,
		Debug:          config.AppConfig.Debug,
	}
	return cfg.InitializeGORMDatabase()
}
