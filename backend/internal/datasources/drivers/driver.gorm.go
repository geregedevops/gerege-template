// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package drivers

import (
	"fmt"
	"time"

	"templatev27/internal/constants"
	"templatev27/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

// GORMConfig нь өгөгдлийн сангийн instance-ийн тохиргоог хадгална.
type GORMConfig struct {
	DataSourceName string
	MaxOpenConns   int
	MaxIdleConns   int
	MaxLifetime    time.Duration
	Debug          bool
}

// InitializeGORMDatabase нь холбогдсон *gorm.DB-г буцаана. Уг handle нь
// gorm otel tracing plugin-ээр дамжуулан OpenTelemetry-ээр хэмжигдсэн —
// Query/Exec бүр semantic-convention атрибутаар (db.system, db.statement)
// тэмдэглэгдсэн span гаргадаг. Энэ нь анхны sqlx суурьтай boilerplate-д
// ашиглагдаж байсан otelsqlx хэмжилтийг орлоно. Үндсэн *sql.DB pool нь
// config-оос тохируулагддаг; pool-ийн статистикийг *sql.DB-г
// pkg/observability-д бүртгэх замаар /metrics-ээр илчилдэг.
func (config *GORMConfig) InitializeGORMDatabase() (*gorm.DB, error) {
	logLevel := gormlogger.Warn
	if config.Debug {
		logLevel = gormlogger.Info
	}

	gormDB, err := gorm.Open(postgres.Open(config.DataSourceName), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
		// TranslateError нь драйверт хамаарах алдаануудыг (жишээ нь
		// Postgres 23505 unique_violation) GORM-ийн зөөвөрлөгдөх
		// sentinel-үүд рүү (gorm.ErrDuplicatedKey, gorm.ErrRecordNotFound)
		// буулгадаг тул repository давхарга нь драйверийн алдааны төрлийг
		// import хийлгүйгээр зөрчлийг илрүүлж чадна.
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// DB statement бүрт зориулсан OpenTelemetry tracing. Tracing
	// идэвхгүй үед global provider нь OTel-ийн no-op байх тул энэ нь бараг
	// ямар ч зардалгүй.
	if err := gormDB.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		return nil, fmt.Errorf("error registering gorm otel plugin: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("error accessing underlying sql.DB: %v", err)
	}

	// өгөгдлийн сан руу нээлттэй холболтын дээд тоог тогтооно
	logger.Info(fmt.Sprintf("Setting maximum number of open connections to %d", config.MaxOpenConns), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryDatabase})
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)

	// pool дахь сул (idle) холболтын дээд тоог тогтооно
	logger.Info(fmt.Sprintf("Setting maximum number of idle connections to %d", config.MaxIdleConns), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryDatabase})
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)

	// шинэ холболтыг хүлээх дээд хугацааг тогтооно
	logger.Info(fmt.Sprintf("Setting maximum lifetime for a connection to %s", config.MaxLifetime), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryDatabase})
	sqlDB.SetConnMaxLifetime(config.MaxLifetime)

	// холболтуудын сул (idle) байх дээд хугацааг тогтооно
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return gormDB, nil
}
