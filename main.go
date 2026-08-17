package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"sequAcc/database"
	"sequAcc/routes"
)

// @title           Sequspace Achievements API
// @version         1.0
// @description     This is the microservice for Sequspace achievements, handling shining stars and other awards.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@sequspace.com

// @host      localhost:4009
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Ensure uploads directory exists
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	// Initialize database connection
	database.ConnectDB()

	// Setup Gin router
	r := gin.Default()

	// Serve uploaded files statically
	r.Static("/uploads", "./uploads")

	// Setup Routes
	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for the microservice
	}

	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
