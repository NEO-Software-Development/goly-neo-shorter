package database

import (
	"fmt"
	"goly-app/auth"
	"goly-app/goly/model"

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

	err = DB.AutoMigrate(&model.Goly{}, &auth.User{}, &auth.Session{})
	if err != nil {
		fmt.Println(err)
	}
}
