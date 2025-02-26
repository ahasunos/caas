package main

import (
	"fmt"
	"log"

	"github.com/ahasunos/inspec-cloud/backend/internal/api"
	"github.com/ahasunos/inspec-cloud/backend/internal/db"

	_ "github.com/ahasunos/inspec-cloud/backend/docs" // Import docs
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Initialize database
	err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Setup router
	r := api.SetupRouter()

	// Serve static files for Swagger JSON
	r.Static("/docs", "./docs")

	// Log registered routes
	for _, route := range r.Routes() {
		fmt.Printf("Registered Route: %s %s\n", route.Method, route.Path)
	}

	// Ensure Swagger UI fetches the correct file
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/docs/swagger.json")))

	// Run the server
	r.Run(":8080")
}
