package main

import (
	"net/http"

	"log"

	"github.com/neozhixuan/gt_assessment/config"
	"github.com/neozhixuan/gt_assessment/database"
	"github.com/neozhixuan/gt_assessment/routes"
)

func main() {
	// Load our .env variables
	config.LoadEnv()

	// Initialize database connection
	database.InitDB()

	// Close the DB connection when the app stops
	defer database.DB.Close()

	// Set up routes
	r := routes.SetupRouter()

	// Initialise the server
	log.Fatal(http.ListenAndServe(":8080", r))
}
