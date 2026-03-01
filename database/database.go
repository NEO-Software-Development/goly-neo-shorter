package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Setup() {
	var err error
	DB, err = gorm.Open(sqlite.Open("goly.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
}
