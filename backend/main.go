package main

import (
	"log"

	"github.com/ahasunos/inspec-cloud/backend/internal/api"
	"github.com/ahasunos/inspec-cloud/backend/internal/db"
)

func main() {
	// Initialize database
	err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Setup router
	r := api.SetupRouter()

	// Run the server
	r.Run(":8080")
}
