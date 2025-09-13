package main

import (
	"goly/model"
	"goly/server"
)



// main is the entry point of the application.
// It sets up the database connection and starts the HTTP server.
func main() {
	model.Setup()
	server.SetupAndListen()
}