// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package seeders

import (
	"templatev27/internal/constants"
	"templatev27/internal/datasources/records"
	"templatev27/pkg/helpers"
	"templatev27/pkg/logger"
)

var pass string
var UserData []records.Users

func init() {
	var err error
	pass, err = helpers.GenerateHash("12345")
	if err != nil {
		logger.Panic(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	}

	UserData = []records.Users{
		{
			Username: "patrick star 7",
			Email:    "patrick@gmail.com",
			Password: pass,
			Active:   true,
			RoleId:   1,
		},
		{
			Username: "john doe",
			Email:    "johndoe@gmail.com",
			Password: pass,
			Active:   false,
			RoleId:   2,
		},
	}
}
