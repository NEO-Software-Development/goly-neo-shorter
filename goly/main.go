package main

import (
	"goly-app/database"
	"goly-app/goly/server"
)

func main() {
	database.Setup()
	server.SetupAndListen()
}