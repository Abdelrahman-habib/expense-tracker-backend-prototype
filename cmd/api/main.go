package main

import (
	"log"

	"github.com/Abdelrahman-habib/expense-tracker/config"
	"github.com/Abdelrahman-habib/expense-tracker/internal/app"
)

// @title           Expense Tracker API
// @version         1.0
// @description     REST API for expense tracking application with user management
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token authentication
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Create and start application
	app, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	// Start the application
	if err := app.Start(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
