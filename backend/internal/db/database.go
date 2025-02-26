package db

import (
	"database/sql"
	"fmt"
	"log"
)

// Database connection details
var db *sql.DB

func init() {
	var err error
	// db, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=password123 dbname=inspec sslmode=disable")
	db, err = sql.Open("postgres", "host=inspec-postgres port=5432 user=postgres password=password123 dbname=inspec sslmode=disable")

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	// Create table if it doesn't exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS inspec_profiles (
	    id SERIAL PRIMARY KEY,
	    name VARCHAR(255) NOT NULL,
	    url TEXT NOT NULL,
	    description TEXT,
	    stars INT DEFAULT 0,
	    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	fmt.Println("Table 'inspec_profiles' ensured to exist.")
}
