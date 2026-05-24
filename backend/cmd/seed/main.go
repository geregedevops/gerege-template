// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package main

import (
	"templatev27/cmd/seed/seeders"
	"templatev27/internal/config"
	"templatev27/internal/constants"
	"templatev27/internal/datasources/drivers"
	"templatev27/pkg/logger"
)

func init() {
	if err := config.InitializeAppConfig(); err != nil {
		logger.Fatal(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryConfig})
	}
	logger.Info("configuration loaded", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryConfig})
}

func main() {
	db, err := drivers.SetupGORMPostgres()
	if err != nil {
		logger.Panic(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	}
	defer func() {
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	}()

	logger.Info("seeding...", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})

	seeder := seeders.NewSeeder(db)
	err = seeder.UserSeeder(seeders.UserData)
	if err != nil {
		logger.Panic(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	}

	logger.Info("seeding success!", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
}
