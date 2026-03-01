package main

import (
	"fmt"
	"goly-app/auth"
	"goly-app/database"
	"goly-app/goly/model"
	"goly-app/goly/server"
)

func main() {
	database.Setup()
	err := database.DB.AutoMigrate(&model.Goly{}, &auth.User{}, &auth.Session{})
	if err != nil {
		fmt.Println(err)
	}
	server.SetupAndListen()
}