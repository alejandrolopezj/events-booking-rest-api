package main

import (
	"os"

	"github.com/alejandrolopezj/events-booking-rest-api/db"
	"github.com/alejandrolopezj/events-booking-rest-api/routes"
	"github.com/alejandrolopezj/events-booking-rest-api/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	os.Exit(runMain(db.InitDB, func(server *gin.Engine) error {
		return server.Run(":8080")
	}))
}

func runMain(initializeDatabase func() error, serve func(*gin.Engine) error) int {
	utils.InitLogger()

	if err := run(initializeDatabase, serve); err != nil {
		utils.Logger.Error("Application stopped", "error", err)
		return 1
	}
	return 0
}

func run(initializeDatabase func() error, serve func(*gin.Engine) error) error {
	utils.Logger.Info("Starting the Event Booking REST API server...")
	if err := initializeDatabase(); err != nil {
		return err
	}
	server := gin.Default()

	utils.Logger.Info("Registering routes...")
	routes.RegisterRoutes(server)

	utils.Logger.Info("Starting the server on port 8080...")
	return serve(server)
}
