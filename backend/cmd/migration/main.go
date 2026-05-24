// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package main

import (
	"context"
	"flag"

	"templatev27/internal/config"
	"templatev27/internal/constants"
	"templatev27/internal/datasources/drivers"
	"templatev27/internal/datasources/migration"
	"templatev27/pkg/logger"
)

// migrationsDir нь модулийн root-оос харьцангуй (make mig-up нь backend/-ээс
// ажилладаг). SQL файлууд нь конвенцийн дагуу backend/migrations/-д байрлана.
const migrationsDir = "migrations"

var (
	up   bool
	down bool
)

func init() {
	if err := config.InitializeAppConfig(); err != nil {
		logger.Fatal(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryConfig})
	}
	logger.Info("configuration loaded", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryConfig})
}

func main() {
	flag.BoolVar(&up, "up", false, "apply new tables, columns, or other structures")
	flag.BoolVar(&down, "down", false, "drop tables, columns, or other structures")
	flag.Parse()

	db, err := drivers.SetupGORMPostgres()
	if err != nil {
		logger.Panic(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryMigration})
	}
	defer func() {
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	}()

	runner := migration.New(db, migrationsDir)
	ctx := context.Background()

	if up {
		// Эхэлд SQL файлууд (өргөтгөлүүд, partial-unique индексүүд,
		// uuid_generate_v4() id анхдагч утга), дараа нь моделоос гарган авсан
		// баганануудыг тааруулахаар GORM AutoMigrate. Хоёулаа идемпотент.
		if err := runner.Up(ctx); err != nil {
			logger.Fatal(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryMigration})
		}
		if err := runner.AutoMigrate(ctx); err != nil {
			logger.Fatal(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryMigration})
		}
	}
	if down {
		if err := runner.Down(ctx); err != nil {
			logger.Fatal(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryMigration})
		}
	}
}
